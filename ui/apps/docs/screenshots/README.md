# Docs screenshot toolkit

Reusable automation that seeds demo data and re-captures the dashboard
screenshots used throughout the documentation. Re-run it whenever the UI
changes instead of hand-rebuilding shots.

It drives a headless Chromium (via `playwright-core`) against a local ShellHub
dev environment, browsing through the **branded host** `shellhub.example.com` so
the UI rebrands itself automatically (SSHIDs, agent install command, etc.) with
no backend change.

> **No `/etc/hosts` entry needed.** The browser reaches the branded host because
> Chromium maps it to `127.0.0.1` via `--host-resolver-rules` — that mapping is
> internal to Chromium and does not affect Node. Node-side code (seeding, dynamic
> route token lookups) uses its own `fetch`, which ignores that mapping, so it
> talks to the REST API on `localhost` (`SHOTS_API_URL`, default
> `http://localhost`) instead. The REST API does not affect branding, so this
> split is safe: the browser stays branded while Node stays resolvable.

> ## ⚠️ Never commit your license
>
> The enterprise license is **private**. Never commit a license file, never put
> a real license path into any tracked file. Keep its path in your gitignored
> `shellhub/.env.override` (see `env.override.example`) and keep the license
> itself outside the repo.

---

## Layout

```
screenshots/
├── config.mjs        # central config from env vars (+ required secrets)
├── lib.mjs           # shared helpers (browser, api, overlay, key/MAC gen)
├── auth.mjs          # ensureAuth(): login -> .state.json (gitignored)
├── seed.mjs          # idempotent data seeding (npm run shots:seed)
├── manifest.mjs      # declarative list of shots
├── capture.mjs       # main runner (npm run shots:capture)
├── env.override.example   # enterprise dev settings for .env.override (TEMPLATE)
├── README.md
└── .gitignore        # .state.json, node_modules
```

---

## Prerequisites

- Node 20+ (uses global `fetch` and top-level `await`).
- `playwright-core` installed (declared as a devDependency of
  `@shellhub/docs`). Run `npm install` in `ui/apps/docs/`.
- A Chromium/Chrome binary. Point `SHOTS_CHROMIUM` at it (e.g. system Chromium,
  or `npx playwright install chromium` and use the printed path).
- A running ShellHub dev environment (`./bin/docker-compose up`).
- `ssh-keygen` (soft dependency, for the public-key seed). If missing, that one
  seed step is skipped with a warning.

---

## Environment variables

| Variable            | Required | Default                                    | Purpose                                              |
| ------------------- | -------- | ------------------------------------------ | ---------------------------------------------------- |
| `SHOTS_CHROMIUM`    | **yes**  | —                                          | Path to the Chromium/Chrome executable.              |
| `SHOTS_PASSWORD`    | **yes**  | —                                          | Password of the login user (no insecure default).    |
| `SHOTS_USERNAME`    | no       | `dev`                                      | Login user.                                          |
| `SHOTS_DOMAIN`      | no       | `shellhub.example.com`                     | Branded host (mapped to 127.0.0.1).                  |
| `SHOTS_BASE_URL`    | no       | `http://shellhub.example.com`              | UI base URL (browser navigates this branded host).   |
| `SHOTS_API_URL`     | no       | `http://localhost`                         | REST API base for Node-side seeding/queries — must be directly resolvable, unlike the browser's branded host. |
| `SHOTS_NAMESPACE`   | no       | `acme`                                     | Namespace name to set.                               |
| `SHOTS_TENANT`      | no       | `00000000-0000-4000-0000-000000000000`     | Dev tenant id.                                       |
| `SHOTS_OUTPUT_DIR`  | no       | `../public/img`                            | Where PNGs are written.                              |

Never set a password or license path inside a tracked file — pass them via the
environment only.

---

## Step by step

### a) Bring up the dev env and create the `dev` user / namespace

```bash
./bin/docker-compose up -d
# Create the login user (pick your own password):
./bin/cli user create dev 'your-dev-password' dev@local
# Sign in once in the browser to create the namespace, or let the UI prompt you.
```

Then export the required secrets:

```bash
export SHOTS_CHROMIUM="$(which chromium)"   # or a playwright-installed chromium
export SHOTS_PASSWORD='your-dev-password'
```

### b) Enterprise screens (firewall, web endpoints, admin auth, recording)

The enterprise shots are skipped automatically on plain CE (the toolkit detects
402/404/405 and moves on). To capture them you need the enterprise features
enabled with **your own** license.

Everything is driven by environment variables in your gitignored
`shellhub/.env.override` — no `docker-compose.override.yml` needed. Append the
template settings (see `env.override.example`) and point `SHELLHUB_LICENSE_FILE`
at your own license:

```bash
# Append the enterprise settings to your (gitignored) override file:
cat ui/apps/docs/screenshots/env.override.example >> .env.override
# Then edit .env.override and set SHELLHUB_LICENSE_FILE to YOUR license path.
# NEVER commit .env.override or your license. (Both are gitignored / private.)
./bin/docker-compose up -d
```

This enables `SHELLHUB_ENTERPRISE=true`, `SHELLHUB_WEB_ENDPOINTS=true`, the
branded `SHELLHUB_DOMAIN`, and — via the env-driven mount in
`cloud/docker-compose.enterprise.dev.yml` — mounts your license at
`/etc/shellhub/license.dat` and sets `ADMIN_API_LICENSE_FILE` accordingly. The
api loads the license once at startup; it then persists in the database.

