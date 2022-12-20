package reporters

import (
	"github.com/shellhub-io/shellhub/api/store"
	"github.com/shellhub-io/shellhub/pkg/models"
)

type Reporters interface {
	SessionUnauthenticated(session *models.Session) error
}

type Reporter struct {
	store store.AnnouncementsStore
}

func NewReporter(store store.Store) *Reporter {
	return &Reporter{
		store: store,
	}
}

func (r *Reporter) SessionUnauthenticated(session *models.Session) error {
	// TODO: implemen me.
	return nil
}
