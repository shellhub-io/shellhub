// Declarative list of screenshots to capture.
//
// Each entry:
//   id          unique slug, also usable with capture.mjs --only=<id,...>
//   route       in-app path to visit (relative to BASE_URL). May contain the
//               token ":device" or ":session", resolved at capture time to the
//               first accepted device / most recent session uid.
//   output      path of the PNG relative to OUTPUT_DIR (matches the filenames
//               already committed under public/img/).
//   edition     'ce' (works on Community Edition) or 'enterprise'
//               (license/flag-gated; skipped with --ce-only).
//   waitForText optional substring to wait for before shooting (stability).
//   click       optional accessible-name / text to click before shooting
//               (best-effort; failures are logged, not fatal).
//   waitMs      optional extra settle time in ms (default applied in capture).
//
// Output names were matched exactly against the committed files under
// ui/apps/docs/public/img/.

export const manifest = [
  // --- Getting started ------------------------------------------------------
  {
    id: "dashboard",
    route: "/dashboard",
    output: "getting-started/dashboard.png",
    edition: "ce",
    waitForText: "Dashboard",
  },

  // --- Devices --------------------------------------------------------------
  {
    id: "device-list",
    route: "/devices",
    output: "devices/device-list.png",
    edition: "ce",
    waitForText: "Devices",
  },
  {
    id: "add-device",
    route: "/devices/add",
    output: "devices/add-device.png",
    edition: "ce",
  },
  {
    id: "devices-pending",
    route: "/devices?tab=pending",
    output: "devices/devices-pending.png",
    edition: "ce",
    waitForText: "Pending",
  },
  {
    id: "device-details",
    route: "/devices/:device",
    output: "devices/device-details.png",
    edition: "ce",
  },

  // --- Public keys ----------------------------------------------------------
  {
    id: "public-keys",
    route: "/sshkeys/public-keys",
    output: "public-keys/public-keys.png",
    edition: "ce",
    waitForText: "Public Keys",
  },

  // --- Firewall (enterprise) ------------------------------------------------
  {
    id: "firewall-rules",
    route: "/firewall-rules",
    output: "firewall/firewall-rules.png",
    edition: "enterprise",
    waitForText: "Firewall",
  },
  {
    id: "firewall-add-rule",
    route: "/firewall-rules",
    output: "firewall/add-rule.png",
    edition: "enterprise",
    click: "Add Rule",
  },

  // --- Sessions -------------------------------------------------------------
  {
    id: "sessions-list",
    route: "/sessions",
    output: "sessions/sessions-list.png",
    edition: "ce",
    waitForText: "Sessions",
  },
  {
    id: "session-detail",
    route: "/sessions/:session",
    output: "sessions/session-detail.png",
    edition: "ce",
  },
  {
    // Recording playback: open a session, click "Play Recording".
    // Requires a recorded session to exist (enterprise + a real PTY session).
    id: "session-recording",
    route: "/sessions/:session",
    output: "sessions/session-recording.png",
    edition: "enterprise",
    click: "Play Recording",
    waitMs: 2500,
  },

  // --- Settings -------------------------------------------------------------
  {
    id: "namespace-settings",
    route: "/settings",
    output: "settings/namespace-settings.png",
    edition: "ce",
    waitForText: "Settings",
  },

  // --- Team -----------------------------------------------------------------
  {
    id: "members",
    route: "/team",
    output: "team/members.png",
    edition: "ce",
    waitForText: "Members",
  },

  // --- Account --------------------------------------------------------------
  {
    id: "profile",
    route: "/profile",
    output: "account/profile.png",
    edition: "ce",
    waitForText: "Profile",
  },
  {
    id: "mfa-enable",
    route: "/profile",
    output: "account/mfa-enable.png",
    edition: "ce",
    click: "Enable MFA",
    waitMs: 1500,
  },

  // --- Admin (enterprise) ---------------------------------------------------
  {
    id: "admin-authentication",
    route: "/admin/settings/authentication",
    output: "auth/admin-authentication.png",
    edition: "enterprise",
    waitForText: "Authentication",
  },

  // --- Web endpoints (enterprise) -------------------------------------------
  {
    id: "web-endpoints",
    route: "/web-endpoints",
    output: "web-endpoints/web-endpoints.png",
    edition: "enterprise",
    waitForText: "Web Endpoints",
  },

  // --- Containers -----------------------------------------------------------
  {
    // The "Add Docker Host" connector drawer is opened from the containers page.
    id: "add-docker-host",
    route: "/containers",
    output: "containers/add-docker-host.png",
    edition: "ce",
    click: "Add Docker Host",
    waitMs: 1500,
  },
];

export default manifest;
