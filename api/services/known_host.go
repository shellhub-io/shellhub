package services

import (
	"context"
	"errors"
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/shellhub-io/shellhub/api/pkg/responses"
	"github.com/shellhub-io/shellhub/api/store"
	"github.com/shellhub-io/shellhub/pkg/api/authorizer"
	"github.com/shellhub-io/shellhub/pkg/api/requests"
	"github.com/shellhub-io/shellhub/pkg/egress"
	"github.com/shellhub-io/shellhub/pkg/models"
	gossh "golang.org/x/crypto/ssh"
)

// ErrKnownHostUnreachable means the host key could not be scanned (host down,
// blocked by the egress guardian, or not speaking SSH).
var ErrKnownHostUnreachable = errors.New("could not read the host key")

// ErrKnownHostInvalidKey means the public key supplied to AcceptKnownHost could
// not be parsed.
var ErrKnownHostInvalidKey = errors.New("invalid host public key")

type KnownHostService interface {
	// ScanKnownHost reads the target's host key and reports it against the stored
	// one (unverified / trusted / changed).
	ScanKnownHost(ctx context.Context, req *requests.KnownHostScan) (*responses.KnownHostScanResult, error)
	// AcceptKnownHost trusts (stores) a host key for the target's scope.
	AcceptKnownHost(ctx context.Context, req *requests.KnownHostAccept) (*models.KnownHost, error)
	// GetKnownHost returns the stored known host for a target, or nil if none.
	GetKnownHost(ctx context.Context, req *requests.KnownHostGet) (*models.KnownHost, error)
	// DeleteKnownHost forgets the stored host key for the target's scope.
	DeleteKnownHost(ctx context.Context, req *requests.KnownHostDelete) error
}

// scopeOwner maps a request scope to the owner id used for storage: personal
// records belong to the caller; namespace (team) records are shared (empty).
func scopeOwner(scope, userID string) string {
	if scope == "personal" {
		return userID
	}

	return ""
}

// scanHostKey opens an SSH handshake to host:port just far enough to capture the
// presented host key, through the SSRF egress guardian. It never authenticates.
func scanHostKey(ctx context.Context, host string, port int) (*responses.KnownHostScanResult, error) {
	addr := net.JoinHostPort(host, strconv.Itoa(port))

	// Reuse the shared SSRF guardian; the host-key handshake wants a longer
	// connect timeout than the plain reachability probe, so override it here.
	dialer := egress.GuardedDialer(port)
	dialer.Timeout = 8 * time.Second

	conn, err := dialer.DialContext(ctx, "tcp", addr)
	if err != nil {
		if egress.IsBlocked(err) {
			return nil, ErrEgressBlocked
		}

		return nil, ErrKnownHostUnreachable
	}
	defer conn.Close() //nolint:errcheck

	var captured gossh.PublicKey
	config := &gossh.ClientConfig{ //nolint:exhaustruct
		User: "shellhub-probe",
		HostKeyCallback: func(_ string, _ net.Addr, key gossh.PublicKey) error {
			captured = key

			return nil
		},
		Timeout: 8 * time.Second,
	}

	// The handshake fails at authentication (no methods), but HostKeyCallback
	// runs first, so the key is captured regardless.
	if sshConn, chans, reqs, err := gossh.NewClientConn(conn, addr, config); err == nil {
		go gossh.DiscardRequests(reqs)
		go func() {
			for ch := range chans {
				ch.Reject(gossh.Prohibited, "") //nolint:errcheck
			}
		}()
		sshConn.Close() //nolint:errcheck
	}

	if captured == nil {
		return nil, ErrKnownHostUnreachable
	}

	return &responses.KnownHostScanResult{
		KeyType:     captured.Type(),
		Fingerprint: gossh.FingerprintSHA256(captured),
		PublicKey:   strings.TrimSpace(string(gossh.MarshalAuthorizedKey(captured))),
	}, nil
}

func (s *service) ScanKnownHost(ctx context.Context, req *requests.KnownHostScan) (*responses.KnownHostScanResult, error) {
	result, err := scanHostKey(ctx, req.Host, req.Port)
	if err != nil {
		return nil, err
	}

	stored, err := s.store.KnownHostResolve(ctx, req.TenantID, scopeOwner(req.Scope, req.UserID), req.Host, req.Port)
	switch {
	case err == nil:
		result.Stored = stored
		if stored.Fingerprint == result.Fingerprint {
			result.Status = models.KnownHostTrusted
		} else {
			result.Status = models.KnownHostChanged
		}
	case errors.Is(err, store.ErrNoDocuments):
		result.Status = models.KnownHostUnverified
	default:
		return nil, err
	}

	return result, nil
}

func (s *service) AcceptKnownHost(ctx context.Context, req *requests.KnownHostAccept) (*models.KnownHost, error) {
	owner := scopeOwner(req.Scope, req.UserID)

	// Any shared (team) trust write needs operator+: a member must not be able to
	// plant or change a host key that every other member's connects will trust.
	// Personal records are the caller's own, so any role may write them.
	if req.Scope == "namespace" && !req.Role.HasAuthority(authorizer.RoleOperator) {
		return nil, ErrAuthForbidden
	}

	// Trust only key material the server can parse, and derive the type and
	// fingerprint from the key itself — never from the client-supplied fields, so
	// a caller can't store a fingerprint that doesn't match the stored key.
	parsed, _, _, _, err := gossh.ParseAuthorizedKey([]byte(req.PublicKey))
	if err != nil {
		return nil, ErrKnownHostInvalidKey
	}

	knownHost := &models.KnownHost{
		TenantID:    req.TenantID,
		OwnerID:     owner,
		Host:        req.Host,
		Port:        req.Port,
		KeyType:     parsed.Type(),
		PublicKey:   strings.TrimSpace(string(gossh.MarshalAuthorizedKey(parsed))),
		Fingerprint: gossh.FingerprintSHA256(parsed),
		AcceptedBy:  req.UserID,
	}

	if err := s.store.KnownHostUpsert(ctx, knownHost); err != nil {
		return nil, err
	}

	return knownHost, nil
}

func (s *service) GetKnownHost(ctx context.Context, req *requests.KnownHostGet) (*models.KnownHost, error) {
	knownHost, err := s.store.KnownHostResolve(ctx, req.TenantID, scopeOwner(req.Scope, req.UserID), req.Host, req.Port)
	if err != nil {
		if errors.Is(err, store.ErrNoDocuments) {
			return nil, nil //nolint:nilnil
		}

		return nil, err
	}

	return knownHost, nil
}

func (s *service) DeleteKnownHost(ctx context.Context, req *requests.KnownHostDelete) error {
	if req.Scope == "namespace" && !req.Role.HasAuthority(authorizer.RoleOperator) {
		return ErrAuthForbidden
	}

	err := s.store.KnownHostDelete(ctx, req.TenantID, scopeOwner(req.Scope, req.UserID), req.Host, req.Port)
	if err != nil && errors.Is(err, store.ErrNoDocuments) {
		return NewErrKnownHostNotFound(req.Host, err)
	}

	return err
}
