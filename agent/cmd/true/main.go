// Package main is a minimal static binary that exits with code 0.
// It is compiled into /bin/true in the agent image and used as a
// build-time smoke-check to verify the scratch-based image can execute
// a statically linked Go binary before the real agent binary is copied.
package main

func main() {}
