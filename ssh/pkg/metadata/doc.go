// Package metadata provides a secure way to store and retrieve data based on a session context,
// preventing race conditions. Each session maintains its own context and won't be overridden by
// another session.
//
// Functions prefixed with `Store` are used to save data, while those prefixed with `MaybeStore`
// are conditional stores. You can also use `Restore` to retrieve stored data.
//
// Additionally, metadata offers a mock implementation for testing the app without requiring a
// real session context.
package metadata
