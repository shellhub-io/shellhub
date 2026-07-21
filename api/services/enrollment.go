package services

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/netip"
	"strings"
	"time"

	"code.dny.dev/ssrf"
	"github.com/shellhub-io/shellhub/api/store"
	"github.com/shellhub-io/shellhub/pkg/api/jwttoken"
	"github.com/shellhub-io/shellhub/pkg/api/requests"
	"github.com/shellhub-io/shellhub/pkg/clock"
	"github.com/shellhub-io/shellhub/pkg/envs"
	"github.com/shellhub-io/shellhub/pkg/models"
	log "github.com/sirupsen/logrus"
)

// enrollmentWebhookAllowedCIDRsEnv names the operator override: a comma-separated list of CIDRs the
// webhook client may reach despite being private/reserved (e.g. an on-prem integrator, or the Docker
// bridge in dev). Empty by default, so only public destinations are allowed.
const enrollmentWebhookAllowedCIDRsEnv = "SHELLHUB_INSTALL_KEY_WEBHOOK_ALLOWED_CIDRS"

// enrollmentWebhookClient builds the HTTP client used to POST to a namespace admin's webhook URL. It
// dials through an SSRF guard so a webhook can't be steered at the host's own network or the cloud
// metadata endpoint (169.254.169.254): the destination IP is checked at dial time, after DNS
// resolution, so a rebinding record can't sneak past. Any port is allowed (webhooks are not bound to
// 80/443); the risk is the address, not the port.
func enrollmentWebhookClient() *http.Client {
	opts := []ssrf.Option{ssrf.WithAnyPort()}

	var v4, v6 []netip.Prefix

	for _, raw := range strings.Split(envs.DefaultBackend.Get(enrollmentWebhookAllowedCIDRsEnv), ",") {
		raw = strings.TrimSpace(raw)
		if raw == "" {
			continue
		}

		prefix, err := netip.ParsePrefix(raw)
		if err != nil {
			log.WithError(err).WithField("cidr", raw).Warn("ignoring invalid webhook allowed CIDR")

			continue
		}

		if prefix.Addr().Is4() {
			v4 = append(v4, prefix)
		} else {
			v6 = append(v6, prefix)
		}
	}

	if len(v4) > 0 {
		opts = append(opts, ssrf.WithAllowedV4Prefixes(v4...))
	}

	if len(v6) > 0 {
		opts = append(opts, ssrf.WithAllowedV6Prefixes(v6...))
	}

	guard := ssrf.New(opts...)

	return &http.Client{
		Transport: &http.Transport{
			DialContext: (&net.Dialer{Control: guard.Safe}).DialContext,
		},
	}
}

// enrollmentDecision is the outcome of an install key's enrollment policy for a device: it maps
// directly to the device's initial status.
type enrollmentDecision string

const (
	enrollAccept  enrollmentDecision = "accept"
	enrollReject  enrollmentDecision = "reject"
	enrollPending enrollmentDecision = "pending"
)

