package agentd

import (
	"context"
	"errors"
	"time"

	"github.com/shellhub-io/shellhub/pkg/clock"
	"github.com/shellhub-io/shellhub/pkg/models"
	log "github.com/sirupsen/logrus"
)

// ErrPairingExpired is returned by [AwaitPairing] when the pairing code expires before the device
// is accepted.
var ErrPairingExpired = errors.New("the pairing code expired before the device was accepted")

// AwaitPairing calls status every interval until it reports the device accepted, and returns the
// tenant it was accepted into. The code expires expiresIn seconds after clk's current time, and
// each pending status that reports the seconds it has left moves that deadline, so the wait lasts
// as long as the server renews the code. A status from a server that reports none, or a failed
// status call, which is logged, leaves the deadline where it was.
//
// It returns [ErrPairingExpired] once clk passes the deadline, and ctx's error when ctx ends first.
func AwaitPairing(ctx context.Context, clk clock.Clock, interval time.Duration, expiresIn int, status func() (*models.DevicePairingStatus, error)) (string, error) {
	deadline := clk.Now().Add(time.Duration(expiresIn) * time.Second)

	for clk.Now().Before(deadline) {
		select {
		case <-ctx.Done():
			return "", ctx.Err()
		case <-time.After(interval):
		}

		current, err := status()
		if err != nil {
			log.WithError(err).Warn("Failed to get the pairing status")

			continue
		}

		if current.Status == models.DeviceStatusAccepted && current.TenantID != "" {
			return current.TenantID, nil
		}

		if current.ExpiresIn > 0 {
			deadline = clk.Now().Add(time.Duration(current.ExpiresIn) * time.Second)
		}
	}

	return "", ErrPairingExpired
}
