// Package auth provides authentication handlers for client connections.
//
// This package includes two authentication methods: [PasswordHandlerWithTunnel] and [PublicKeyHandlerWithTunnel].
// [PasswordHandlerWithTunnel] is the second authentication method tried by the server to connect the client to the agent,
// while [PublicKeyHandlerWithTunnel] is the first authentication method attempted.
package auth
