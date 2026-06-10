//go:build docker
// +build docker

package command

import "testing"

// Compile-time assertion: CheckCredentialSwitch is exported and has the expected
// signature (func() error) under the docker build tag.  This mirrors the
// identical assertion in command_native_test.go for the !docker tag so that
// neither tag set can silently break the build.
var _ func() error = CheckCredentialSwitch

// TestCheckCredentialSwitchCompiles is a named test that anchors the
// compile-time assertion above so it shows up in the test run output.
// The assertion is evaluated at compile time; the test body itself is a
// simple pass-through that confirms the symbol is reachable at runtime too.
func TestCheckCredentialSwitchCompiles(t *testing.T) {
	// Calling the function ensures the linker includes the symbol and that
	// it actually returns nil in docker mode (documented no-op).
	if err := CheckCredentialSwitch(); err != nil {
		t.Errorf("CheckCredentialSwitch() returned unexpected error under -tags docker: %v", err)
	}
}
