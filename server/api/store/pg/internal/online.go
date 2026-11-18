package internal

// DeviceOnline is the condition that the device row aliased "device" is online at the cutoff
// bound to its one placeholder: not disconnected, and last seen after the cutoff.
const DeviceOnline = `("device"."disconnected_at" IS NULL AND "device"."last_seen" > ?)`
