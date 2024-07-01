package mongo

import (
	"context"

	"github.com/shellhub-io/shellhub/api/store"
	"go.mongodb.org/mongo-driver/mongo"
)

func (s *Store) WithTransaction(ctx context.Context, cb store.TransactionCb) error {
	session, err := s.db.Client().StartSession()
	if err != nil {
		return store.ErrStartTransactionFailed
	}
	defer session.EndSession(ctx)

	// The [session.WithTransaction] function expects a callback that returns an [interface{}] and an error.
	// To meet this requirement, we need to wrap our cb so that it always returns nil  as the [interface{}],
	// along with the error from our callback function.
	fn := func(ctx mongo.SessionContext) (interface{}, error) {
		return nil, cb(ctx)
	}

	_, err = session.WithTransaction(ctx, fn)

	return err
}
