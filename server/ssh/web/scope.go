package web

// reasonWebHandoffDeviceResolve justifies resolving a device across every namespace during the
// web terminal's credential handoff. The credential carries a device UID, and the tenant it belongs
// to is what the public key lookup that follows is bounded by.
const reasonWebHandoffDeviceResolve = "a web terminal credential names a device UID; its tenant is what the public key lookup that follows is bounded by"
