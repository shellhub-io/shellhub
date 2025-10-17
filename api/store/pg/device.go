package pg

import (
	"context"
	"time"

	"github.com/shellhub-io/shellhub/api/store"
	"github.com/shellhub-io/shellhub/pkg/models"
)

func (pg *Pg) DeviceCreate(ctx context.Context, device *models.Device) (insertedUID string, err error)

func (pg *Pg) DeviceList(ctx context.Context, acceptable store.DeviceAcceptable, opts ...store.QueryOption) ([]models.Device, int, error)

func (pg *Pg) DeviceResolve(ctx context.Context, resolver store.DeviceResolver, value string, opts ...store.QueryOption) (*models.Device, error)

func (pg *Pg) DeviceConflicts(ctx context.Context, target *models.DeviceConflicts) (conflicts []string, has bool, err error)

func (pg *Pg) DeviceUpdate(ctx context.Context, device *models.Device) error

func (pg *Pg) DeviceHeartbeat(ctx context.Context, uids []string, lastSeen time.Time) (modifiedCount int64, err error)

func (pg *Pg) DeviceDelete(ctx context.Context, device *models.Device) error

func (pg *Pg) DeviceDeleteMany(ctx context.Context, uids []string) (deletedCount int64, err error)
