// Package enterprise provides the extension point for cloud/enterprise features.
// In Community Edition builds (without the "cloud" build tag), this package is
// empty — it compiles but registers nothing.
//
// In Enterprise/Cloud builds (-tags enterprise), imports.go blank-imports the cloud
// packages, triggering their init() functions which register billing providers,
// route extensions, and worker extensions before the server starts.
package enterprise
