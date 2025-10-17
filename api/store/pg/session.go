package pg

import (
	"context"

	"github.com/shellhub-io/shellhub/api/store"
	"github.com/shellhub-io/shellhub/pkg/models"
)

func (pg *Pg) SessionList(ctx context.Context, opts ...store.QueryOption) ([]models.Session, int, error)

func (pg *Pg) SessionGet(ctx context.Context, uid models.UID) (*models.Session, error)

func (pg *Pg) SessionCreate(ctx context.Context, session models.Session) (*models.Session, error)

func (pg *Pg) SessionUpdate(ctx context.Context, uid models.UID, sess *models.Session, update *models.SessionUpdate) error

func (pg *Pg) SessionSetLastSeen(ctx context.Context, uid models.UID) error

func (pg *Pg) SessionDeleteActives(ctx context.Context, uid models.UID) error

func (pg *Pg) SessionUpdateDeviceUID(ctx context.Context, oldUID models.UID, newUID models.UID) error

func (pg *Pg) SessionSetRecorded(ctx context.Context, uid models.UID, recorded bool) error

func (pg *Pg) SessionSetType(ctx context.Context, uid models.UID, kind string) error

func (pg *Pg) SessionCreateActive(ctx context.Context, uid models.UID, session *models.Session) error

func (pg *Pg) SessionEvent(ctx context.Context, uid models.UID, event *models.SessionEvent) error

func (pg *Pg) SessionListEvents(ctx context.Context, uid models.UID, seat int, event models.SessionEventType, opts ...store.QueryOption) ([]models.SessionEvent, int, error)

func (pg *Pg) SessionDeleteEvents(ctx context.Context, uid models.UID, seat int, event models.SessionEventType) error
