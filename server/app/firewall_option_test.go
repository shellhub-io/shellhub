package app

import (
	"context"
	"testing"

	"github.com/shellhub-io/shellhub/pkg/cache"
	"github.com/shellhub-io/shellhub/pkg/geoip"
	"github.com/shellhub-io/shellhub/server/api/services"
	"github.com/shellhub-io/shellhub/server/api/store"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFirewallEvaluatorGetsTheChosenLocator(t *testing.T) {
	previous := services.FirewallEvaluatorFactory()
	t.Cleanup(func() { services.RegisterFirewallEvaluator(previous) })

	var received geoip.Locator
	services.RegisterFirewallEvaluator(func(_ context.Context, _ store.Store, _ cache.Cache, locator geoip.Locator) (services.FirewallEvaluator, error) {
		received = locator

		return nil, nil
	})

	chosen := geoip.NewNullGeoLite()
	_, err := (&Server{}).firewallEvaluatorOption(context.Background(), nil, nil, chosen)
	require.NoError(t, err)

	assert.Same(t, chosen, received)
}
