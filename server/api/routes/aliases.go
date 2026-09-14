package routes

var rootAliases = map[string]string{
	"^/healthcheck":    "/api" + HealthCheckURL,
	"^/healthcheck?*":  "/api" + HealthCheckURL + "?$1",
	"^/info":           "/api" + GetSystemInfoURL,
	"^/info?*":         "/api" + GetSystemInfoURL + "?$1",
	"^/install.sh":     "/api" + GetSystemDownloadInstallScriptURL,
	"^/install.sh?*":   "/api" + GetSystemDownloadInstallScriptURL + "?$1",
	"^/kickstart.sh":   "/api" + GetSystemDownloadInstallScriptURL,
	"^/kickstart.sh?*": "/api" + GetSystemDownloadInstallScriptURL + "?$1",
}

// RewrittenFromRoot reports whether path is where a root alias lands. Those endpoints are
// published at the root of the domain and answered under /api, so what a caller reaches is
// the alias and not this path.
func RewrittenFromRoot(path string) bool {
	for _, target := range rootAliases {
		if target == path {
			return true
		}
	}

	return false
}
