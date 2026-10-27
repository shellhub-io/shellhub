package services

import (
	"context"
	"testing"

	"github.com/shellhub-io/shellhub/pkg/clock"
	"github.com/shellhub-io/shellhub/pkg/models"
	storemock "github.com/shellhub-io/shellhub/server/api/store/mocks"
	"github.com/stretchr/testify/require"
)

func TestRecordDepartureRemovals(t *testing.T) {
	t.Run("a departure that removed no device records nothing", func(t *testing.T) {
		require.NoError(t, RecordDepartureRemovals(context.Background(), storemock.NewMockStore(t), clock.New(), &models.MemberDeparted{}))
	})

	t.Run("a departure with no tenant is refused rather than unbounded", func(t *testing.T) {
		err := RecordDepartureRemovals(context.Background(), storemock.NewMockStore(t), clock.New(), &models.MemberDeparted{RemovedDevices: []string{"uid"}})
		require.ErrorIs(t, err, ErrForbidden)
	})
}
