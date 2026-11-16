package cache

import (
	"context"
	"time"

	"github.com/shellhub-io/shellhub/pkg/uuid"
	log "github.com/sirupsen/logrus"
)

// AdmitOne reserves key in c for ttl so that one caller at a time runs what it guards, such as one
// password guess per account. It reports whether the caller was admitted, and returns c's error,
// unadmitted, when the reservation cannot be taken. release gives the reservation back; it is a
// no-op for a caller not admitted, deletes the key only while it still holds this caller's
// reservation, so a caller whose ttl ran out cannot release the one taken after it, and runs
// without ctx's cancellation, so a cancelled request still frees the key. A failed release is
// logged and the key expires with ttl.
func AdmitOne(ctx context.Context, c Cache, key string, ttl time.Duration) (release func(), admitted bool, err error) {
	owner := uuid.Generate()

	admitted, err = c.SetNX(ctx, key, owner, ttl)
	if err != nil || !admitted {
		return func() {}, false, err
	}

	return func() {
		if _, err := c.CompareAndDelete(context.WithoutCancel(ctx), key, owner); err != nil {
			log.WithError(err).Warn("unable to release an in-flight reservation")
		}
	}, true, nil
}
