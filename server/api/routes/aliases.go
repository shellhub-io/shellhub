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
