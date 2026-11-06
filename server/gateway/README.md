# Gateway

The `gateway` package is ShellHub's entry point, and it runs inside the server process. It
terminates TLS, obtains certificates, and hands every API, agent tunnel and web terminal request to
the server's router directly. Only the console is proxied, to the `ui` container.

The server starts it when `SHELLHUB_GATEWAY` is `true`, which Docker Compose sets by default. Leave
it `false` when a load balancer or a Kubernetes ingress fronts the server: the router is still
served on `:8080`, and authentication, the root aliases and the per-device tunnel subdomains are
resolved by the API itself.

## Architecture

The package embeds [Caddy](https://caddyserver.com) as a library. Caddy registers its features as
modules in `init()`, so what the gateway can do is the import list in `modules.go`, which is why
adding a DNS provider is a line of Go here rather than a rebuild step with `xcaddy`.

`server.go` adds one module of its own, the `shellhub_server` handler. It ends a route by calling
the `http.Handler` passed to `Start(cfg, handler)`, the server's router, in the same process. No
connection to the server is opened per request or per tunnel, so a device costs one socket and no
copy buffers. The client's address reaches the router on the request itself: `X-Real-IP`,
`X-Forwarded-For`, `X-Forwarded-Host` and `X-Forwarded-Proto` are set from the connection, never
taken from the client.

At start-up the server builds a Caddyfile from the environment, adapts it to Caddy's JSON
configuration and loads it. Nothing is written to disk and nothing reloads it afterwards: a
different configuration means restarting the server.

## Configuration

Every setting comes from an environment variable, parsed into `GatewayConfig` in `config.go` and
laid out by `Caddyfile.tmpl`.

### TLS

With `SHELLHUB_AUTO_SSL` enabled, Caddy obtains and renews certificates itself, answering the
ACME challenge in memory. The wildcard that serves web endpoints is proven over DNS-01, using the
provider named by `SHELLHUB_WEB_ENDPOINTS_DNS_PROVIDER`.

`SHELLHUB_ACME_CA_SERVER` chooses which CA to ask. Leave it empty for Let's Encrypt; point it at
their staging directory to rehearse without spending the real rate limit, or at a private CA.
Staging certificates are signed by a root nothing trusts, so they validate the flow and never
serve traffic. Switching back to production means clearing the storage first, or Caddy keeps
using the account and certificates it already has.

The account key and the certificates live in `/data`, which Docker Compose mounts from the
`gateway_data` volume. Losing it means asking the CA for everything again, and Let's Encrypt caps
how often it will answer.

`SHELLHUB_TLS_CERT_FILE` and `SHELLHUB_TLS_KEY_FILE` serve a certificate the operator supplies
instead of obtaining one, which is the only way to run HTTPS on a name no public authority will
sign: an internal hostname, or a domain this deployment does not own. Both are paths inside the
server container, so whatever holds them has to be mounted there.

They are read only when `SHELLHUB_AUTO_SSL` is enabled, and they are all or nothing: setting one
without the other stops the server from starting, checked even with automatic HTTPS off. A
half-set pair is a typo in every case, and the alternative to refusing it is quietly asking a
public CA for a name the operator meant to serve themselves.

### Behind a load balancer

`SHELLHUB_PROXY` enables the PROXY protocol, and `SHELLHUB_PROXY_TRUSTED_IPS` names the peers
allowed to use it. This list matters: a peer that may send the preamble is a peer that chooses
the client's address, and that address is what the login lockout and the GeoIP rules are keyed
on. It has no default: `LoadGatewayConfig` refuses `SHELLHUB_PROXY=true` without it. The router
reads the same list, through `server/api/pkg/origin`, to decide which peers of `:8080` may name
the client in `X-Real-IP` and `X-Forwarded-*`.

## Development

The server's `air` watches this directory, so a change here, `Caddyfile.tmpl` included, rebuilds
and restarts the server. The template is compiled in, so a rebuild is what applies it.
