package main

import (
	"bytes"
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"errors"
	"flag"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"regexp"
	"slices"
	"strings"
	"time"

	"golang.org/x/crypto/ssh"
	"golang.org/x/net/http2"
	"golang.org/x/net/websocket"
)

type config struct {
	httpBase   string
	sshAddr    string
	user       string
	password   string
	tenant     string
	deviceUser string
	devicePass string
	keyFile    string
	device     string
	namespace  string
	hostname   string
	mode       string
	iterations int
	label      string
	output     string
	timeout    time.Duration
	connect    string
	deviceName string
	profile    string
	websocket  string
}

type phase struct {
	Name string  `json:"name"`
	MS   float64 `json:"ms"`
}

type sample struct {
	Label   string  `json:"label"`
	Mode    string  `json:"mode"`
	Device  string  `json:"device"`
	Tenant  string  `json:"tenant"`
	Profile string  `json:"profile"`
	Iter    int     `json:"iter"`
	At      string  `json:"at"`
	Total   float64 `json:"total_ms"`
	Phases  []phase `json:"phases"`
	Error   string  `json:"error,omitempty"`
}

type stopwatch struct {
	start  time.Time
	phases []phase
}

func newStopwatch() *stopwatch {
	return &stopwatch{start: time.Now()} //nolint:forbidigo // the start of an elapsed-time measurement
}

func (s *stopwatch) mark(name string) {
	s.phases = append(s.phases, phase{Name: name, MS: ms(time.Since(s.start))})
}

func ms(d time.Duration) float64 { return float64(d.Microseconds()) / 1000 }

var prompt = regexp.MustCompile(`[$#] (\x1b\[[0-9;?]*[a-zA-Z])*$`)

const askpassEnv = "CONNBENCH_ASKPASS" //nolint:gosec // the name of the variable ssh's askpass reads the password from

func main() {
	if password, ok := os.LookupEnv(askpassEnv); ok {
		_, _ = fmt.Fprintln(os.Stdout, password)

		return
	}

	cfg := config{}
	flag.StringVar(&cfg.httpBase, "http", "http://localhost:80", "ShellHub HTTP base URL")
	flag.StringVar(&cfg.sshAddr, "ssh", "localhost:2222", "ShellHub SSH address")
	flag.StringVar(&cfg.user, "user", "dev", "ShellHub account")
	flag.StringVar(&cfg.password, "password", "Secret123!-", "ShellHub account password")
	flag.StringVar(&cfg.tenant, "tenant", "00000000-0000-4000-0000-000000000000", "namespace tenant")
	flag.StringVar(&cfg.deviceUser, "device-user", "shellhub", "OS user on the device")
	flag.StringVar(&cfg.devicePass, "device-password", "shellhub", "OS password on the device")
	flag.StringVar(&cfg.keyFile, "key", "/tmp/connbench_ed25519", "identity key file, created by setup")
	flag.StringVar(&cfg.mode, "mode", "cli-password", "cli-password|cli-key|web-password|web-key|openssh-password|openssh-key")
	flag.IntVar(&cfg.iterations, "n", 10, "iterations")
	flag.StringVar(&cfg.label, "label", "baseline", "label stored with each sample")
	flag.StringVar(&cfg.output, "o", "", "append JSON lines here")
	flag.DurationVar(&cfg.timeout, "timeout", 30*time.Second, "per-attempt timeout")
	flag.StringVar(&cfg.deviceName, "device", "", "device name; the first accepted device when empty")
	flag.StringVar(&cfg.profile, "profile", "lan", "network profile stored with each sample")
	flag.StringVar(&cfg.connect, "connect", "", "dial this address for every HTTP request, keeping the URL's Host")
	flag.StringVar(&cfg.websocket, "websocket", "h1", "h1 opens the web terminal on its own connection, h2 over the API's HTTP/2 connection")
	flag.Parse()

	base, ok := http.DefaultTransport.(*http.Transport)
	if !ok {
		fmt.Fprintln(os.Stderr, "the default HTTP transport is not an *http.Transport")
		os.Exit(1)
	}

	transport := base.Clone()
	transport.TLSClientConfig = benchTLS
	transport.ForceAttemptHTTP2 = true

	if cfg.connect != "" {
		transport.DialContext = func(ctx context.Context, network, _ string) (net.Conn, error) {
			return (&net.Dialer{}).DialContext(ctx, network, cfg.connect)
		}
	}

	http.DefaultClient.Transport = transport

	if cfg.websocket == "h2" {
		http.DefaultClient.Transport = extendedConnectTransport(cfg.connect)
	}

	if flag.Arg(0) == "setup" {
		if err := setup(cfg); err != nil {
			fmt.Fprintln(os.Stderr, "setup:", err)
			os.Exit(1)
		}

		return
	}

	if err := run(cfg); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

type api struct {
	base  string
	token string
}

func (a *api) do(method, path string, body, out any) error {
	var reader io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return err
		}

		reader = bytes.NewReader(data)
	}

	req, err := http.NewRequestWithContext(context.Background(), method, a.base+path, reader)
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	if a.token != "" {
		req.Header.Set("Authorization", "Bearer "+a.token)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()

	data, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 300 {
		return fmt.Errorf("%s %s: %d %s", method, path, resp.StatusCode, data)
	}

	if out != nil {
		return json.Unmarshal(data, out)
	}

	return nil
}