// evaluateEnrollment runs the install key's mode as a policy and returns the enrollment decision for
// the device. A nil key (no key presented and no legacy key resolved) defaults to pending for manual
// review, preserving the historical keyless behavior.
func (s *service) evaluateEnrollment(ctx context.Context, key *models.InstallKey, req requests.DeviceAuth, uid, hostname string, paired bool) enrollmentDecision {
	if key == nil {
		return enrollPending
	}

	// The pairing-code flow is its own acceptance: the user already approved by entering the code, and
	// the whole stack (the agent's accepted-only waiter, the code-accept UI) assumes pairing == accepted.
	// So a paired enrollment accepts outright, ignoring the pairing key's mode — never firing the
	// webhook POST, the allowlist reject, or a manual/pending hold, which would hang the agent or lie in
	// the UI. The pairing key's mode is meaningless by construction.
	if paired {
		return enrollAccept
	}

	switch key.Mode {
	case models.InstallKeyModeAutomatic:
		return enrollAccept
	case models.InstallKeyModeManual:
		return enrollPending
	case models.InstallKeyModeAllowlist:
		mac := strings.ToLower(strings.TrimSpace(req.Identity.MAC))
		for _, allowed := range key.AllowedMACs {
			if allowed == mac {
				return enrollAccept
			}
		}

		return enrollReject
	case models.InstallKeyModeWebhook:
		// Hand the integrator a signed, scoped, expiring callback URL so it can defer and decide later
		// without any standing credential.
		callbackURL := s.enrollmentCallbackURL(key, req, uid)

		decision, err := s.callEnrollmentWebhook(ctx, key, req, uid, hostname, callbackURL)
		if err != nil {
			// Fail closed: a webhook error/timeout must never open the door. Leave the device pending
			// for manual review.
			log.WithError(err).WithField("install_key", key.Name).Warn("enrollment webhook failed; device remains pending")

			return enrollPending
		}

		return decision
	default:
		// An unknown/empty mode is treated as manual: the safe default is human review.
		return enrollPending
	}
}

// applyEnrollmentDecision carries out a fresh enrollment's policy decision on the just-created (or
// re-registered) pending device and returns the resulting status. accept goes through the canonical
// UpdateDeviceStatus so the license gate, billing report, MAC-merge, counters and the install-key
// event all run; reject transitions the device to rejected through the same path (no license gate);
// pending leaves it for manual review. A keyless enrollment (nil key) has no key to consume or record.
func (s *service) applyEnrollmentDecision(ctx context.Context, decision enrollmentDecision, key *models.InstallKey, req requests.DeviceAuth, uid, hostname string, reRegistration, record bool) models.DeviceStatus {
	// Append the append-only enrollment history event before applying the decision. UpdateDeviceStatus
	// stamps the device's newest event with the accept/reject outcome, so the event must already exist
	// for a fresh automatic/allowlist enrollment or the decision is lost. A reconcile re-evaluation
	// passes record=false: it is not a new enrollment, so it appends nothing and the stamp lands on the
	// device's original event instead. A keyless enrollment (nil key) has nothing to record against.
	if record {
		s.recordEnrollment(ctx, key, req, uid, hostname, reRegistration)
	}

	switch decision {
	case enrollAccept:
		// Reserve a use before accepting. The store guards the increment with the usage limit, so if
		// concurrent enrollments raced us to the last slot this fails and the device stays pending
		// instead of being accepted past the limit.
		if key != nil {
			if err := s.store.InstallKeyIncrementUsage(ctx, key); err != nil {
				log.WithError(err).WithField("install_key", key.Name).Warn("install key exhausted; device remains pending")

				return models.DeviceStatusPending
			}
		}

		acceptReq := &requests.DeviceUpdateStatus{
			TenantID: req.TenantID,
			UID:      uid,
			Status:   string(models.DeviceStatusAccepted),
		}
		if err := s.UpdateDeviceStatus(ctx, acceptReq); err != nil {
			// The accept failed after we reserved a use; return it so a later reconcile can retry
			// against a key that still has the slot.
			if key != nil {
				if releaseErr := s.store.InstallKeyDecrementUsage(ctx, key); releaseErr != nil {
					log.WithError(releaseErr).WithField("install_key", key.Name).Warn("failed to release reserved install key use")
				}
			}

			if errors.Is(err, ErrDeviceLicenseLimit) {
				log.WithError(err).WithField("device_uid", uid).Warn("license limit reached; device remains pending")
			} else {
				log.WithError(err).WithField("device_uid", uid).Warn("auto-accept failed; device remains pending")
			}

			return models.DeviceStatusPending
		}

		return models.DeviceStatusAccepted
	case enrollReject:
		rejectReq := &requests.DeviceUpdateStatus{
			TenantID: req.TenantID,
			UID:      uid,
			Status:   string(models.DeviceStatusRejected),
		}
		if err := s.UpdateDeviceStatus(ctx, rejectReq); err != nil {
			log.WithError(err).WithField("device_uid", uid).Warn("enrollment reject failed; device remains pending")

			return models.DeviceStatusPending
		}

		return models.DeviceStatusRejected
	default: // enrollPending
		return models.DeviceStatusPending
	}
}

