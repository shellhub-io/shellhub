# Gateway

The `gateway` is ShellHub's entry point. It terminates TLS and routes inbound HTTP requests and
WebSocket connections to the server and the console.

Routing is all it does. Authentication, the root aliases and the per-device tunnel subdomains are
resolved by the API itself, which is what lets a Kubernetes deployment expose the server and the
console on an Ingress with this proxy nowhere in the path.

## Architecture

The binary is the proxy: it embeds [Caddy](https://caddyserver.com) as a library rather than
supervising a separate process. Caddy registers its features as modules in `init()`, so what this
proxy can do is the import list in `modules.go` — which is why adding a DNS provider is a line of
Go here rather than a rebuild step with `xcaddy`.

At startup the binary builds a Caddyfile from the environment, adapts it to Caddy's JSON
configuration and loads it. Nothing is written to disk and nothing reloads it afterwards: a
different configuration means a different container.

## Configuration

Every setting comes from an environment variable, parsed into `GatewayConfig` in `config.go` and
laid out by `Caddyfile.tmpl`.

To read what an environment produces, ask the binary:

```
docker compose exec gateway /gateway config
```

### TLS

With `SHELLHUB_AUTO_SSL` enabled, Caddy obtains and renews certificates itself, answering the
ACME challenge in memory. The wildcard that serves web endpoints is proven over DNS-01, using the
provider named by `SHELLHUB_WEB_ENDPOINTS_DNS_PROVIDER`.

`SHELLHUB_ACME_CA_SERVER` chooses which CA to ask. Leave it empty for Let's Encrypt; point it at
their staging directory to rehearse without spending the real rate limit, or at a private CA.
Staging certificates are signed by a root nothing trusts, so they validate the flow and never
serve traffic — and switching back to production means clearing the storage first, or Caddy keeps
using the account and certificates it already has.

The account key and the certificates live in `/data`, which the image declares as a volume.
Losing it means asking the CA for everything again, and Let's Encrypt caps how often it will
answer, so a deployment that cares should mount a named volume there rather than rely on the
anonymous one Docker creates.

`SHELLHUB_TLS_CERT_FILE` and `SHELLHUB_TLS_KEY_FILE` serve a certificate the operator supplies
instead of obtaining one, which is the only way to run HTTPS on a name no public authority will
sign: an internal hostname, or a domain this deployment does not own. Both are paths inside the
gateway container, so whatever holds them has to be mounted there.

They are read only when `SHELLHUB_AUTO_SSL` is enabled, and they are all or nothing: setting one
without the other stops the gateway from starting, checked even with automatic HTTPS off. A
half-set pair is a typo in every case, and the alternative to refusing it is quietly asking a
public CA for a name the operator meant to serve themselves.

### Behind a load balancer

`SHELLHUB_PROXY` enables the PROXY protocol, and `SHELLHUB_PROXY_TRUSTED_IPS` names the peers
allowed to use it. This list matters: a peer that may send the preamble is a peer that chooses
the client's address, and that address is what the login lockout and the GeoIP rules are keyed
on. The default trusts anyone, which is what the previous configuration did.

## Development

`air` rebuilds the binary on change, including changes to `Caddyfile.tmpl` — the template is
compiled in, so a rebuild is what applies it.
