#!/usr/bin/env bats

SCAN="$(cd "$(dirname "${BATS_TEST_FILENAME}")/../.." && pwd -P)/devscripts/comment-scan"

scan() {
    local f="$BATS_TEST_TMPDIR/x.go"
    cat > "$f"
    run "$SCAN" "$f"
}

scan_ts() {
    local f="$BATS_TEST_TMPDIR/x.tsx"
    cat > "$f"
    run "$SCAN" "$f"
}

scan_as() {
    local f="$BATS_TEST_TMPDIR/$1"
    mkdir -p "$(dirname "$f")"
    cat > "$f"
    run "$SCAN" "$f"
}

assert_kept() {
    [ "$status" -eq 0 ] || { echo "scanner failed: $output" >&2; return 1; }
    [ -z "$output" ] || { echo "expected nothing rejected, got:" >&2; echo "$output" >&2; return 1; }
}

assert_rejected() {
    [[ "$output" == *"$1"* ]] || {
        echo "expected '$1' to be rejected, got:" >&2
        echo "${output:-<nothing>}" >&2
        return 1
    }
}

@test "a doc comment on an exported declaration is kept" {
    scan <<'EOF'
package pkg

// Work runs the job and returns ErrBusy when another holds the lock.
func Work() error { return nil }
EOF
    assert_kept
}

@test "a doc comment on an unexported declaration is rejected" {
    scan <<'EOF'
package pkg

// work is the unexported half.
func work() error { return nil }
EOF
    assert_rejected "the unexported half"
}

@test "narration above a statement is rejected" {
    scan <<'EOF'
package pkg

func Work() {
	// fetch the user and check
	get()
}
EOF
    assert_rejected "fetch the user and check"
}

@test "a comment on an exported const group is kept" {
    scan <<'EOF'
package pkg

// The sort directions a query may ask for.
const (
	OrderAsc  = "asc"
	OrderDesc = "desc"
)
EOF
    assert_kept
}

@test "a doc comment on an exported member is kept" {
    scan <<'EOF'
package pkg

type Config struct {
	// Timeout bounds the whole dial, not each attempt.
	Timeout int
}
EOF
    assert_kept
}

@test "a block comment is read whole, not just its opener" {
    scan <<'EOF'
package utmp

/*	At session start a record is written to UtmpxFile.
	If a record with the same ID exists it is overwritten;
	otherwise it is appended.
*/

type Utmpx struct {
	Pid int32
}
EOF
    assert_kept
}

@test "a block comment inside a function is rejected in full" {
    scan <<'EOF'
package pkg

func Work() {
	/* the retry loop backs off
	   because the peer rate limits */
	retry()
}
EOF
    assert_rejected "the retry loop backs off"
    assert_rejected "because the peer rate limits"
}

@test "a licence header is kept" {
    scan <<'EOF'
// Copyright 2019 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package pkg
EOF
    assert_kept
}

@test "a build directive is kept" {
    scan <<'EOF'
//go:build docker

package pkg
EOF
    assert_kept
}

@test "a suppression with a reason is kept" {
    scan <<'EOF'
package pkg

func Work() {
	//nolint:errcheck // the caller closes conn on this path
	conn.Close()
}
EOF
    assert_kept
}

@test "the rows of a raw string literal are not comments" {
    scan <<'EOF'
package pkg

const banner = `
******************************
* Welcome to ShellHub        *
******************************
`
EOF
    assert_kept
}

@test "a doc comment detached by a blank line is still the declaration's" {
    scan <<'EOF'
package pkg

// Work runs the job.

func Work() error { return nil }
EOF
    assert_kept
}

@test "an eslint suppression is kept in TypeScript" {
    scan_ts <<'EOF'
export function useThing() {
  // eslint-disable-next-line react-hooks/exhaustive-deps
  useEffect(() => {}, []);
}
EOF
    assert_kept
}

@test "a JSDoc block on an export is kept whole" {
    scan_ts <<'EOF'
/**
 * Formats the device label, falling back to the hostname.
 */
export function label(device: Device): string {
  return device.name;
}
EOF
    assert_kept
}

@test "a local const is not exported, so its doc goes" {
    scan <<'EOF'
package pkg

func Work() {
	// MaxWait bounds one retry, not the whole loop.
	const MaxWait = time.Hour
}
EOF
    assert_rejected "MaxWait bounds one retry"

    scan_ts <<'EOF'
export function Row() {
  // build the label and render
  const label = make();
  return label;
}
EOF
    assert_rejected "build the label and render"
}

@test "an empty file and a file with no comments are both clean" {
    scan <<'EOF'
package pkg

func Work() error { return nil }
EOF
    assert_kept

    scan <<'EOF'
EOF
    assert_kept
}

@test "a lone asterisk in JSX is a required marker, not a JSDoc continuation" {
    scan_ts <<'EOF'
export function Field() {
  return (
    <label>
      Hostname
      <span aria-label="required" className="text-accent-red">
        *
      </span>
    </label>
  );
}
EOF
    assert_kept
}

