// Command enrollment-webhook is a self-contained demo integrator for ShellHub's
// install-key "webhook" enrollment mode. ShellHub POSTs a signed enrollment
// payload here at device enrollment; this server verifies the HMAC signature,
// applies a live policy (accept/reject/pending/allowlist), and serves a small
// dark web UI with a live feed of received enrollments.
//
// Stdlib only, no third-party dependencies.
package main

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	_ "embed"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"strings"
	"sync"
	"time"
)

//go:embed index.html
var indexHTML string

// deviceInfo mirrors ShellHub's enrollment payload `info` object.
type deviceInfo struct {
	ID         string `json:"id"`
	PrettyName string `json:"pretty_name"`
	Version    string `json:"version"`
	Arch       string `json:"arch"`
	Platform   string `json:"platform"`
}

// webhookRequest is the enrollment payload ShellHub POSTs to the integrator.
// Field names/tags match api/services/enrollment.go:enrollmentWebhookRequest.
type webhookRequest struct {
	TenantID       string      `json:"tenant_id"`
	InstallKeyID   string      `json:"install_key_id"`   // SHA256 digest of the key (stable, non-secret) — key policy on this
	InstallKeyName string      `json:"install_key_name"` // human label, for logs/recognition only — never key policy on it
	DeviceUID      string      `json:"device_uid"`
	MAC            string      `json:"mac"`
	Hostname       string      `json:"hostname"`
	Info           *deviceInfo `json:"info,omitempty"`
	SourceIP       string      `json:"source_ip"`
	Timestamp      time.Time   `json:"timestamp"`
	// CallbackURL is a ready, token-authenticated URL the integrator can POST a
	// late decision to. Present when ShellHub knows its public base.
	CallbackURL string `json:"callback_url,omitempty"`
}

// webhookResponse is the integrator's answer. ShellHub accepts only
// accept/reject/pending and fails closed to pending on anything else.
type webhookResponse struct {
	Decision string `json:"decision"`
	Reason   string `json:"reason,omitempty"`
}

// event is one recorded enrollment for the live feed.
type event struct {
	ID       string         `json:"id"` // stable, targetable by /api/decide
	Time     time.Time      `json:"time"`
	Payload  webhookRequest `json:"payload"`
	Verified bool           `json:"verified"`
	Decision string         `json:"decision"` // the synchronous decision returned to ShellHub
	Reason   string         `json:"reason,omitempty"`
	// CallbackURL is stored when the sync decision was "defer", so /api/decide
	// can POST the late decision to it.
	CallbackURL string `json:"callback_url,omitempty"`
	// Async resolution, set once a deferred event is decided via the callback.
	FinalDecision   string `json:"final_decision,omitempty"`
	CallbackOutcome string `json:"callback_outcome,omitempty"` // e.g. "callback → 200" or an error
}

const eventCap = 50

type server struct {
	secret string
	addr   string

	mu        sync.Mutex
	policy    string   // accept | reject | pending | allowlist | defer
	allowlist []string // MACs (lowercased) used by the allowlist policy
	events    []event  // ring buffer, oldest first
	nextID    int      // monotonic event id source
}

func main() {
	addr := flag.String("addr", ":9090", "listen address")
	secret := flag.String("secret", "dev-secret", "webhook HMAC secret (must match the ShellHub install key)")
	flag.Parse()

	s := &server{
		secret: *secret,
		addr:   *addr,
		policy: "accept",
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/webhook", s.handleWebhook)
	mux.HandleFunc("/api/state", s.handleState)
	mux.HandleFunc("/api/policy", s.handlePolicy)
	mux.HandleFunc("/api/simulate", s.handleSimulate)
	mux.HandleFunc("/api/decide", s.handleDecide)
	mux.HandleFunc("/api/callback-sink/", s.handleCallbackSink)
	mux.HandleFunc("/", s.handleIndex)

	log.Printf("enrollment-webhook demo listening on %s", *addr)
	log.Printf("webhook endpoint: POST http://<host>%s/webhook", *addr)
	log.Printf("shared secret: %s", *secret)
	log.Fatal(http.ListenAndServe(*addr, mux))
}

// sign returns the hex HMAC-SHA256 of body under the shared secret, matching
// what ShellHub sends in X-ShellHub-Signature.
func sign(secret string, body []byte) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(body)

	return hex.EncodeToString(mac.Sum(nil))
}

