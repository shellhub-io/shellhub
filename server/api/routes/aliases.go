package routes

// rootAliases maps the paths ShellHub publishes at the root of its domain onto
// the canonical routes under /api. Clients hard-code them: the console's
// generated SDK is built with a base URL that carries no /api prefix and calls
// /info and /healthcheck while booting, and the install script is published as
// <domain>/install.sh, an address printed in the documentation and embedded in
// every agent installed from it.
//
// The edge proxy rewrites them today, which is why the API only ever saw the
// canonical form. Serving them here means the API answers the same URLs on its
// own, so it can sit behind any ingress, or behind none.
//
// Rules are anchored with ^ so each matches a whole request line. Unanchored,
// they are suffix matches, and would also catch a path a device serves through
// a web endpoint tunnel.
var rootAliases = map[string]string{
	"^/healthcheck":   "/api" + HealthCheckURL,
	"^/healthcheck?*": "/api" + HealthCheckURL + "?$1",
	"^/info":          "/api" + GetSystemInfoURL,
	"^/info?*":        "/api" + GetSystemInfoURL + "?$1",
	// The install script reads tenant_id, preferred_hostname and friends from
	// the query string, so it has to survive the rewrite.
	"^/install.sh":   "/api" + GetSystemDownloadInstallScriptURL,
	"^/install.sh?*": "/api" + GetSystemDownloadInstallScriptURL + "?$1",
}