func login(cfg config) (*api, error) {
	a := &api{base: cfg.httpBase}

	var auth struct {
		Token string `json:"token"`
	}
	if err := a.do(http.MethodPost, "/api/login", map[string]string{"username": cfg.user, "password": cfg.password}, &auth); err != nil {
		return nil, err
	}

	a.token = auth.Token

	if cfg.tenant != "" {
		var switched struct {
			Token string `json:"token"`
		}
		if err := a.do(http.MethodGet, "/api/auth/token/"+cfg.tenant, nil, &switched); err == nil && switched.Token != "" {
			a.token = switched.Token
		}
	}

	return a, nil
}

type device struct {
	UID       string `json:"uid"`
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
}

func findDevice(a *api, name string) (device, error) {
	var devices []device
	if err := a.do(http.MethodGet, "/api/devices?status=accepted&per_page=100", nil, &devices); err != nil {
		return device{}, err
	}

	if name != "" {
		devices = slices.DeleteFunc(devices, func(d device) bool { return d.Name != name })
	}

	if len(devices) == 0 {
		return device{}, errors.New("no accepted device")
	}

	return devices[0], nil
}

func loadKey(path string) (ed25519.PrivateKey, ssh.Signer, error) {
	data, err := os.ReadFile(path) //nolint:gosec // the key path is the operator's own flag
	if err != nil {
		return nil, nil, err
	}

	raw, err := ssh.ParseRawPrivateKey(data)
	if err != nil {
		return nil, nil, err
	}

	priv, ok := raw.(*ed25519.PrivateKey)
	if !ok {
		return nil, nil, errors.New("key is not ed25519")
	}

	signer, err := ssh.NewSignerFromKey(*priv)
	if err != nil {
		return nil, nil, err
	}

	return *priv, signer, nil
}

