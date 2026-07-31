package main

// The proxy's feature set, chosen by import. Caddy registers every module in an
// init(), so this list is the whole of what xcaddy would otherwise build for us
// -- and since this binary is ours, adding a DNS provider is a line here rather
// than a rebuild step a customer has to run.
import (
	_ "github.com/caddy-dns/acmedns"
	_ "github.com/caddy-dns/cloudflare"
	_ "github.com/caddy-dns/digitalocean"
	_ "github.com/caddyserver/caddy/v2/modules/standard"
)