// recordEnrollment appends the append-only history event for an enrollment, best-effort. It is a
// no-op for a keyless enrollment with no legacy key resolved.
func (s *service) recordEnrollment(ctx context.Context, key *models.InstallKey, req requests.DeviceAuth, uid, hostname string, reRegistration bool) {
	if key == nil {
		return
	}

	s.appendInstallKeyEvent(ctx, key, req, uid, hostname, reRegistration)
}

// reconcileEnrollment re-evaluates a still-pending device's enrollment policy on a later AuthDevice,
// giving a decision that couldn't be reached at first enrollment (a deferred or down webhook
// integrator, or an accept the license limit blocked) another chance whenever the agent phones home.
// It resolves the device's original enrollment key by its stored digest, applies only to modes that
// can leave a device recoverably pending, and is throttled by EnrollmentReconcileInterval. Mutations
// land on the passed device; the caller persists them via its DeviceUpdate.
func (s *service) reconcileEnrollment(ctx context.Context, device *models.Device, req requests.DeviceAuth, uid, hostname string) {
	if device.InstallKeyID == "" {
		return
	}

	if device.LastEnrollmentAttemptAt != nil && clock.Now().Sub(*device.LastEnrollmentAttemptAt) < models.EnrollmentReconcileInterval {
		return
	}

	key, err := s.store.InstallKeyResolve(ctx, store.InstallKeyIDResolver, device.InstallKeyID, s.store.Options().InNamespace(req.TenantID))
	// IsValid gates the same as a fresh enrollment: a key revoked, disabled, expired, or exhausted after
	// the device landed pending must not accept it on a later phone-home.
	if err != nil || key == nil || !key.IsValid() || !key.ReconcilableOnAuth() {
		return
	}

	// Stamp the attempt before evaluating so a defer-again or a transport error still backs off.
	now := clock.Now()
	device.LastEnrollmentAttemptAt = &now

	// A reconcile re-evaluates a still-pending device (webhook/allowlist); a pairing enrollment never
	// reaches this path (it accepts outright), so paired is always false here.
	status := s.applyEnrollmentDecision(ctx, s.evaluateEnrollment(ctx, key, req, uid, hostname, false), key, req, uid, hostname, false, false)
	if status != models.DeviceStatusPending {
		device.Status = status
		device.StatusUpdatedAt = clock.Now()
	}
}

// enrollmentWebhookRequest is the signed body POSTed to a webhook-mode key's integrator endpoint.
type enrollmentWebhookRequest struct {
	TenantID string `json:"tenant_id"`
	// InstallKeyID is the key's digest: a stable, non-secret identifier the integrator keys its policy
	// off (it survives a rename; the name can't). InstallKeyName rides along only for human
	// recognizability in the integrator's logs/UI — never key policy on it. The secret itself is never
	// sent, not even a fragment: authenticity comes from the X-ShellHub-Signature HMAC instead.
	InstallKeyID   string               `json:"install_key_id"`
	InstallKeyName string               `json:"install_key_name"`
	DeviceUID      string               `json:"device_uid"`
	MAC            string               `json:"mac"`
	Hostname       string               `json:"hostname"`
	Info           *requests.DeviceInfo `json:"info,omitempty"`
	SourceIP       string               `json:"source_ip"`
	Timestamp      time.Time            `json:"timestamp"`
	// CallbackURL is a ready, token-authenticated URL the integrator can POST its decision to later
	// (after answering "defer"). Omitted when the public base is unknown.
	CallbackURL string `json:"callback_url,omitempty"`
}

