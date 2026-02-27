package migrate

import "github.com/google/uuid" //nolint:depguard // UUID v5 (deterministic) requires google/uuid directly

// migrationNamespace is a fixed UUID namespace used for deterministic ObjectID-to-UUID conversion.
// This ensures the same MongoDB ObjectID always produces the same UUID across migration runs.
var migrationNamespace = uuid.MustParse("a1b2c3d4-e5f6-4a7b-8c9d-0e1f2a3b4c5d")

// ObjectIDToUUID converts a MongoDB ObjectID hex string to a deterministic UUID v5.
func ObjectIDToUUID(hexID string) string {
	return uuid.NewSHA1(migrationNamespace, []byte(hexID)).String()
}
