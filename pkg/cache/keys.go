package cache

// SystemKey caches the instance's singleton system row.
//
// It is deliberately shared across editions and layers: the Community API service reads it,
// the Cloud/Enterprise service reads it, and the cloud store deletes it when an administrator
// reconfigures authentication. Those live in three packages and two repositories, so the name
// has to be defined once — a second literal spelling of it silently stops invalidating.
const SystemKey = "system"