@test "a JSDoc continuation is still rejected with the block that opens it" {
    scan_ts <<'EOF'
/**
 * Formats a duration.
 */
function format(ms: number) {
  return ms;
}
EOF
    assert_rejected "Formats a duration"
    assert_rejected "/**"
}

@test "a triple-slash reference is a compiler directive, not a comment" {
    scan_ts <<'EOF'
/// <reference types="vite/client" />

export const mode = import.meta.env.MODE;
EOF
    assert_kept
}

@test "a vitest environment pragma is a directive" {
    scan_ts <<'EOF'
// @vitest-environment node

import { readFileSync } from "node:fs";
EOF
    assert_kept
}

@test "a doc comment on a method of an exported class is what the linter asks for" {
    scan_ts <<'EOF'
export class LocalVaultBackend {
  /**
   * Reads the vault header, or null when no vault exists here.
   */
  loadMeta(): Promise<VaultMeta | null> {
    return Promise.resolve(null);
  }
}
EOF
    assert_kept
}

@test "a doc comment on a declaration the export list names is kept" {
    scan_ts <<'EOF'
/**
 * Edits an API key's name and role.
 */
function EditKeyDrawer() {
  return null;
}

export default EditKeyDrawer;
EOF
    assert_kept
}

@test "a doc comment on a declaration nothing exports is still rejected" {
    scan_ts <<'EOF'
/**
 * Builds the label.
 */
function buildLabel() {
  return "";
}
EOF
    assert_rejected "Builds the label"
}

@test "narration inside an exported class is still rejected" {
    scan_ts <<'EOF'
export class Recorder {
  start(): void {
    // reset the clock before writing the header
    this.startMs = 0;
  }
}
EOF
    assert_rejected "reset the clock"
}

@test "narration inside a method of an exported class is rejected" {
    scan_ts <<'EOF'
export class Recorder {
  finish(): void {
    // write the trailer before closing so a truncated file still replays
    this.write("");
  }
}
EOF
    assert_rejected "write the trailer"
}

@test "a JSX comment in a component is rejected" {
    scan_ts <<'EOF'
export function Header() {
  return (
    <div>
      {/* the title sits left of the actions */}
      <h1>Devices</h1>
    </div>
  );
}
EOF
    assert_rejected "the title sits left"
}

@test "a comment line in a workflow is rejected" {
    scan_as ci.yml <<'EOF'
jobs:
  test:
    # run the suite on every push
    runs-on: ubuntu-24.04
EOF
    assert_rejected "run the suite on every push"
}

@test "the version after a pinned action is kept" {
    scan_as ci.yml <<'EOF'
steps:
  - uses: actions/checkout@3d3c42e5aac5ba805825da76410c181273ba90b1 # v7
EOF
    assert_kept
}

@test "a comment inside a run script is rejected" {
    scan_as ci.yml <<'EOF'
steps:
  - name: Build
    run: |
      # build twice so the cache warms
      make build
EOF
    assert_rejected "build twice so the cache warms"
}

@test "a markdown heading inside a text block scalar is kept" {
    scan_as template.yml <<'EOF'
body:
  - type: markdown
    attributes:
      value: |
        # Bug report
        Describe what happened.
EOF
    assert_kept
}

@test "a heredoc body inside a run script is kept" {
    scan_as ci.yml <<'EOF'
steps:
  - run: |
      cat > summary.md <<'MD'
      # Results
      MD
      echo done
EOF
    assert_kept
}

@test "a comment after a block scalar ends is rejected" {
    scan_as compose.yml <<'EOF'
services:
  api:
    description: >
      # not a comment
# the gateway fronts every service
  gateway:
    image: caddy
EOF
    assert_rejected "the gateway fronts every service"
    [[ "$output" != *"not a comment"* ]]
}

@test "yaml tool directives are kept" {
    scan_as ci.yml <<'EOF'
# yaml-language-server: $schema=https://json.schemastore.org/github-workflow.json
on:
  pull_request_target: # zizmor: ignore[dangerous-triggers]
EOF
    assert_kept
}

@test "a shell script keeps its shebang and suppressions and loses its narration" {
    scan_as run.sh <<'EOF'
#!/usr/bin/env bash
# shellcheck disable=SC2086
set -eu
# wait for the api before seeding
seed
EOF
    assert_rejected "wait for the api before seeding"
    [[ "$output" != *"shellcheck"* && "$output" != *"#!/"* ]]
}

@test "a heredoc body in a shell script is kept" {
    scan_as run.sh <<'EOF'
cat > notes.md <<-NOTES
	# Release notes
	NOTES
EOF
    assert_kept
}

@test "a herestring does not open a heredoc" {
    scan_as run.sh <<'EOF'
read -r a b <<< "$pair"
# split the pair
echo "$a"
EOF
    assert_rejected "split the pair"
}

@test "a shell script without an extension is read by its shebang" {
    scan_as bin/tool <<'EOF'
#!/bin/sh
# forward everything to compose
exec docker compose "$@"
EOF
    assert_rejected "forward everything to compose"
}

@test "a file without an extension or a shell shebang is not scanned" {
    scan_as bin/data <<'EOF'
# not a script
EOF
    assert_kept
}

