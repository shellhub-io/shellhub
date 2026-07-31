package main

// What this proxy can do, chosen one module at a time. Caddy registers each in
// an init(), so this list is the feature set: naming them individually rather
// than importing modules/standard keeps out an ACME server, a PHP gateway, a
// template engine with a syntax highlighter and an OpenTelemetry exporter, none
// of which an edge that routes requests has any use for -- and all of which
// would be ours to patch, since they compile into this binary.
//
// A module that is needed but missing fails when the configuration is adapted,
// which TestCaddyfileAdapts does for every shape the environment can take.
import (
	_ "github.com/caddyserver/caddy/v2/caddyconfig/caddyfile"
	_ "github.com/caddyserver/caddy/v2/modules/caddyhttp"
	_ "github.com/caddyserver/caddy/v2/modules/caddyhttp/encode"
	_ "github.com/caddyserver/caddy/v2/modules/caddyhttp/encode/gzip"
	_ "github.com/caddyserver/caddy/v2/modules/caddyhttp/encode/zstd"
	_ "github.com/caddyserver/caddy/v2/modules/caddyhttp/headers"
	_ "github.com/caddyserver/caddy/v2/modules/caddyhttp/logging"
	_ "github.com/caddyserver/caddy/v2/modules/caddyhttp/map"
	_ "github.com/caddyserver/caddy/v2/modules/caddyhttp/proxyprotocol"
	_ "github.com/caddyserver/caddy/v2/modules/caddyhttp/reverseproxy"
	_ "github.com/caddyserver/caddy/v2/modules/caddyhttp/rewrite"
	_ "github.com/caddyserver/caddy/v2/modules/caddypki"
	_ "github.com/caddyserver/caddy/v2/modules/caddytls"
	_ "github.com/caddyserver/caddy/v2/modules/caddytls/standardstek"
	_ "github.com/caddyserver/caddy/v2/modules/filestorage"
	_ "github.com/caddyserver/caddy/v2/modules/logging"
	_ "github.com/caddyserver/caddy/v2/modules/metrics"

	// The DNS providers that can answer the wildcard's DNS-01 challenge. Adding
	// another is a line here.
	_ "github.com/caddy-dns/acmedns"
	_ "github.com/caddy-dns/cloudflare"
	_ "github.com/caddy-dns/digitalocean"
)
