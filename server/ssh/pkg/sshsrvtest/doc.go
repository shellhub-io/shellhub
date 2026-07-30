// Package sshsrvtest provides utilities for setting up testable SSH servers.
// The `Conn` struct represents the test SSH connection and can be created using
// the `New` function, which binds a random port to simulate the client.
//
// To start the server and the fictional connection, use the `Conn.Start()`.
// The server, listener, and related resources can be destroyed using `Conn.Teardown()`.
//
// It is important to note that if any error occurs within the package, the program will panic.
package sshsrvtest
