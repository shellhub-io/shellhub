# Compose wrapper tests

bats tests for `bin/docker-compose`. They verify the wrapper's decisions —
which overlays it chains into `COMPOSE_FILE`, which env files it loads into
`COMPOSE_ENV_FILES`, and that the cloud-without-enterprise guard aborts.

The tests do not start containers and do not invoke Compose. They stub
`docker` in `PATH` with a script that echoes the wrapper's exported env, then
assert on that output.

## Layout

```
tests/compose/
  helpers.bash          # capture_with() + require_cloud() + docker stub setup
  wrapper.bats          # scenarios that don't need cloud/
  with_cloud.bats       # scenarios that do; skip per test when ../cloud/ is absent
```

Each scenario is declared inline via `capture_with VAR1=value VAR2=value ...`,
which writes a tmpfile, points the wrapper at it via `ENV_OVERRIDE`, and
removes it after. No persistent fixtures, no golden files.

## Running locally

Install bats-core: <https://github.com/bats-core/bats-core>

Then from the repo root:

```sh
bats tests/compose/                # full suite; with_cloud skips if cloud/ is absent
bats tests/compose/wrapper.bats    # subset that doesn't need cloud/
```

## CI

`compose-tests.yml` runs `wrapper.bats` on every PR that touches the wrapper,
the env files, or the compose YAMLs. The cloud-dependent scenarios are
covered by the cloud repo's CI workflow (it already clones both repos side
by side); wiring is a separate change.