func setup(cfg config) error {
	a, err := login(cfg)
	if err != nil {
		return err
	}

	if _, err := os.Stat(cfg.keyFile); errors.Is(err, os.ErrNotExist) {
		_, priv, err := ed25519.GenerateKey(rand.Reader)
		if err != nil {
			return err
		}

		block, err := ssh.MarshalPrivateKey(priv, "")
		if err != nil {
			return err
		}

		if err := os.WriteFile(cfg.keyFile, pem.EncodeToMemory(block), 0o600); err != nil {
			return err
		}
	}

	_, signer, err := loadKey(cfg.keyFile)
	if err != nil {
		return err
	}

	line := strings.TrimSpace(string(ssh.MarshalAuthorizedKey(signer.PublicKey())))

	if err := a.do(http.MethodPost, "/api/ssh-identities", map[string]string{"name": "connbench", "data": line}, nil); err != nil {
		fmt.Fprintln(os.Stderr, "identity:", err)
	}

	policy := map[string]any{
		"name":    "connbench",
		"subject": map[string]string{"type": "role", "value": "owner"},
		"logins":  []string{"*"},
		"filter":  map[string]any{"all": true},
		"action":  "allow",
	}
	if err := a.do(http.MethodPost, "/api/access-policies", policy, nil); err != nil {
		fmt.Fprintln(os.Stderr, "policy:", err)
	}

	encoded := base64.StdEncoding.EncodeToString([]byte(line))
	if err := a.do(http.MethodPost, "/api/sshkeys/public-keys", map[string]any{
		"name": "connbench", "data": encoded, "username": ".*", "filter": map[string]string{"hostname": ".*"},
	}, nil); err != nil {
		fmt.Fprintln(os.Stderr, "public key:", err)
	}

	_, _ = fmt.Fprintln(os.Stdout, "ok", ssh.FingerprintSHA256(signer.PublicKey()))

	return nil
}

func run(cfg config) error {
	a, err := login(cfg)
	if err != nil {
		return err
	}

	d, err := findDevice(a, cfg.deviceName)
	if err != nil {
		return err
	}

	cfg.device, cfg.namespace, cfg.hostname = d.UID, d.Namespace, d.Name

	out := io.Discard
	if cfg.output != "" {
		f, err := os.OpenFile(cfg.output, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o600)
		if err != nil {
			return err
		}
		defer func() { _ = f.Close() }()

		out = f
	}

	var totals []float64
	failures := 0

	for i := range cfg.iterations {
		var sw *stopwatch
		var err error

		switch cfg.mode {
		case "cli-password", "cli-key":
			sw, err = cliAttempt(cfg)
		case "openssh-password", "openssh-key":
			sw, err = opensshAttempt(cfg)
		case "web-password", "web-key":
			sw, err = webAttempt(cfg, a)
		default:
			return fmt.Errorf("unknown mode %q", cfg.mode)
		}

		s := sample{Label: cfg.label, Mode: cfg.mode, Device: cfg.hostname, Tenant: cfg.tenant, Profile: cfg.profile, Iter: i, At: time.Now().UTC().Format(time.RFC3339Nano), Phases: sw.phases} //nolint:forbidigo // a log timestamp
		if err != nil {
			s.Error = err.Error()
			failures++
		} else {
			s.Total = sw.phases[len(sw.phases)-1].MS
			totals = append(totals, s.Total)
		}

		line, err := json.Marshal(s)
		if err != nil {
			return err
		}

		if _, err := fmt.Fprintln(out, string(line)); err != nil {
			return err
		}
		fmt.Fprintf(os.Stderr, "%s #%d %s\n", cfg.mode, i, describe(s))

		time.Sleep(200 * time.Millisecond)
	}

	if len(totals) > 0 {
		slices.Sort(totals)
		_, _ = fmt.Fprintf(os.Stdout, "%s %s %s %s n=%d fail=%d min=%.1f p50=%.1f p90=%.1f max=%.1f\n", cfg.label, cfg.profile, cfg.hostname, cfg.mode, len(totals), failures,
			totals[0], percentile(totals, 0.5), percentile(totals, 0.9), totals[len(totals)-1])
	}

	if failures > 0 {
		return fmt.Errorf("%d failures", failures)
	}

	return nil
}

func percentile(sorted []float64, p float64) float64 {
	idx := int(float64(len(sorted)-1) * p)

	return sorted[idx]
}

