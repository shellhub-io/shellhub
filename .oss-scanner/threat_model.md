# Threat model

## What ShellHub does and where untrusted input enters

ShellHub is a centralized SSH gateway. Devices run an agent that dials out to the server and keeps a
reverse tunnel open, so users reach a device behind NAT or a firewall with
`ssh <user>@<namespace>.<device>@<server>`. One server holds many namespaces, ShellHub's tenants.
Each namespace has its own devices, members, API keys, public keys, firewall rules and session
recordings.

Untrusted input enters at:

- The HTTP API in `server/api`, built on Echo. It takes user logins, JWTs, API keys in the
  `X-API-KEY` header, and the calls the agent makes to register and authenticate a device.
- The SSH server in `server/ssh`, built on ShellHub's fork of gliderlabs/ssh. Any client on the
  internet can connect and present an SSHID, a password or a public key.
- The agent connection. An accepted device talks to the server over a reverse tunnel through
  `pkg/revdial` and `pkg/wsconnadapter`. Treat every device as untrusted, since its owner may be
  hostile to other namespaces.
- The web terminal's WebSocket sessions in `server/ssh/web`.
- The HTTP tunnels to device ports. The server opens them through `server/ssh/pkg/dialer`, the
  agent serves them in `agent/pkg/agentd`, and `pkg/webendpoints` parses their hostnames.
- The gateway in `gateway/`, which embeds Caddy as a library. It is the public entry point and
  only routes requests. It does no authentication and passes client headers through, so the API
  must delete and rewrite every identity header, such as `X-ID`, `X-Tenant-ID` and `X-Role`. That
  happens in `server/api/pkg/gateway`.
- The React web console in `ui/apps/console`. It renders data that other users and devices
  control, such as device names, tags and session output.

## Components that matter most and least

Most important:

- Namespace isolation. A user, API key or device of one namespace must never read or change data
  of another namespace, or open a session to its devices.
- Authentication and authorization in `server/api`, including the role checks for observer,
  operator, administrator and owner, and the permission each route declares.
- The SSH path: SSHID parsing, authentication, firewall rules, session handling, and the dialer
  that connects a session to a device tunnel.
- The agent in `agent/`. It runs as root on customer devices. A way for the server path, another
  device or a network attacker to run code or read files through the agent is serious.

Less important but still in scope: the gateway's Caddy configuration, the console UI and
`install.sh`.

Out of scope: the `tests/` integration suite, `devscripts/`, the `docker-compose*.yml` development
setups, the static sites in `ui/apps/website` and `ui/apps/docs`, and vulnerabilities in
third-party dependencies with no exploit path through ShellHub.

## How to exercise it

- The image builds `/out/server`, `/out/agent` and `/out/gateway`. The console build is in
  `ui/apps/console/dist`.
- Run the unit tests with `go test ./...` inside `server/`, `gateway/` and `agent/`, adding
  `-tags docker` for the agent, and with `go test ./pkg/...` from the root. Run `npm test` in
  `ui/apps/console` for the console.
- The server needs PostgreSQL and Redis to run. Some tests start them through testcontainers and
  need Docker, which the scan environment lacks, so they fail with `rootless Docker not found`.
  These are the PostgreSQL tests in `server/api/store/pg` and `server/api/services`, the Redis
  tests in `pkg/cache` and `pkg/worker/asynq`, `pkg/testrun`, and the `tests/` integration suite.
  A failure there is not a regression. Use unit tests and code reading for those paths.
- The OpenAPI spec in `openapi/spec/` describes the API.

## How we rate severity

We score with CVSS 3.1 and use this guidance.

Critical:

- unauthenticated remote code execution on the server, or on a device through the agent
- unauthenticated access to a device shell
- authentication bypass on the API or the SSH server
- shell access to another namespace's devices, or to its active sessions

High:

- reading or writing another namespace's devices, members, keys or recordings without shell access
- privilege escalation inside a namespace, such as an observer gaining owner rights
- a malicious device reaching the server host or another namespace
- stored XSS in the console that runs in another user's session

Medium:

- disclosure inside one namespace of data the caller's role does not allow
- SSRF with limited reach
- CSRF on a state-changing action
- denial of service that an unauthenticated attacker triggers with little traffic

Low:

- an issue that needs an administrator or owner to harm their own namespace
- self-XSS
- missing hardening with no demonstrated impact

A report should show the impact with a reproducer or a test. A finding with no demonstrated path to
impact is Low at most.

## Anything to leave alone

- Rate limiting, and denial of service by traffic volume against a deployment.
- An owner or administrator acting within their own namespace, since they already hold that
  authority.
- Findings that need the server host, its database or its secrets already compromised.
