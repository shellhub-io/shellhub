// Package challenge names the keyboard-interactive challenges the SSH gateway
// sends, so the gateway and the web terminal bridge match on one spelling
// instead of each holding its own copy of the string.
package challenge

const (
	// Approval asks the person to approve the login in the console. The bridge
	// matches this name to know a challenge carries an approval code rather
	// than prose meant for a terminal.
	Approval = "shellhub-approval"

	// Denied carries the reason a login was refused. It is deliberately not
	// Approval: the bridge must render a refusal as an error rather than open
	// an approval dialog that can no longer be answered.
	Denied = "shellhub-approval-denied"
)
