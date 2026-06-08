// Package govulncheck provides a wrapper around the govulncheck tool that
// parses its JSON output, applies an allowlist to suppress known vulnerabilities,
// logs each suppression to GITHUB_STEP_SUMMARY, and fails on stale allowlist entries.
package govulncheck

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
)

// Reachability describes how a vulnerable symbol is reached.
type Reachability string

const (
	// ReachabilityCalled means the vulnerable function is in the call graph.
	ReachabilityCalled Reachability = "called"
	// ReachabilityImported means the vulnerable package is imported but the
	// vulnerable function itself is not called.
	ReachabilityImported Reachability = "imported"
)

// Finding is a single vulnerability finding emitted by govulncheck.
type Finding struct {
	// OSVID is the Go vulnerability database identifier, e.g. "GO-2024-0001".
	OSVID string
	// Called is true when the vulnerable function appears in the call graph
	// (i.e. the trace contains at least one frame with a function name).
	// False means the vulnerable package is imported but the vulnerable
	// function is not reachable.
	Called bool
}

// AllowEntry is one line from .govulncheck-allow.txt.
type AllowEntry struct {
	// OSVID is the Go vulnerability database identifier.
	OSVID string
	// Reachability is the kind of finding this entry suppresses.
	Reachability Reachability
}

// govulncheckMessage is the top-level JSON object in govulncheck -json output.
// Only the fields relevant to this wrapper are decoded.
type govulncheckMessage struct {
	Finding *govulncheckFinding `json:"finding"`
}

// govulncheckFinding is the "finding" field of a govulncheck JSON message.
type govulncheckFinding struct {
	OSV   string             `json:"osv"`
	Trace []govulncheckFrame `json:"trace"`
}

// govulncheckFrame is one entry in the trace.
// When Function is non-empty the vulnerable symbol is in the call graph.
type govulncheckFrame struct {
	Function string `json:"function"`
}

// ParseFindings reads the newline-delimited JSON produced by
// `govulncheck -json` and returns all vulnerability findings.
func ParseFindings(r io.Reader) ([]Finding, error) {
	var findings []Finding

	scanner := bufio.NewScanner(r)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		var msg govulncheckMessage

		if err := json.Unmarshal([]byte(line), &msg); err != nil {
			return nil, fmt.Errorf("govulncheck: failed to parse JSON line: %w", err)
		}

		if msg.Finding == nil {
			continue
		}

		f := Finding{
			OSVID:  msg.Finding.OSV,
			Called: isCalled(msg.Finding.Trace),
		}

		findings = append(findings, f)
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("govulncheck: error reading input: %w", err)
	}

	return findings, nil
}

// isCalled returns true when any frame in trace contains a non-empty Function.
func isCalled(trace []govulncheckFrame) bool {
	for _, frame := range trace {
		if frame.Function != "" {
			return true
		}
	}

	return false
}

// LoadAllowList reads an allowlist file in which each non-comment, non-blank
// line has one of the following formats:
//
//	<OSV-ID> <reachability>
//	<OSV-ID> reachability=<reachability> # <owner> — <justification>
//
// where <reachability> is either "called" or "imported".
// Lines beginning with '#' and inline comments (text after ' #') are ignored.
func LoadAllowList(path string) ([]AllowEntry, error) {
	f, err := os.Open(path) //nolint:gosec
	if err != nil {
		return nil, fmt.Errorf("govulncheck: failed to open allow list %q: %w", path, err)
	}

	defer f.Close() //nolint:errcheck

	var entries []AllowEntry

	scanner := bufio.NewScanner(f)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Strip inline comments: anything from " #" onward.
		if idx := strings.Index(line, " #"); idx >= 0 {
			line = strings.TrimSpace(line[:idx])
		}

		if line == "" {
			continue
		}

		parts := strings.Fields(line)
		if len(parts) < 2 {
			return nil, fmt.Errorf("govulncheck: invalid allow list line %q: expected \"<OSV-ID> <reachability>\"", line)
		}

		// Support both "called" and "reachability=called" forms.
		rawReachability := strings.TrimPrefix(parts[1], "reachability=")

		r := Reachability(rawReachability)
		if r != ReachabilityCalled && r != ReachabilityImported {
			return nil, fmt.Errorf("govulncheck: unknown reachability %q in allow list", rawReachability)
		}

		entries = append(entries, AllowEntry{
			OSVID:        parts[0],
			Reachability: r,
		})
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("govulncheck: error reading allow list: %w", err)
	}

	return entries, nil
}

// CheckFindings applies the allowlist to the findings:
//   - A finding is suppressed when the allowlist contains an entry whose OSV ID
//     and reachability both match.
//   - Each suppression is written to summaryWriter (for GITHUB_STEP_SUMMARY).
//   - A stale allowlist entry (one that matches no finding) causes an error.
//   - Any unsuppressed finding causes an error.
//
// summaryWriter may be nil; in that case no summary output is written.
func CheckFindings(findings []Finding, allow []AllowEntry, summaryWriter io.Writer) error {
	// usedEntries tracks which allow entries actually matched a finding.
	usedEntries := make(map[int]bool)

	var unsuppressed []Finding

	for _, finding := range findings {
		suppressed := false

		for idx, entry := range allow {
			if entry.OSVID != finding.OSVID {
				continue
			}

			if entry.Reachability != reachabilityOf(finding) {
				continue
			}

			suppressed = true
			usedEntries[idx] = true

			writeSuppression(summaryWriter, finding, entry)

			break
		}

		if !suppressed {
			unsuppressed = append(unsuppressed, finding)
		}
	}

	var errs []string

	if len(unsuppressed) > 0 {
		for _, f := range unsuppressed {
			errs = append(errs, fmt.Sprintf("unsuppressed vulnerability: %s (%s)", f.OSVID, reachabilityOf(f)))
		}
	}

	for idx, entry := range allow {
		if !usedEntries[idx] {
			errs = append(errs, fmt.Sprintf("stale allow list entry: %s %s (no matching finding)", entry.OSVID, entry.Reachability))
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("govulncheck: %s", strings.Join(errs, "; "))
	}

	return nil
}

// reachabilityOf converts a Finding's Called flag into a Reachability value.
func reachabilityOf(f Finding) Reachability {
	if f.Called {
		return ReachabilityCalled
	}

	return ReachabilityImported
}

// writeSuppression appends a suppression note to w.
func writeSuppression(w io.Writer, f Finding, entry AllowEntry) {
	if w == nil {
		return
	}

	fmt.Fprintf(w, "suppressed: %s (%s) — matched allow entry %s %s\n",
		f.OSVID, reachabilityOf(f), entry.OSVID, entry.Reachability)
}

// Run is the CLI entry-point: it reads govulncheck JSON from r, loads the
// allowlist from allowListPath, and reports results to summaryWriter.
//
// summaryWriter defaults to the file pointed at by GITHUB_STEP_SUMMARY when nil.
func Run(r io.Reader, allowListPath string, summaryWriter io.Writer) error {
	if summaryWriter == nil {
		if path := os.Getenv("GITHUB_STEP_SUMMARY"); path != "" {
			f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644) //nolint:gosec
			if err == nil {
				defer f.Close() //nolint:errcheck

				summaryWriter = f
			}
		}

		if summaryWriter == nil {
			summaryWriter = io.Discard
		}
	}

	findings, err := ParseFindings(r)
	if err != nil {
		return err
	}

	allow, err := LoadAllowList(allowListPath)
	if err != nil {
		return err
	}

	return CheckFindings(findings, allow, summaryWriter)
}