func (s *server) handleWebhook(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)

		return
	}

	body, err := readBody(r)
	if err != nil {
		http.Error(w, "cannot read body", http.StatusBadRequest)

		return
	}

	verified := hmac.Equal([]byte(sign(s.secret, body)), []byte(r.Header.Get("X-ShellHub-Signature")))

	var payload webhookRequest
	_ = json.Unmarshal(body, &payload) // a malformed body still gets recorded below

	var resp webhookResponse
	if !verified {
		// Record the attempt but refuse to act on an unsigned/forged request.
		// Fail closed to pending so the demo surfaces the reason instead of erroring.
		resp = webhookResponse{Decision: "pending", Reason: "signature mismatch"}
	} else {
		resp = s.decide(payload)
	}

	e := event{
		Time:     time.Now(),
		Payload:  payload,
		Verified: verified,
		Decision: resp.Decision,
		Reason:   resp.Reason,
	}
	// A deferred decision keeps the callback URL so /api/decide can answer later.
	if resp.Decision == "defer" {
		e.CallbackURL = payload.CallbackURL
	}
	s.record(e)

	writeJSON(w, http.StatusOK, resp)
}

// decide applies the current policy to a payload. Called with a verified payload.
func (s *server) decide(p webhookRequest) webhookResponse {
	s.mu.Lock()
	policy := s.policy
	allowlist := append([]string(nil), s.allowlist...)
	s.mu.Unlock()

	switch policy {
	case "accept":
		return webhookResponse{Decision: "accept", Reason: "policy: accept all"}
	case "reject":
		return webhookResponse{Decision: "reject", Reason: "policy: reject all"}
	case "pending":
		return webhookResponse{Decision: "pending", Reason: "policy: manual review"}
	case "defer":
		return webhookResponse{Decision: "defer", Reason: "policy: decide later via callback"}
	case "allowlist":
		mac := strings.ToLower(strings.TrimSpace(p.MAC))
		for _, allowed := range allowlist {
			if allowed == mac {
				return webhookResponse{Decision: "accept", Reason: "MAC in allowlist"}
			}
		}

		return webhookResponse{Decision: "reject", Reason: "MAC not in allowlist"}
	default:
		return webhookResponse{Decision: "pending", Reason: "unknown policy"}
	}
}

func (s *server) record(e event) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.nextID++
	e.ID = fmt.Sprintf("ev-%d", s.nextID)

	s.events = append(s.events, e)
	if len(s.events) > eventCap {
		s.events = s.events[len(s.events)-eventCap:]
	}
}

func (s *server) handleState(w http.ResponseWriter, _ *http.Request) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Newest first for the feed.
	ordered := make([]event, len(s.events))
	for i, e := range s.events {
		ordered[len(s.events)-1-i] = e
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"policy":       s.policy,
		"allowlist":    s.allowlist,
		"secret":       s.secret,
		"webhook_path": "/webhook",
		"events":       ordered,
	})
}

func (s *server) handlePolicy(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)

		return
	}

	var in struct {
		Policy    string   `json:"policy"`
		Allowlist []string `json:"allowlist"`
	}
	body, _ := readBody(r)
	if err := json.Unmarshal(body, &in); err != nil {
		http.Error(w, "bad json", http.StatusBadRequest)

		return
	}

	switch in.Policy {
	case "accept", "reject", "pending", "allowlist", "defer":
	default:
		http.Error(w, "invalid policy", http.StatusBadRequest)

		return
	}

	normalized := make([]string, 0, len(in.Allowlist))
	for _, m := range in.Allowlist {
		if m = strings.ToLower(strings.TrimSpace(m)); m != "" {
			normalized = append(normalized, m)
		}
	}

	s.mu.Lock()
	s.policy = in.Policy
	s.allowlist = normalized
	s.mu.Unlock()

	writeJSON(w, http.StatusOK, map[string]any{"policy": in.Policy, "allowlist": normalized})
}

