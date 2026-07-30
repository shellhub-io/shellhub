package session

// reasonSSHIDDeviceResolve justifies resolving a device across every namespace when a connection
// addressed it by UID rather than by SSHID. The namespace is the answer, not the question:
// everything after this lookup runs bounded to the tenant it returns.
const reasonSSHIDDeviceResolve = "an SSH client addressed a device by UID alone; the namespace is what this lookup produces, and the session runs bounded to it"