func describe(s sample) string {
	parts := make([]string, 0, len(s.Phases))
	prev := 0.0

	for _, p := range s.Phases {
		parts = append(parts, fmt.Sprintf("%s=+%.1f", p.Name, p.MS-prev))
		prev = p.MS
	}

	if s.Error != "" {
		parts = append(parts, "error="+s.Error)
	}

	return fmt.Sprintf("total=%.1f %s", s.Total, strings.Join(parts, " "))
}

func awaitPrompt(sw *stopwatch, r io.Reader, deadline time.Time) error {
	buf := make([]byte, 0, 4096)
	chunk := make([]byte, 4096)
	first := true

	for time.Now().Before(deadline) { //nolint:forbidigo // a deadline
		n, err := r.Read(chunk)
		if n > 0 {
			if first {
				sw.mark("first-byte")
				first = false
			}

			buf = append(buf, chunk[:n]...)
			if prompt.Match(bytes.TrimRight(buf, "\x00")) {
				sw.mark("prompt")

				return nil
			}
		}

		if err != nil {
			return fmt.Errorf("before prompt: %w (got %q)", err, tail(buf))
		}
	}

	return fmt.Errorf("timeout waiting for prompt (got %q)", tail(buf))
}

func tail(b []byte) []byte {
	if len(b) > 200 {
		return b[len(b)-200:]
	}

	return b
}

func cliAttempt(cfg config) (*stopwatch, error) {
	sw := newStopwatch()
	deadline := time.Now().Add(cfg.timeout) //nolint:forbidigo // a deadline

	dialer := net.Dialer{Timeout: cfg.timeout}

	conn, err := dialer.DialContext(context.Background(), "tcp", cfg.sshAddr)
	if err != nil {
		return sw, err
	}
	defer func() { _ = conn.Close() }()

	_ = conn.SetDeadline(deadline)
	sw.mark("tcp")

	sshid := fmt.Sprintf("%s@%s.%s", cfg.deviceUser, cfg.namespace, cfg.hostname)

	var methods []ssh.AuthMethod

	if cfg.mode == "cli-key" {
		_, signer, err := loadKey(cfg.keyFile)
		if err != nil {
			return sw, err
		}

		methods = []ssh.AuthMethod{ssh.PublicKeysCallback(func() ([]ssh.Signer, error) {
			sw.mark("kex")

			return []ssh.Signer{signer}, nil
		})}
	} else {
		methods = []ssh.AuthMethod{ssh.PasswordCallback(func() (string, error) {
			sw.mark("kex")

			return cfg.devicePass, nil
		})}
	}

	cc, chans, reqs, err := ssh.NewClientConn(conn, cfg.sshAddr, &ssh.ClientConfig{
		User:            sshid,
		Auth:            methods,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(), //nolint:gosec // a benchmark against a local stack; nosemgrep
		BannerCallback:  func(string) error { return nil },
	})
	if err != nil {
		return sw, err
	}

	sw.mark("auth")

	client := ssh.NewClient(cc, chans, reqs)
	defer func() { _ = client.Close() }()

	session, err := client.NewSession()
	if err != nil {
		return sw, err
	}
	defer func() { _ = session.Close() }()

	sw.mark("channel")

	stdout, err := session.StdoutPipe()
	if err != nil {
		return sw, err
	}

	if err := session.RequestPty("xterm", 40, 120, ssh.TerminalModes{ssh.ECHO: 1}); err != nil {
		return sw, err
	}

	sw.mark("pty")

	if err := session.Shell(); err != nil {
		return sw, err
	}

	sw.mark("shell")

	return sw, awaitPrompt(sw, stdout, deadline)
}

type wsReply struct {
	Kind int    `json:"kind"`
	Data string `json:"data"`
}

type wsMessage struct {
	Kind int             `json:"kind"`
	Data json.RawMessage `json:"data"`
}

