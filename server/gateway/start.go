package gateway

import (
	"fmt"
	"net/http"
	"sync"

	"github.com/caddyserver/caddy/v2"
	"github.com/caddyserver/caddy/v2/caddyconfig"
	log "github.com/sirupsen/logrus"
)

var loading sync.Mutex

// Start builds the Caddyfile for cfg and runs it in this process, serving the server's routes
// with handler, and replaces whatever an earlier Start loaded. It returns once the listeners are
// up: ErrNoServerHandler for a nil handler, the template's or the adapter's error for a
// configuration that does not build, and Caddy's for one it cannot load, such as a port already
// taken.
func Start(cfg *GatewayConfig, handler http.Handler) error {
	if handler == nil {
		return ErrNoServerHandler
	}

	loading.Lock()
	defer loading.Unlock()

	rendered, err := Caddyfile(cfg)
	if err != nil {
		return err
	}

	adapted, warnings, err := caddyconfig.GetAdapter("caddyfile").Adapt(rendered, nil)
	if err != nil {
		return fmt.Errorf("failed to adapt the configuration: %w", err)
	}

	for _, warning := range warnings {
		log.WithField("line", warning.Line).Warn(warning.Message)
	}

	serverHandler = handler

	return caddy.Load(adapted, true)
}

// Stop shuts down what Start loaded, closing its listeners, and returns Caddy's error when an app
// fails to stop. Calling it with nothing running is not an error.
func Stop() error {
	loading.Lock()
	defer loading.Unlock()

	serverHandler = nil

	return caddy.Stop()
}