> **GeoIP needs no mounts.** `cloud/.env` ships a default `MAXMIND_MIRROR` that
> isn't reachable for local dev, and the enterprise api FATALs at boot if its
> configured GeoIP source can't be fetched. The template sets
> `SHELLHUB_MAXMIND_MIRROR=` (empty) so the api boots with the GeoIP locator
> **disabled** — geo lookups are unused by the screenshots (sessions are local).
> For real geo data instead, drop that line and set `SHELLHUB_MAXMIND_LICENSE`
> to your free MaxMind license key.

### c) Seed demo data

```bash
npm run shots:seed   # in ui/apps/docs/
```

Idempotent. Renames the namespace to `acme`, registers/accepts a set of fake
devices (leaving one pending), adds a sample public key, and — when enterprise
is enabled — firewall rules and a web endpoint. Gated steps log
`skipped (needs enterprise license / feature flag)` and are not fatal.

### d) Capture

```bash
npm run shots:capture            # all shots
npm run shots:capture -- --ce-only
npm run shots:capture -- --only=device-list,dashboard
npm run shots                    # seed + capture in one go
```

PNGs are written to `public/img/...` (1440×900, deviceScaleFactor 2, dark
theme). A summary of captured / skipped / errored shots is printed at the end.

---

## Page map (old → new) and output files

| Shot id                | Route                              | Output (`public/img/…`)            | Edition     |
| ---------------------- | ---------------------------------- | ---------------------------------- | ----------- |
| `dashboard`            | `/dashboard`                       | `getting-started/dashboard.png`    | ce          |
| `device-list`          | `/devices`                         | `devices/device-list.png`          | ce          |
| `add-device`           | `/devices/add`                     | `devices/add-device.png`           | ce          |
| `devices-pending`      | `/devices?tab=pending`             | `devices/devices-pending.png`      | ce          |
| `device-details`       | `/devices/:device`                 | `devices/device-details.png`       | ce          |
| `public-keys`          | `/sshkeys/public-keys`             | `public-keys/public-keys.png`      | ce          |
| `firewall-rules`       | `/firewall-rules`                  | `firewall/firewall-rules.png`      | enterprise  |
| `firewall-add-rule`    | `/firewall-rules` (+click)         | `firewall/add-rule.png`            | enterprise  |
| `sessions-list`        | `/sessions`                        | `sessions/sessions-list.png`       | ce          |
| `session-detail`       | `/sessions/:session`               | `sessions/session-detail.png`      | ce          |
| `session-recording`    | `/sessions/:session` (+Play)       | `sessions/session-recording.png`   | enterprise  |
| `namespace-settings`   | `/settings`                        | `settings/namespace-settings.png`  | ce          |
| `members`              | `/team`                            | `team/members.png`                 | ce          |
| `profile`              | `/profile`                         | `account/profile.png`              | ce          |
| `mfa-enable`           | `/profile` (+Enable MFA)           | `account/mfa-enable.png`           | ce          |
| `admin-authentication` | `/admin/settings/authentication`   | `auth/admin-authentication.png`    | enterprise  |
| `web-endpoints`        | `/web-endpoints`                   | `web-endpoints/web-endpoints.png`  | enterprise  |
| `add-docker-host`      | `/containers` (+Add Docker Host)   | `containers/add-docker-host.png`   | ce          |

The `:device` / `:session` tokens are resolved at capture time to the first
accepted device / most recent session uid.

### SSHID format

ShellHub SSHIDs are `<namespace>.<device>@<server>` — e.g.
`acme.rpi-edge-01@shellhub.example.com`. Because the UI derives this from
`window.location`, browsing via the branded host produces correctly branded
SSHIDs in every screenshot.

---

## Members (manual prerequisite)

Members must reference **existing** users — the seed script does not create
them (that needs the CLI). To populate the team members screen:

```bash
./bin/cli user create alice 'pw' alice@acme.test
./bin/cli user create bob   'pw' bob@acme.test
```

Then add them to the namespace (roles: `administrator` | `operator` |
`observer`):

```
POST /api/namespaces/<tenant>/members { "email": "alice@acme.test", "role": "operator" }
```

---

## Capturing sessions and recordings (manual / optional)

The `session-detail` and `session-recording` shots need a **real** session in
the database. A live SSH session requires an online agent, an authorized SSH
key, and the deny-root firewall rule temporarily removed — this is hard to do
reliably from Node, so `seed.mjs` does **not** attempt it.

To produce a session manually:

1. Make sure an agent is online (e.g. `shellhub-agent-1`) and accepted (named
   `rpi-edge-01` by the seed).
2. Authorize an SSH public key for it (Public Keys screen or the seeded key) and
   temporarily disable/remove the deny-`root` firewall rule.
3. Open a PTY session (recording requires a PTY — the `-tt` flag):

   ```bash
   ssh -tt -i <key> root@acme.rpi-edge-01@shellhub.example.com
   ```

   Type a few commands, then exit. The session (and, on enterprise with
   recording enabled, the recording) will now appear under `/sessions`.

4. Re-run `npm run shots:capture -- --only=session-detail,session-recording`.

---

## Known skips / TODOs

- **Device tags**: the tags endpoints returned 404/405 in testing; tagging is a
  documented skip in `seed.mjs` (see the `TODO(tags)` note). Not blocking.
- **Enterprise-gated endpoints** (firewall, web endpoints) are skipped
  gracefully on CE.
- **Sessions / recordings** are a documented manual step (above).