// enrollDeferDecision is the special webhook response meaning "I'll decide later via the callback URL".
// It lands the device pending, like the default, but signals async intent.
const enrollDeferDecision = "defer"

// enrollmentWebhookResponse is the integrator's answer: accept/reject/defer, with an optional
// human-readable reason.
type enrollmentWebhookResponse struct {
	Decision string `json:"decision"`
	Reason   string `json:"reason,omitempty"`
}

// signEnrollmentWebhook returns the hex HMAC-SHA256 of the body under the key's secret, sent in the
// X-ShellHub-Signature header so the integrator can verify the request came from ShellHub.
func signEnrollmentWebhook(secret string, body []byte) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(body)

	return hex.EncodeToString(mac.Sum(nil))
}

// enrollmentCallbackURL mints a signed, scoped, expiring callback token for the device and builds the
// absolute URL the integrator can redeem it at. It returns "" when the public base is unknown (no
// forwarded host) or the token can't be minted, in which case no callback URL is offered.
func (s *service) enrollmentCallbackURL(key *models.InstallKey, req requests.DeviceAuth, uid string) string {
	if req.ForwardedHost == "" {
		return ""
	}

	proto := req.ForwardedProto
	if proto == "" {
		proto = "https"
	}

	ttl := time.Duration(key.WebhookCallbackTTLOrDefault()) * time.Second
	token, err := jwttoken.EncodeEnrollmentDecisionClaims(jwttoken.EnrollmentDecisionClaims{
		DeviceUID:    uid,
		TenantID:     key.TenantID,
		InstallKeyID: key.ID,
	}, ttl, s.privKey)
	if err != nil {
		log.WithError(err).Warn("failed to mint enrollment callback token")

		return ""
	}

	return fmt.Sprintf("%s://%s/api/devices/enroll/callback/%s", proto, req.ForwardedHost, token)
}

// callEnrollmentWebhook POSTs the signed enrollment payload to the key's webhook URL and returns the
// integrator's decision. Any transport error, non-2xx status, unparseable body, or unrecognized
// decision returns an error so the caller can fail closed to pending.
func (s *service) callEnrollmentWebhook(ctx context.Context, key *models.InstallKey, req requests.DeviceAuth, uid, hostname, callbackURL string) (enrollmentDecision, error) {
	payload := enrollmentWebhookRequest{
		TenantID:       key.TenantID,
		InstallKeyID:   key.ID,
		InstallKeyName: key.Name,
		DeviceUID:      uid,
		MAC:            req.Identity.MAC,
		Hostname:       hostname,
		Info:           req.Info,
		SourceIP:       req.RealIP,
		Timestamp:      clock.Now(),
		CallbackURL:    callbackURL,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}

	ctx, cancel := context.WithTimeout(ctx, time.Duration(key.WebhookTimeoutOrDefault())*time.Second)
	defer cancel()

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, key.WebhookURL, bytes.NewReader(body))
	if err != nil {
		return "", err
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("X-ShellHub-Signature", signEnrollmentWebhook(key.WebhookSecret, body))

	resp, err := enrollmentWebhookClient().Do(httpReq)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close() //nolint:errcheck

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return "", fmt.Errorf("enrollment webhook returned status %d", resp.StatusCode)
	}

	var decoded enrollmentWebhookResponse
	if err := json.NewDecoder(resp.Body).Decode(&decoded); err != nil {
		return "", err
	}

	switch decoded.Decision {
	case string(enrollAccept):
		return enrollAccept, nil
	case string(enrollReject):
		return enrollReject, nil
	case enrollDeferDecision, string(enrollPending):
		// "defer" is the async signal; "pending" is accepted as an explicit "leave for manual". Both
		// land the device pending (the callback URL stays valid either way).
		return enrollPending, nil
	default:
		return "", fmt.Errorf("enrollment webhook returned invalid decision %q", decoded.Decision)
	}
}
