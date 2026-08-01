// Package auth provides authentication handlers for client connections.
//
// Public key authentication runs in two phases, because the SSH protocol lets a
// client ask whether a key would be accepted without signing anything.
// [PublicKeyOffer] answers that question and only reads; [PublicKeyVerified]
// runs once the signature checks out and is where the browser approval, the
// session registration and the connection to the agent happen.
//
// [PasswordHandler] is the second authentication method tried by the server to
// connect the client to the agent. A password has no query form, so it
// authenticates in one step.
package auth