// handleSimulate crafts a realistic fake enrollment, signs it, and POSTs it to
// this server's own /webhook so the full verification path is exercised.
func (s *server) handleSimulate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)

		return
	}

	payload := fakeEnrollment(s.addr)
	body, _ := json.Marshal(payload)

	// Loop back over the network so signature verification runs for real.
	target := "http://" + localHost(s.addr) + "/webhook"
	req, err := http.NewRequest(http.MethodPost, target, bytes.NewReader(body))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)

		return
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-ShellHub-Signature", sign(s.secret, body))

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadGateway)

		return
	}
	defer resp.Body.Close() //nolint:errcheck

	out, _ := readAll(resp.Body)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(resp.StatusCode)
	_, _ = w.Write(out)
}

// handleDecide resolves a previously deferred enrollment: it looks up the stored
// callback_url for the event and POSTs the late decision to it (server-side, so
// the browser never hits ShellHub directly). The token in the URL is the only
// credential, so no signature or API key is attached. The callback's outcome is
// recorded on the event and returned.
func (s *server) handleDecide(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)

		return
	}

	var in struct {
		ID       string `json:"id"`
		Decision string `json:"decision"`
		Reason   string `json:"reason"`
	}
	body, _ := readBody(r)
	if err := json.Unmarshal(body, &in); err != nil {
		http.Error(w, "bad json", http.StatusBadRequest)

		return
	}
	if in.Decision != "accept" && in.Decision != "reject" {
		http.Error(w, "decision must be accept or reject", http.StatusBadRequest)

		return
	}

	s.mu.Lock()
	idx := -1
	for i := range s.events {
		if s.events[i].ID == in.ID {
			idx = i

			break
		}
	}
	if idx < 0 {
		s.mu.Unlock()
		http.Error(w, "unknown event", http.StatusNotFound)

		return
	}
	callbackURL := s.events[idx].CallbackURL
	alreadyResolved := s.events[idx].FinalDecision != ""
	s.mu.Unlock()

	if callbackURL == "" {
		http.Error(w, "event has no callback url", http.StatusConflict)

		return
	}
	if alreadyResolved {
		http.Error(w, "event already resolved", http.StatusConflict)

		return
	}

	reason := in.Reason
	if reason == "" {
		reason = "resolved by integrator"
	}

	outcome, ok := s.postCallback(callbackURL, webhookResponse{Decision: in.Decision, Reason: reason})

	s.mu.Lock()
	if idx < len(s.events) && s.events[idx].ID == in.ID { // ring buffer may have shifted
		s.events[idx].FinalDecision = in.Decision
		s.events[idx].CallbackOutcome = outcome
	}
	s.mu.Unlock()

	status := http.StatusOK
	if !ok {
		status = http.StatusBadGateway
	}
	writeJSON(w, status, map[string]any{"id": in.ID, "final_decision": in.Decision, "outcome": outcome, "ok": ok})
}

// postCallback POSTs a decision to a token-authenticated callback URL and returns
// a human-readable outcome plus whether it landed on a 2xx.
func (s *server) postCallback(url string, resp webhookResponse) (string, bool) {
	body, _ := json.Marshal(resp)

	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return "invalid callback url: " + err.Error(), false
	}
	req.Header.Set("Content-Type", "application/json")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return "callback failed: " + err.Error(), false
	}
	defer res.Body.Close() //nolint:errcheck

	ok := res.StatusCode >= http.StatusOK && res.StatusCode < http.StatusMultipleChoices

	return fmt.Sprintf("callback → %d", res.StatusCode), ok
}

// handleCallbackSink is a local stand-in for ShellHub's real callback receiver so
// the deferred path can be demoed end-to-end without a running ShellHub. Simulated
// enrollments point their callback_url here; it just acknowledges with 200.
func (s *server) handleCallbackSink(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)

		return
	}

	body, _ := readBody(r)
	log.Printf("callback-sink received %s: %s", r.URL.Path, string(body))
	writeJSON(w, http.StatusOK, map[string]any{"ok": true})
}