@test "a licence header in a shell script is kept" {
    scan_as run.sh <<'EOF'
#!/bin/sh
# Copyright 2020 O.S. Systems
# SPDX-License-Identifier: Apache-2.0
echo ok
EOF
    assert_kept
}

@test "a Dockerfile keeps its parser directive and RUN heredoc and loses its narration" {
    scan_as Dockerfile <<'EOF'
# syntax=docker/dockerfile:1
FROM alpine
# install tools first so the layer caches
RUN <<EOT
# not a Dockerfile comment
apk add curl
EOT
EOF
    assert_rejected "install tools first"
    [[ "$output" != *"syntax="* && "$output" != *"not a Dockerfile comment"* ]]
}

@test "a Makefile comment is rejected" {
    scan_as Makefile <<'EOF'
# build everything
all:
	go build ./...
EOF
    assert_rejected "build everything"
}

@test "a CSS comment is rejected and a stylelint suppression kept" {
    scan_as base.css <<'EOF'
/* stylelint-disable-next-line selector-class-pattern */
.Legacy { color: red; }
/*
 * the shadow separates the sidebar
 */
.seam { box-shadow: none; }
EOF
    assert_rejected "the shadow separates the sidebar"
    [[ "$output" != *"stylelint"* ]]
}

@test "an Astro markup comment is rejected, every line of it" {
    scan_as Page.astro <<'EOF'
---
/** Props the page takes from its route. */
export interface Props { title: string }
---
<!--
  keep the hero above the fold
-->
<h1>{Astro.props.title}</h1>
EOF
    assert_rejected "keep the hero above the fold"
    assert_rejected "<!--"
    [[ "$output" != *"Props the page takes"* ]]
}

@test "a comment carrying a template action is kept" {
    scan_as install.sh <<'EOF'
#!/bin/sh

# Overridden variables from Go template: {{.Overrides}}

echo ok
EOF
    assert_kept
}

@test "the doc comment on an Astro component's Props is kept" {
    scan_as Link.astro <<'EOF'
---
/**
 * The one control a page sends its reader to.
 */
interface Props {
  href: string;
}
---
<a href={Astro.props.href}><slot /></a>
EOF
    assert_kept
}

@test "--language names the language a file is read as" {
    run "$SCAN" --language .github/workflows/qa.yml
    [ "$output" = yaml ]
    run "$SCAN" --language docs/README.md
    [ -z "$output" ]
}

repo() {
    REPO="$BATS_TEST_TMPDIR/repo"
    git init -q -b master "$REPO"
    git -C "$REPO" config user.email bats@example.com
    git -C "$REPO" config user.name bats
    cd "$REPO"
}

commit() {
    git add -A
    git commit -q -m "$1"
}

@test "--diff reports a comment the branch adds and fails" {
    repo
    printf 'jobs:\n  test:\n    runs-on: ubuntu-24.04\n' > ci.yml
    commit base
    git checkout -q -b feature
    printf 'jobs:\n  # explain the job\n  test:\n    runs-on: ubuntu-24.04\n' > ci.yml
    commit feature
    run "$SCAN" --diff master
    [ "$status" -eq 1 ]
    [[ "$output" == *"ci.yml:2: # explain the job"* ]]
}

@test "--diff leaves a comment the base already had" {
    repo
    printf '#!/bin/sh\n# old narration\necho a\n' > run.sh
    commit base
    git checkout -q -b feature
    printf '#!/bin/sh\n# old narration\necho a\necho b\n' > run.sh
    commit feature
    run "$SCAN" --diff master
    [ "$status" -eq 0 ]
    [ -z "$output" ]
}

@test "--diff checks uncommitted and untracked files" {
    repo
    printf 'package pkg\n' > x.go
    commit base
    printf 'package pkg\n\nfunc f() {\n\t// narrate the call\n\tg()\n}\n' > x.go
    printf '.a {\n  /* explain the rule */\n  color: red;\n}\n' > new.css
    run "$SCAN" --diff master
    [ "$status" -eq 1 ]
    [[ "$output" == *"x.go:4:"*"narrate the call"* ]]
    [[ "$output" == *"new.css:2:"*"explain the rule"* ]]
}

@test "--diff annotates each finding under GitHub Actions" {
    repo
    printf 'all:\n\ttrue\n' > Makefile
    commit base
    printf '# build everything\nall:\n\ttrue\n' > Makefile
    GITHUB_ACTIONS=true run "$SCAN" --diff master
    [ "$status" -eq 1 ]
    [[ "$output" == *"::error file=Makefile,line=1::Comment not allowed: # build everything"* ]]
}

@test "--diff without a merge base explains what to fetch" {
    repo
    printf 'a: 1\n' > x.yml
    commit base
    run "$SCAN" --diff origin/master
    [ "$status" -eq 2 ]
    [[ "$output" == *"no merge base with origin/master"* ]]
}

@test "--help prints the usage and no argument is an error" {
    run "$SCAN" --help
    [ "$status" -eq 0 ]
    [[ "$output" == *"Usage: devscripts/comment-scan --diff"* ]]
    run "$SCAN"
    [ "$status" -eq 2 ]
}