//nolint:staticcheck // net/http.Transport refuses the :protocol pseudo-header an RFC 8441 extended CONNECT needs
func extendedConnectTransport(connect string) *http2.Transport {
	return &http2.Transport{
		TLSClientConfig: benchTLS,
		DialTLSContext: func(ctx context.Context, network, addr string, config *tls.Config) (net.Conn, error) {
			if connect != "" {
				addr = connect
			}

			return (&tls.Dialer{Config: config}).DialContext(ctx, network, addr)
		},
	}
}

var benchTLS = &tls.Config{
	MinVersion:         tls.VersionTLS12,
	InsecureSkipVerify: true, //nolint:gosec // the bench front serves a throwaway certificate
	ClientSessionCache: tls.NewLRUClientSessionCache(8),
}

type xnetChannel struct {
	ws *websocket.Conn
}

func (c xnetChannel) receive() ([]byte, bool, error) {
	var frame []byte
	if err := websocket.Message.Receive(c.ws, &frame); err != nil {
		return nil, false, err
	}

	return frame, c.ws.PayloadType == websocket.BinaryFrame, nil
}

func (c xnetChannel) sendText(data []byte) error {
	return websocket.Message.Send(c.ws, string(data))
}

func (c xnetChannel) Close() error {
	return c.ws.Close()
}

type wsReader struct {
	ws      wsChannel
	priv    ed25519.PrivateKey
	sw      *stopwatch
	pending []byte
}

func (r *wsReader) Read(p []byte) (int, error) {
	for len(r.pending) == 0 {
		frame, binaryFrame, err := r.ws.receive()
		if err != nil {
			return 0, err
		}

		if binaryFrame {
			r.pending = frame

			break
		}

		var msg wsMessage
		if err := json.Unmarshal(frame, &msg); err != nil {
			r.pending = frame

			break
		}

		switch msg.Kind {
		case 3:
			var challenge string
			_ = json.Unmarshal(msg.Data, &challenge)

			data, err := base64.StdEncoding.DecodeString(challenge)
			if err != nil {
				return 0, err
			}

			r.sw.mark("sign")

			sig := base64.StdEncoding.EncodeToString(ed25519.Sign(r.priv, data))
			reply, err := json.Marshal(wsReply{Kind: 3, Data: sig})
			if err != nil {
				return 0, err
			}

			if err := r.ws.sendText(reply); err != nil {
				return 0, err
			}
		case 4:
			return 0, fmt.Errorf("server error: %s", msg.Data)
		case 5:
			r.sw.mark("session-uid")
		}
	}

	n := copy(p, r.pending)
	r.pending = r.pending[n:]

	return n, nil
}

func webAttempt(cfg config, a *api) (*stopwatch, error) {
	sw := newStopwatch()
	deadline := time.Now().Add(cfg.timeout) //nolint:forbidigo // a deadline

	body := map[string]string{"device": cfg.device, "username": cfg.deviceUser}

	var priv ed25519.PrivateKey

	if cfg.mode == "web-key" {
		k, signer, err := loadKey(cfg.keyFile)
		if err != nil {
			return sw, err
		}

		priv = k
		body["fingerprint"] = ssh.FingerprintSHA256(signer.PublicKey())
		body["public_key"] = strings.TrimSpace(string(ssh.MarshalAuthorizedKey(signer.PublicKey())))
	} else {
		body["password"] = cfg.devicePass
	}

	var session struct {
		Token string `json:"token"`
	}
	if err := a.do(http.MethodPost, "/ws/ssh/session", body, &session); err != nil {
		return sw, err
	}

	sw.mark("token")

	ws, err := openTerminal(cfg, session.Token, deadline, sw)
	if err != nil {
		return sw, err
	}
	defer func() { _ = ws.Close() }()

	sw.mark("websocket")

	return sw, awaitPrompt(sw, &wsReader{ws: ws, priv: priv, sw: sw}, deadline)
}

