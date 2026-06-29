// Authentication: log into the branded UI and persist a Playwright
// storageState so the capture run starts already signed in.
//
// The app stores its session under the localStorage key `shellhub-session`,
// which storageState captures automatically.

import fs from "node:fs";

import {
  BASE_URL,
  PASSWORD,
  STATE_MAX_AGE_MS,
  STATE_PATH,
  USERNAME,
} from "./config.mjs";
import { contextOptions, launchBrowser } from "./lib.mjs";

/** True if a cached storageState exists and is younger than STATE_MAX_AGE_MS. */
function freshStateExists() {
  try {
    const stat = fs.statSync(STATE_PATH);
    return Date.now() - stat.mtimeMs < STATE_MAX_AGE_MS;
  } catch {
    return false;
  }
}

/**
 * Ensure a valid auth state exists at STATE_PATH.
 *
 * Reuses a fresh cached state when present (unless `force` is true), otherwise
 * performs an interactive login through the branded host and writes the state.
 *
 * @param {{ force?: boolean }} [options]
 * @returns {Promise<string>} the path to the storageState file
 */
export async function ensureAuth({ force = false } = {}) {
  if (!force && freshStateExists()) {
    console.log(`auth: reusing cached state (${STATE_PATH})`);
    return STATE_PATH;
  }

  console.log(`auth: logging in as ${USERNAME} at ${BASE_URL}/login`);
  const browser = await launchBrowser();
  try {
    const context = await browser.newContext(contextOptions());
    const page = await context.newPage();

    await page.goto(`${BASE_URL}/login`, {
      waitUntil: "networkidle",
      timeout: 30000,
    });

    await page.getByPlaceholder("username").fill(USERNAME);
    await page.getByPlaceholder("password").fill(PASSWORD);

    await Promise.all([
      page
        .waitForURL((url) => url.pathname.startsWith("/dashboard"), {
          timeout: 30000,
        })
        .catch(() => {}),
      page.getByRole("button", { name: /sign in/i }).click(),
    ]);

    if (!page.url().includes("/dashboard")) {
      throw new Error(
        `Login did not reach /dashboard (ended at ${page.url()}). ` +
          "Check credentials and that the namespace exists.",
      );
    }

    await context.storageState({ path: STATE_PATH });
    console.log(`auth: saved state to ${STATE_PATH}`);
    return STATE_PATH;
  } finally {
    await browser.close();
  }
}

// Allow running standalone: `node auth.mjs [--force]`
if (import.meta.url === `file://${process.argv[1]}`) {
  await ensureAuth({ force: process.argv.includes("--force") });
}