func (s *server) handleIndex(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)

		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	// No-store so a rebuilt binary's HTML is never masked by a stale browser cache.
	w.Header().Set("Cache-Control", "no-store")
	_, _ = w.Write([]byte(indexHTML))
}

// --- fake enrollment data for the simulate button ---

var (
	fakeHostnames = []string{"web-01", "db-primary", "edge-eu-3", "ci-runner-7", "gpu-node-2", "kiosk-lobby", "vpn-gw", "cache-05"}
	fakeKeyNames  = []string{"ci-runners", "edge-gateways", "production-fleet", "lab-ephemeral", "staging-once"}
	fakeOS        = []deviceInfo{
		{ID: "debian", PrettyName: "Debian GNU/Linux 12", Version: "v0.18.0", Arch: "amd64", Platform: "docker"},
		{ID: "ubuntu", PrettyName: "Ubuntu 24.04 LTS", Version: "v0.18.0", Arch: "arm64", Platform: "docker"},
		{ID: "alpine", PrettyName: "Alpine Linux v3.20", Version: "v0.18.0", Arch: "amd64", Platform: "docker"},
		{ID: "raspbian", PrettyName: "Raspbian GNU/Linux 12", Version: "v0.17.4", Arch: "arm", Platform: "native"},
		{ID: "fedora", PrettyName: "Fedora Linux 40", Version: "v0.18.0", Arch: "amd64", Platform: "native"},
	}
)

func fakeEnrollment(addr string) webhookRequest {
	info := fakeOS[rand.Intn(len(fakeOS))]

	// Point the callback at this server's own sink so a simulated deferred
	// enrollment can still be resolved end-to-end (real ShellHub sends its own
	// token-authenticated URL here instead).
	callbackURL := "http://" + localHost(addr) + "/api/callback-sink/" + randHex(24)

	return webhookRequest{
		TenantID:       "3dd0d1f8-9b2a-4c7e-8f1a-0a1b2c3d4e5f",
		InstallKeyID:   randHex(64), // SHA256 digest, 64 hex chars
		InstallKeyName: fakeKeyNames[rand.Intn(len(fakeKeyNames))],
		DeviceUID:      randHex(32),
		MAC:            randMAC(),
		Hostname:       fakeHostnames[rand.Intn(len(fakeHostnames))],
		Info:           &info,
		SourceIP:       fmt.Sprintf("203.0.113.%d", 1+rand.Intn(254)),
		Timestamp:      time.Now().UTC(),
		CallbackURL:    callbackURL,
	}
}

func randHex(n int) string {
	const hexd = "0123456789abcdef"
	b := make([]byte, n)
	for i := range b {
		b[i] = hexd[rand.Intn(16)]
	}

	return string(b)
}

func randMAC() string {
	b := make([]byte, 6)
	for i := range b {
		b[i] = byte(rand.Intn(256))
	}

	return fmt.Sprintf("%02x:%02x:%02x:%02x:%02x:%02x", b[0], b[1], b[2], b[3], b[4], b[5])
}

// --- small helpers ---

// localHost turns a listen address (which may be host-less like ":9090" or
// "0.0.0.0:9090") into a dialable host:port for the self-loopback simulate call.
func localHost(addr string) string {
	if strings.HasPrefix(addr, ":") {
		return "127.0.0.1" + addr
	}
	if strings.HasPrefix(addr, "0.0.0.0:") {
		return "127.0.0.1:" + strings.TrimPrefix(addr, "0.0.0.0:")
	}

	return addr
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func readBody(r *http.Request) ([]byte, error) {
	defer r.Body.Close() //nolint:errcheck

	return readAll(r.Body)
}

func readAll(r interface{ Read([]byte) (int, error) }) ([]byte, error) {
	var buf bytes.Buffer
	tmp := make([]byte, 4096)
	for {
		n, err := r.Read(tmp)
		if n > 0 {
			buf.Write(tmp[:n])
		}
		if err != nil {
			if err.Error() == "EOF" {
				return buf.Bytes(), nil
			}

			return buf.Bytes(), nil
		}
	}
}
