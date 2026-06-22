// Central configuration for the docs screenshot toolkit.
//
// Everything is driven by environment variables with sane defaults so the
// toolkit can run unattended in CI or locally. Two values are intentionally
// REQUIRED (no insecure defaults): the login password and the Chromium binary
// path.
//
// NOTE: never put a real password or a license path in here. Pass secrets via
// the environment only.

import path from "node:path";
import { fileURLToPath } from "node:url";

const __dirname = path.dirname(fileURLToPath(import.meta.url));

/** Resolve a path relative to the screenshots/ directory. */
export const resolveFromHere = (p) => path.resolve(__dirname, p);

/** Read a required env var or throw a helpful error. */
function requireEnv(name, hint) {
  const value = process.env[name];
  if (!value) {
    throw new Error(
      `Missing required environment variable ${name}.\n${hint}`,
    );
  }
  return value;
}

// --- Target / branding ------------------------------------------------------

// The branded host. The UI derives SSHIDs and the agent install command from
// window.location, so capturing through the branded host rebrands the whole UI
// without any backend change. lib.launchBrowser() maps this domain to
// 127.0.0.1 via Chromium's --host-resolver-rules.
export const DOMAIN = process.env.SHOTS_DOMAIN || "shellhub.example.com";
export const BASE_URL = process.env.SHOTS_BASE_URL || `http://${DOMAIN}`;

// Two different worlds reach the backend:
//   - The BROWSER navigates BASE_URL (the branded host). Chromium's
//     --host-resolver-rules maps that host to 127.0.0.1, so no /etc/hosts entry
//     is needed for page.goto().
//   - Node's global fetch() (seed.mjs, capture.mjs token resolution) does NOT
//     honor the browser's host-resolver, so it must hit a host the OS can
//     actually resolve. Hence API_URL defaults to localhost. The REST API does
//     not affect branding (SSHIDs come from window.location in the UI), so
//     Node can safely talk to localhost while the browser uses the branded host.
export const API_URL = process.env.SHOTS_API_URL || "http://localhost";
export const API = `${API_URL}/api`;

export const NAMESPACE = process.env.SHOTS_NAMESPACE || "acme";
export const TENANT =
  process.env.SHOTS_TENANT || "00000000-0000-4000-0000-000000000000";

// --- Credentials ------------------------------------------------------------

export const USERNAME = process.env.SHOTS_USERNAME || "dev";

// No insecure hardcoded default: the password must come from the environment.
export const PASSWORD = requireEnv(
  "SHOTS_PASSWORD",
  "Set it to the password of the dev user, e.g.:\n" +
    "  export SHOTS_PASSWORD='your-dev-password'\n" +
    "Create the user first with: ./bin/cli user create dev <password> dev@local",
);

// --- Browser ----------------------------------------------------------------

// Path to a Chromium/Chrome executable usable by playwright-core. Required so we
// never hardcode a /nix/store (or any machine-specific) path.
export const CHROMIUM = requireEnv(
  "SHOTS_CHROMIUM",
  "Point it at a Chromium/Chrome binary, e.g.:\n" +
    "  export SHOTS_CHROMIUM=\"$(which chromium)\"\n" +
    "  # or the path printed by: npx playwright install chromium",
);

export const VIEWPORT = { width: 1440, height: 900 };
export const DEVICE_SCALE_FACTOR = 2;

// --- Output / state ---------------------------------------------------------

// Where PNGs are written. Defaults to the docs app's committed image folder so
// re-running the toolkit refreshes the shots in place.
export const OUTPUT_DIR = process.env.SHOTS_OUTPUT_DIR
  ? path.resolve(process.env.SHOTS_OUTPUT_DIR)
  : resolveFromHere("../public/img");

// Playwright storageState cache (gitignored).
export const STATE_PATH = resolveFromHere(".state.json");

// Reuse a cached auth state if it is younger than this many milliseconds.
export const STATE_MAX_AGE_MS = 1000 * 60 * 60 * 6; // 6 hours
