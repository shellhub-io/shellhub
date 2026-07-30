package http

// reasonAgentDeviceResolve justifies resolving a device across every namespace on the agent-facing
// connection routes. An agent authenticates as a device, and agents older than 0.15 send no tenant
// header at all, so the namespace is read back off the device that is resolved.
const reasonAgentDeviceResolve = "an agent identifies itself by device UID; agents before 0.15 send no tenant, so the namespace is read back off the resolved device"
