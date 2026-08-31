//go:build docker

package command

import "testing"

var _ func() error = CheckCredentialSwitch

// TestCheckCredentialSwitchCompiles is a named test that anchors the
// compile-time assertion above so it shows up in the test run output.
// The assertion is evaluated at compile time; the test body itself is a
// simple pass-through that confirms the symbol is reachable at runtime too.
func TestCheckCredentialSwitchCompiles(t *testing.T) {
	if err := CheckCredentialSwitch(); err != nil {
		t.Errorf("CheckCredentialSwitch() returned unexpected error under -tags docker: %v", err)
	}
}
