# Registration webhook — demo integrator

A tiny, self-contained integrator for ShellHub's install-key **webhook** registration mode.

When an install key is set to webhook mode, ShellHub POSTs a signed JSON payload
to the key's webhook URL at device registration and waits for a decision. This demo
receives that call, verifies the HMAC signature, applies a live policy
(accept / reject / pending / defer / MAC allowlist), and shows a live feed in a
small dark web UI.

It exercises both decision paths: the **synchronous** response and the **async
deferred** callback (return `defer`, then POST the real decision to the payload's
`callback_url` later).

Stdlib only, no third-party dependencies.

## Run

```bash
go run . -addr :9090 -secret my-shared-secret
```

Flags:

- `-addr` — listen address (default `:9090`)
- `-secret` — the webhook HMAC secret; must match the install key's secret (default `dev-secret`)

Open the UI at http://localhost:9090.

## Point a ShellHub install key at it

On the install key (webhook mode), configure:

- **Webhook URL**: `http://<this-host>:9090/webhook` — must be reachable from the
  ShellHub API container. When the API runs in Docker, `localhost` there is the
  container, not your machine; use the host's LAN IP or a tunnel.
- **Secret**: the exact value passed to `-secret`.

Use the "Simulate registration" button to exercise the full signed path without a
real device.

## The contract

ShellHub POSTs to `/webhook`:

```json
{
  "tenant_id": "3dd0d1f8-...",
  "install_key_id": "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
  "install_key_name": "ci-runners",
  "device_uid": "13b0c8ea...",
  "mac": "00:1a:2b:3c:4d:5e",
  "hostname": "web-01",
  "info": { "id": "debian", "pretty_name": "Debian GNU/Linux 12", "version": "v0.18.0", "arch": "amd64", "platform": "docker" },
  "source_ip": "203.0.113.7",
  "timestamp": "2026-01-02T03:04:05Z",
  "callback_url": "http://.../api/devices/enroll/callback/<token>"
}
```

`install_key_id` is the key's SHA256 digest — stable and non-secret. Key your policy
on it: it survives a rename, the name can't. `install_key_name` rides along only for
human recognition in logs/UI — don't match on it. The secret is never sent, not even
a fragment; authenticity comes from the signature header below. `callback_url` is a
ready, token-authenticated URL for a late decision, present whenever ShellHub knows
its public base.

Header: `X-ShellHub-Signature: hex(HMAC-SHA256(secret, rawBody))`.

The integrator answers with 2xx and:

```json
{ "decision": "accept" | "reject" | "pending" | "defer", "reason": "optional" }
```

The decision maps to the device's initial status (accept → accepted,
reject → rejected, pending → manual queue). ShellHub **fails closed to pending**
on timeout, a non-2xx status, an unparseable body, or an unknown decision.

### Async / deferred decision

Returning `{"decision":"defer"}` lands the device **pending** and tells ShellHub
you'll decide later. The integrator then POSTs the real decision to the payload's
`callback_url` — no API key, the token in the URL is the credential:

```
POST <callback_url>
{ "decision": "accept" | "reject", "reason": "optional" }
```

## Endpoints

- `GET  /` — the web UI
- `POST /webhook` — the ShellHub-facing endpoint (signature-verified)
- `GET  /api/state` — current policy, allowlist, secret, and recent events (newest first, each with a stable `id`)
- `POST /api/policy` — set the live policy: `{ "policy": "...", "allowlist": ["aa:bb:.."] }`
- `POST /api/simulate` — craft a signed fake registration and POST it to `/webhook`
- `POST /api/decide` — resolve a deferred event: `{ "id": "<event id>", "decision": "accept"|"reject" }`; the server POSTs the decision to that event's stored `callback_url` and records the HTTP outcome
- `POST /api/callback-sink/<token>` — local stand-in for ShellHub's callback receiver, so simulated deferred registrations resolve end-to-end without a running ShellHub

## Policies

- **accept** — accept every registration
- **reject** — reject every registration
- **pending** — leave every registration for ShellHub's manual queue
- **defer** — answer `defer` synchronously, then decide later via the callback URL (Accept / Reject buttons on the row)
- **allowlist** — accept if `mac` is in the allowlist (case-insensitive), else reject

An unverified signature is always recorded but answered with
`{"decision":"pending","reason":"signature mismatch"}`.

## Demo the async path

1. Set the policy to **Defer (async)**.
2. Click **Simulate registration** — the row shows a blue `defer` pill (the device
   would be pending in ShellHub). Its `callback_url` points at this server's local
   sink.
3. Click **Accept** or **Reject** on that row. The server POSTs the decision to the
   `callback_url` and the row shows the outcome (`callback → 200` in green).

With a real ShellHub, steps 1–2 happen when a device registers, and step 3 POSTs to
ShellHub's real token-authenticated callback URL.
