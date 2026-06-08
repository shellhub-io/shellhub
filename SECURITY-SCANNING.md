# Security Scanning

This document describes the security scanning infrastructure for ShellHub.

**Quick reference**

| Topic | Section |
|---|---|
| Frozen required-check name | [Required check — `security-gate`](#required-check--security-gate) |
| Full list of scan tools | [Scan inventory](#scan-inventory) |
| Suppression files and how to add entries | [Baseline suppression files](#baseline-suppression-files) |
| Inline `#nosec` and `# nosemgrep` conventions | [Inline suppression conventions](#inline-suppression-conventions) |
| `gosec` exclusion review and rationale | [`gosec` exclusion review](#gosec-exclusion-review) |
| amd64-only scan residual risk | [Agent amd64-scan residual risk](#agent-amd64-scan-residual-risk) |
| Fork and Dependabot behavior | [Fork and Dependabot PRs](#fork-and-dependabot-prs) |
| How to flip `security-gate` to blocking | [Rollout order](#rollout-order) |

---

## Required check — `security-gate`

**`security-gate` is the single frozen required-check name** registered in branch protection.
Every security workflow (Security, Semgrep, CodeQL) exposes a terminal job named exactly
`security-gate` that aggregates the results of its scan jobs.  A PR is not mergeable until
all three `security-gate` jobs report success (or skipped).

> **Do not rename `security-gate`.**  Renaming it silently removes the branch-protection
> requirement and allows unreviewed code to merge.

---

## Scan inventory

| Workflow file | Tool | What it scans | Runs on |
|---|---|---|---|
| `.github/workflows/security.yml` | **govulncheck** | Known Go CVEs in all modules (`.`, `api`, `agent`, `ssh`, `cli`, `gateway`, `openapi`, `tests`) | PR + push to master + weekly |
| `.github/workflows/security.yml` | **Trivy (image)** | OS/library CVEs in service images (`api`, `ssh`, `gateway`, `cli`, `ui`, `agent`) | PR + push to master + weekly |
| `.github/workflows/semgrep.yml` | **Semgrep** | Static analysis via `p/golang`, `p/dockerfile`, `p/ci`; PR mode uses `--baseline-commit` so only _new_ findings block | PR + push to master + weekly |
| `.github/workflows/codeql.yml` | **CodeQL** | Semantic Go analysis across all modules; SARIF uploaded to GitHub Security tab | PR + push to master + weekly |
| `.github/workflows/build-agent.yml` | **Trivy (image, amd64)** | Agent image CVEs as part of the agent build pipeline | Push to master + tags |

All workflows upload SARIF reports to the GitHub Security tab (skipped on fork PRs because
those lack the necessary `security-events: write` permission).

---

## Baseline suppression files

Three baseline files gate new suppressions behind a team-lead code-owner review
(see `.github/CODEOWNERS`).

### `.trivyignore` — Trivy CVE suppression

```
# CVE-YYYY-NNNNN  # owner: @handle — reason; tracked: ISSUE-123
```

**How to add an entry:**

1. Verify the CVE is genuinely not exploitable in this context (e.g., the affected
   component is not reachable, or a mitigating control already exists).
2. Add a line in the format above with your GitHub handle as `owner` and a brief
   justification.
3. Open a PR — `.github/CODEOWNERS` routes `.trivyignore` changes to
   `@shellhub-io/team-lead` for mandatory review.
4. Schedule a quarterly review to remove the entry once the vulnerability is patched.

### `.govulncheck-allow.txt` — govulncheck allowlist

```
# GO-YYYY-NNNN reachability=called|imported # owner: @handle — justification; tracked: ISSUE-123
```

Fields:

- `GO-YYYY-NNNN` — Go vulnerability database ID (required).
- `reachability=called` — the vulnerable function appears in the call graph.
- `reachability=imported` — the vulnerable package is imported but the function is not called.

**How to add an entry:** follow the same owner + justification + quarterly-review process as
`.trivyignore`.  Stale entries (where the vulnerability no longer appears in findings) cause
CI to fail — remove them promptly.

### `.semgrepignore` — Semgrep path exclusion

```
# <glob pattern>  # reason
```

Use `.semgrepignore` to exclude paths that produce only false positives (test fixtures,
generated mocks).  **Do not suppress production source paths** — fix the finding instead.

---

## Inline suppression conventions

### `#nosec` (gosec / govulncheck)

Append `//nolint:gosec` or `#nosec GXXX` directly on the line that triggers the finding:

```go
conn, err := tls.Dial("tcp", addr, cfg) //nolint:gosec // G402: minimum TLS version enforced by cfg
```

Always include a short rationale after the directive.

### `# nosemgrep: <rule-id> -- <reason>` (Semgrep)

Add a comment on the line **above** the offending code using the exact format
`# nosemgrep: <rule-id> -- <reason>`:

```python
# nosemgrep: go.lang.security.audit.dangerous-exec-command -- input validated by allowlist
exec.Command(cmd, args...)
```

For Go files the comment is an inline `//`:

```go
// nosemgrep: go.lang.security.audit.crypto.use_of_weak_crypto -- MD5 used only for non-security cache key
hash := md5.Sum(data) //nolint:gosec
```

The rule ID must match the Semgrep rule identifier exactly (visible in the Semgrep output
or the SARIF report).

---

## `gosec` exclusion review

Before this work, `.golangci.yaml` blanket-excluded four gosec rules: G104, G301, G302, and G304. The review reduced the exclusion list to G104 only, which is redundant with the project-wide `errcheck` disable. The remaining rules were handled per-site:

| Rule | What it flags | Sites found | Resolution |
|---|---|---|---|
| G301 | Directory created with permissions > 0750 | `gateway/nginx.go` (`MkdirAll 0o755`) | Tightened to `0o750`; annotated with `//nolint:gosec` because gosec still flags the call form |
| G302 | File created with permissions > 0600 | `agent/server/utmp/utmp.go` (`OpenFile 0o644`) | Kept at `0o644` — utmp/wtmp files are conventionally world-readable per POSIX; tighter permissions break system accounting tools. Annotated |
| G304 | File path from variable (potential traversal) | `agent/pkg/keygen/keygen.go` (`Create`, `ReadFile`), `agent/pkg/sysinfo/network.go` (`ReadFile`), `api/store/pg/` tests (`Open`, `ReadFile`) | keygen: path comes from a configured env var, not user input — annotated. sysinfo: path rooted under `/sys/class/net` (read-only kernel VFS) — annotated. agent: added `cleanKeyPath` guard rejecting raw `..` sequences before `generatePrivateKey`/`readPublicKey`. Tests: paths from `runtime.Caller`, not user input — annotated |

The gateway module received its own `.golangci.yaml` (separate Go module, golangci-lint v2 does not inherit parent configs) with the same G104-only exclusion.

---

## Agent amd64-scan residual risk

The agent image is a multi-arch build (`linux/amd64`, `linux/arm64/v8`, `linux/arm/v7`,
`linux/arm/v6`, `linux/386`).  Trivy scans only the **amd64** variant during CI
(`.github/workflows/build-agent.yml`).

**Residual risk:** architecture-specific CVEs for `arm` and `386` variants are **not
blocked by CI**.  They are surfaced by the weekly scheduled scan
(`.github/workflows/security.yml` `trivy-image` matrix includes `agent`) and tracked
manually.  If a CVE affects a non-amd64 variant only, add it to `.trivyignore` with an
explicit `owner` and architectural scope note.

---

## Fork and Dependabot PRs

- **Fork PRs** lack the `security-events: write` permission required to upload SARIF reports
  to the GitHub Security tab.  The SARIF upload steps are conditional on
  `!github.event.pull_request.head.repo.fork` and are skipped automatically.  Scan jobs
  themselves still run and `security-gate` still blocks — only the SARIF upload is skipped.

- **Dependabot PRs** are treated as first-party PRs (Dependabot has write access to the
  repo).  All scans run normally.  Because Dependabot bumps one dependency at a time the
  blast radius of a failing scan is narrow and easy to triage.

---

## Rollout order

Follow this sequence when enabling `security-gate` as a blocking required check on a new
branch or workflow:

1. **Land scans as non-blocking** — merge the workflow files with `security-gate` jobs but
   do not yet mark them as required checks in branch protection.  Let the scans accumulate
   data for at least one week.

2. **Seed the baseline files** — review findings from step 1 and populate `.trivyignore`,
   `.govulncheck-allow.txt`, and `.semgrepignore` with justified suppressions for any
   pre-existing issues that cannot be fixed immediately.  Each entry requires an owner and
   justification comment.

3. **Flip to blocking-on-new** — configure Semgrep to run with `--baseline-commit origin/master`
   on PRs (already the default in `semgrep.yml`) so that only _new_ findings introduced by
   a PR cause `security-gate` to fail.  Confirm that existing suppressed findings do not
   re-surface.

4. **Admin marks `security-gate` required + Require review from Code Owners** — in the
   GitHub repository settings → Branches → Branch protection rules for `master`:
   - Under "Require status checks to pass before merging", add **`security-gate`** (the
     exact frozen name) from each of the three workflows (Security, Semgrep, CodeQL).
   - Enable **"Require review from Code Owners"** so that changes to `.trivyignore`,
     `.govulncheck-allow.txt`, and `.semgrepignore` always go through `@shellhub-io/team-lead`.

> **Note:** Steps 1–3 are performed by contributors; step 4 requires a repository admin.