func openTerminal(cfg config, token string, deadline time.Time, sw *stopwatch) (wsChannel, error) { //nolint:ireturn // the terminal's transport depends on the flag
	base, err := url.Parse(cfg.httpBase)
	if err != nil {
		return nil, err
	}

	query := "/ws/ssh?token=" + token + "&cols=120&rows=40"

	if cfg.websocket == "h2" {
		return dialExtendedConnect(deadline, base.Scheme+"://"+base.Host+query, cfg.httpBase)
	}

	scheme := "ws"
	if base.Scheme == "https" {
		scheme = "wss"
	}

	wcfg, err := websocket.NewConfig(scheme+"://"+base.Host+query, cfg.httpBase)
	if err != nil {
		return nil, err
	}

	addr := cfg.connect
	if addr == "" {
		addr = base.Host
		if base.Port() == "" {
			addr += map[string]string{"http": ":80", "https": ":443"}[base.Scheme]
		}
	}

	dialer := net.Dialer{Timeout: cfg.timeout}

	raw, err := dialer.DialContext(context.Background(), "tcp", addr)
	if err != nil {
		return nil, err
	}

	_ = raw.SetDeadline(deadline)

	if base.Scheme == "https" {
		sw.mark("tcp")

		tlsConfig := benchTLS.Clone()
		tlsConfig.ServerName = base.Hostname()
		tlsConfig.NextProtos = []string{"http/1.1"}

		ctx, cancel := context.WithDeadline(context.Background(), deadline)
		defer cancel()

		secured := tls.Client(raw, tlsConfig)
		if err := secured.HandshakeContext(ctx); err != nil {
			_ = raw.Close()

			return nil, err
		}

		sw.mark("tls")

		raw = secured
	}

	ws, err := websocket.NewClient(wcfg, raw)
	if err != nil {
		_ = raw.Close()

		return nil, err
	}

	return xnetChannel{ws: ws}, nil
}

func opensshAttempt(cfg config) (*stopwatch, error) {
	sw := newStopwatch()

	host, port, err := net.SplitHostPort(cfg.sshAddr)
	if err != nil {
		return sw, err
	}

	askpass, err := os.Executable()
	if err != nil {
		return sw, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), cfg.timeout)
	defer cancel()

	args := []string{
		"-tt", "-p", port,
		"-o", "StrictHostKeyChecking=no", "-o", "UserKnownHostsFile=/dev/null", "-o", "LogLevel=ERROR",
		"-o", "ControlPath=none", "-o", "ConnectTimeout=30",
	}

	if cfg.mode == "openssh-key" {
		args = append(args, "-i", cfg.keyFile, "-o", "IdentitiesOnly=yes", "-o", "PreferredAuthentications=publickey")
	} else {
		args = append(args, "-o", "PubkeyAuthentication=no", "-o", "PreferredAuthentications=password,keyboard-interactive")
	}

	args = append(args, fmt.Sprintf("%s@%s.%s@%s", cfg.deviceUser, cfg.namespace, cfg.hostname, host))

	cmd := exec.CommandContext(ctx, "ssh", args...) //nolint:gosec // the arguments are the benchmark's own flags
	cmd.Env = append(os.Environ(), "SSH_ASKPASS="+askpass, "SSH_ASKPASS_REQUIRE=force", "DISPLAY=none", askpassEnv+"="+cfg.devicePass)

	stdin, err := cmd.StdinPipe()
	if err != nil {
		return sw, err
	}
	defer func() { _ = stdin.Close() }()

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return sw, err
	}

	if err := cmd.Start(); err != nil {
		return sw, err
	}

	defer func() {
		_ = cmd.Process.Kill()
		_ = cmd.Wait()
	}()

	sw.mark("spawn")

	return sw, awaitPrompt(sw, stdout, time.Now().Add(cfg.timeout)) //nolint:forbidigo // a deadline
}
