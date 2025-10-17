package pg

import (
	"errors"

	"github.com/shellhub-io/shellhub/api/store"
	"github.com/shellhub-io/shellhub/pkg/api/query"
	"github.com/shellhub-io/shellhub/pkg/models"
)

var ErrQueryNotFound = errors.New("query not found in context") // ErrQueryNotFound is returned when the query context value is not found or has the wrong type

func (pg *Pg) Options() store.QueryOptions {
	return pg.options
}

func (*queryOptions) Paginate(page *query.Paginator) store.QueryOption

func (*queryOptions) Sort(sorter *query.Sorter) store.QueryOption

func (*queryOptions) Match(filters *query.Filters) store.QueryOption

func (*queryOptions) WithMember(userID string) store.QueryOption

func (*queryOptions) InNamespace(namespaceID string) store.QueryOption

func (*queryOptions) WithDeviceStatus(status models.DeviceStatus) store.QueryOption
