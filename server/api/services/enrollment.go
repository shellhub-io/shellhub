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
	"slices"
	"strings"
	"time"

	"code.dny.dev/ssrf"
	"github.com/shellhub-io/shellhub/pkg/api/jwttoken"
	"github.com/shellhub-io/shellhub/pkg/api/requests"
	"github.com/shellhub-io/shellhub/pkg/api/scope"
	"github.com/shellhub-io/shellhub/pkg/clock"
	"github.com/shellhub-io/shellhub/pkg/envs"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/shellhub-io/shellhub/server/api/store"
	log "github.com/sirupsen/logrus"
)

const enrollmentWebhookAllowedCIDRsEnv = "SHELLHUB_INSTALL_KEY_WEBHOOK_ALLOWED_CIDRS"

func enrollmentWebhookClient() *http.Client {
	opts := []ssrf.Option{ssrf.WithAnyPort()}

	var v4, v6 []netip.Prefix

	for raw := range strings.SplitSeq(envs.DefaultBackend.Get(enrollmentWebhookAllowedCIDRsEnv), ",") {
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

type enrollmentDecision string

const (
	enrollAccept  enrollmentDecision = "accept"
	enrollReject  enrollmentDecision = "reject"
	enrollPending enrollmentDecision = "pending"
)

func (s *service) evaluateEnrollment(ctx context.Context, key *models.InstallKey, req requests.DeviceAuth, uid, hostname string, paired bool) enrollmentDecision {
	if key == nil {
		return enrollPending
	}

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
		if slices.Contains(key.AllowedMACs, mac) {
			return enrollAccept
		}

		return enrollReject
	case models.InstallKeyModeWebhook:
		callbackURL := s.enrollmentCallbackURL(key, req, uid)

		decision, err := s.callEnrollmentWebhook(ctx, key, req, uid, hostname, callbackURL)
		if err != nil {
			log.WithError(err).WithField("install_key", key.Name).Warn("enrollment webhook failed; device remains pending")

			return enrollPending
		}

		return decision
	default:
		return enrollPending
	}
}

func (s *service) applyEnrollmentDecision(ctx context.Context, decision enrollmentDecision, key *models.InstallKey, req requests.DeviceAuth, uid, hostname string, reRegistration, record bool) models.DeviceStatus {
	if record {
		s.recordEnrollment(ctx, key, req, uid, hostname, reRegistration)
	}

	switch decision {
	case enrollAccept:
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

func (s *service) recordEnrollment(ctx context.Context, key *models.InstallKey, req requests.DeviceAuth, uid, hostname string, reRegistration bool) {
	if key == nil {
		return
	}

	s.appendInstallKeyEvent(ctx, key, req, uid, hostname, reRegistration)
}

func (s *service) reconcileEnrollment(ctx context.Context, device *models.Device, req requests.DeviceAuth, uid, hostname string) {
	if device.InstallKeyID == "" {
		return
	}

	if device.LastEnrollmentAttemptAt != nil && clock.Now().Sub(*device.LastEnrollmentAttemptAt) < models.EnrollmentReconcileInterval {
		return
	}

	key, err := s.store.InstallKeyResolve(ctx, scope.MustBounded(device.TenantID), store.InstallKeyIDResolver, device.InstallKeyID)
	if err != nil || key == nil || !key.IsValid() || !key.ReconcilableOnAuth() {
		return
	}

	now := clock.Now()
	device.LastEnrollmentAttemptAt = &now

	status := s.applyEnrollmentDecision(ctx, s.evaluateEnrollment(ctx, key, req, uid, hostname, false), key, req, uid, hostname, false, false)
	if status != models.DeviceStatusPending {
		device.Status = status
		device.StatusUpdatedAt = clock.Now()
	}
}

type enrollmentWebhookRequest struct {
	TenantID       string               `json:"tenant_id"`
	InstallKeyID   string               `json:"install_key_id"`
	InstallKeyName string               `json:"install_key_name"`
	DeviceUID      string               `json:"device_uid"`
	MAC            string               `json:"mac"`
	Hostname       string               `json:"hostname"`
	Info           *requests.DeviceInfo `json:"info,omitempty"`
	SourceIP       string               `json:"source_ip"`
	Timestamp      time.Time            `json:"timestamp"`
	CallbackURL    string               `json:"callback_url,omitempty"`
}

const enrollDeferDecision = "defer"

type enrollmentWebhookResponse struct {
	Decision string `json:"decision"`
	Reason   string `json:"reason,omitempty"`
}

func signEnrollmentWebhook(secret string, body []byte) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(body)

	return hex.EncodeToString(mac.Sum(nil))
}

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
		return enrollPending, nil
	default:
		return "", fmt.Errorf("enrollment webhook returned invalid decision %q", decoded.Decision)
	}
}
