package main

import (
	"context"
	"testing"

	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/shellhub-io/shellhub/tests/environment"
	"github.com/stretchr/testify/assert"
)

func TestEnterpriseSessionPosition(t *testing.T) {
	ctx := context.Background()

	florianopolis := environment.Location{Country: "BR", Latitude: -27.5954, Longitude: -48.548}

	compose := newEnterpriseEnvironment(t, ctx, environment.New(t, run).WithEveryAddressAt(florianopolis))

	signer := registerDeviceKey(t, ctx, compose)
	_, device := startAcceptedAgent(t, ctx, compose)

	uid := finishSession(t, ctx, compose, device, signer, "true")

	assert.Equal(t,
		models.SessionPosition{Latitude: florianopolis.Latitude, Longitude: florianopolis.Longitude},
		getSession(t, ctx, compose, uid).Position,
		"the session records where GeoIP places the address it came from")
}
