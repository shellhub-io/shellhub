// Main screenshot runner.
//
// Flow: ensureAuth() -> load manifest -> launch browser (branded host mapped to
// localhost) -> iterate shots -> write PNGs to OUTPUT_DIR -> print a summary.
//
// Flags:
//   --ce-only         skip shots with edition:'enterprise'
//   --only=<id,id>    capture only the listed manifest ids
//
// Usage: node capture.mjs [--ce-only] [--only=device-list,dashboard]

import fs from "node:fs";
import path from "node:path";

import { ensureAuth } from "./auth.mjs";
import {
  API,
  BASE_URL,
  OUTPUT_DIR,
  PASSWORD,
  STATE_PATH,
  USERNAME,
} from "./config.mjs";
import {
  api,
  contextOptions,
  launchBrowser,
  login,
  removeDevOverlay,
  sleep,
} from "./lib.mjs";
import manifest from "./manifest.mjs";

const DEFAULT_SETTLE_MS = 1500;

// --- CLI args ---------------------------------------------------------------

function parseArgs(argv) {
  const ceOnly = argv.includes("--ce-only");
  const onlyArg = argv.find((a) => a.startsWith("--only="));
  const only = onlyArg
    ? onlyArg
        .slice("--only=".length)
        .split(",")
        .map((s) => s.trim())
        .filter(Boolean)
    : null;
  return { ceOnly, only };
}

function selectShots({ ceOnly, only }) {
  let shots = manifest;
  if (ceOnly) shots = shots.filter((s) => s.edition !== "enterprise");
  if (only) shots = shots.filter((s) => only.includes(s.id));
  return shots;
}

// --- Dynamic route tokens ---------------------------------------------------

/**
 * Resolve :device / :session tokens to real uids via the API. Returns a map of
 * token -> uid (or null when unavailable). Done once up front so we don't query
 * per shot.
 */
async function resolveTokens(shots) {
  const needsDevice = shots.some((s) => s.route.includes(":device"));
  const needsSession = shots.some((s) => s.route.includes(":session"));
  const tokens = {};

  if (!needsDevice && !needsSession) return tokens;

  const request = api(null);
  const token = await login(request, USERNAME, PASSWORD);
  const authed = api(token);

  if (needsDevice) {
    const res = await authed("GET", "/devices?status=accepted");
    tokens[":device"] = Array.isArray(res.data) ? res.data[0]?.uid : null;
  }
  if (needsSession) {
    const res = await authed("GET", "/sessions");
    tokens[":session"] = Array.isArray(res.data) ? res.data[0]?.uid : null;
  }
  return tokens;
}

function applyTokens(route, tokens) {
  for (const [token, value] of Object.entries(tokens)) {
    if (route.includes(token)) {
      if (!value) return null; // unresolved -> caller should skip
      route = route.replace(token, value);
    }
  }
  return route;
}

// --- Capture ----------------------------------------------------------------

async function captureShot(page, shot, tokens) {
  const route = applyTokens(shot.route, tokens);
  if (route === null) {
    return { id: shot.id, status: "skipped", reason: "no data for dynamic route" };
  }

  await page.goto(`${BASE_URL}${route}`, {
    waitUntil: "networkidle",
    timeout: 30000,
  });

  if (shot.waitForText) {
    const found = await page
      .getByText(shot.waitForText, { exact: false })
      .first()
      .waitFor({ timeout: 15000 })
      .then(() => true)
      .catch(() => false);
    if (!found) {
      console.log(
        `capture: warn ${shot.id} — waitForText "${shot.waitForText}" not found, proceeding`,
      );
    }
  }

  await sleep(shot.waitMs ?? DEFAULT_SETTLE_MS);

  if (shot.click) {
    const clicked = await page
      .getByRole("button", { name: shot.click, exact: false })
      .first()
      .click({ timeout: 8000 })
      .then(() => true)
      .catch(() =>
        page
          .getByText(shot.click, { exact: false })
          .first()
          .click({ timeout: 5000 })
          .then(() => true)
          .catch(() => false),
      );
    if (!clicked) {
      return {
        id: shot.id,
        status: "skipped",
        reason: `click target not found: "${shot.click}"`,
      };
    }
    await sleep(shot.waitMs ?? DEFAULT_SETTLE_MS);
  }

  await removeDevOverlay(page);

  const outPath = path.join(OUTPUT_DIR, shot.output);
  fs.mkdirSync(path.dirname(outPath), { recursive: true });
  await page.screenshot({ path: outPath });

  return { id: shot.id, status: "ok", output: shot.output };
}

// --- Main -------------------------------------------------------------------

async function main() {
  const args = parseArgs(process.argv.slice(2));
  const shots = selectShots(args);

  if (shots.length === 0) {
    console.log("capture: no shots selected.");
    return;
  }

  console.log(`capture: ${shots.length} shot(s) via ${BASE_URL} (api: ${API})`);
  if (args.ceOnly) console.log("capture: --ce-only (skipping enterprise shots)");

  await ensureAuth();

  const tokens = await resolveTokens(shots);

  const browser = await launchBrowser();
  const results = [];
  try {
    const context = await browser.newContext(
      contextOptions({ storageState: STATE_PATH }),
    );
    const page = await context.newPage();
    // Tailwind media dark mode needs this on the PAGE, not just the context.
    await page.emulateMedia({ colorScheme: "dark" });

    for (const shot of shots) {
      try {
        const result = await captureShot(page, shot, tokens);
        results.push(result);
        const tag = result.status === "ok" ? "shot" : "skip";
        console.log(
          `capture: ${tag} ${shot.id}` +
            (result.reason ? ` (${result.reason})` : ` -> ${result.output}`),
        );
      } catch (err) {
        results.push({ id: shot.id, status: "error", reason: err.message });
        console.log(`capture: FAIL ${shot.id} (${err.message})`);
      }
    }
  } finally {
    await browser.close();
  }

  // --- Summary --------------------------------------------------------------
  const ok = results.filter((r) => r.status === "ok").length;
  const skipped = results.filter((r) => r.status === "skipped").length;
  const errored = results.filter((r) => r.status === "error").length;

  console.log("\ncapture summary:");
  console.log(`  captured: ${ok}`);
  console.log(`  skipped:  ${skipped}`);
  console.log(`  errors:   ${errored}`);
  console.log(`  output:   ${OUTPUT_DIR}`);

  if (errored > 0) process.exitCode = 1;
}

await main();
