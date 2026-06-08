// Command govulncheck-wrapper wraps govulncheck JSON output, applies the
// project allowlist, writes suppression notes to GITHUB_STEP_SUMMARY, and
// exits non-zero on unsuppressed findings or stale allowlist entries.
//
// Usage:
//
//	govulncheck -json ./... | govulncheck-wrapper <allowlist-path>
package main

import (
	"fmt"
	"os"

	"github.com/shellhub-io/shellhub/.github/govulncheck"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintf(os.Stderr, "usage: govulncheck-wrapper <allowlist-path>\n")
		os.Exit(2)
	}

	if err := govulncheck.Run(os.Stdin, os.Args[1], nil); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}
}
