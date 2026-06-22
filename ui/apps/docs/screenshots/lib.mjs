// Shared helpers for the docs screenshot toolkit.

import crypto from "node:crypto";
import { chromium } from "playwright-core";

import {
  API,
  CHROMIUM,
  DEVICE_SCALE_FACTOR,
  DOMAIN,
  VIEWPORT,
} from "./config.mjs";

/**
 * Launch a headless Chromium suitable for capturing the dashboard.
 *
 * The --host-resolver-rules arg makes the branded DOMAIN resolve to localhost
 * so the UI rebrands itself (SSHIDs, install command) from window.location.
 */
export async function launchBrowser() {
  return chromium.launch({
    executablePath: CHROMIUM,
    args: [
      "--no-sandbox",
      "--disable-dev-shm-usage",
      `--host-resolver-rules=MAP ${DOMAIN} 127.0.0.1`,
    ],
  });
}

/** Standard browser context options shared by auth + capture. */
export function contextOptions(extra = {}) {
  return {
    viewport: VIEWPORT,
    deviceScaleFactor: DEVICE_SCALE_FACTOR,
    // colorScheme here is not sufficient for Tailwind media dark mode; callers
    // must ALSO call page.emulateMedia({ colorScheme: 'dark' }) on each page.
    colorScheme: "dark",
    ...extra,
  };
}

/**
 * Minimal fetch wrapper for the ShellHub REST API.
 *
 * Returns a function `request(method, path, body?)` that resolves to
 * `{ status, ok, data }`. Never throws on non-2xx so callers can branch on
 * enterprise-gated status codes (402/404/405).
 */
export function api(token) {
  return async function request(method, path, body) {
    const headers = { "Content-Type": "application/json" };
    if (token) headers.Authorization = `Bearer ${token}`;

    const res = await fetch(`${API}${path}`, {
      method,
      headers,
      body: body === undefined ? undefined : JSON.stringify(body),
    });

    let data = null;
    const text = await res.text();
    if (text) {
      try {
        data = JSON.parse(text);
      } catch {
        data = text;
      }
    }

    return { status: res.status, ok: res.ok, data };
  };
}

/** Log in and return a bearer token (throws on failure). */
export async function login(request, username, password) {
  const res = await request("POST", "/login", { username, password });
  if (!res.ok || !res.data?.token) {
    throw new Error(
      `Login failed (status ${res.status}). Check SHOTS_USERNAME / SHOTS_PASSWORD ` +
        "and that the dev user exists (./bin/cli user create ...).",
    );
  }
  return res.data.token;
}

/**
 * Remove dev-only fixed overlays before a screenshot.
 *
 * The TanStack Query devtools button renders into a shadow DOM host
 * ([class*="tsqd"]); we remove the host element. We also defensively strip any
 * tiny high-z-index fixed widget pinned to the bottom of the viewport (chat
 * bubbles, etc.).
 */
export async function removeDevOverlay(page) {
  await page
    .evaluate(() => {
      document
        .querySelectorAll(
          '[class*="tsqd"], .tsqd-open-btn-container, .tsqd-parent-container',
        )
        .forEach((el) => el.remove());

      for (const el of document.querySelectorAll("body *")) {
        const style = getComputedStyle(el);
        const rect = el.getBoundingClientRect();
        if (
          style.position === "fixed" &&
          parseInt(style.zIndex || "0", 10) >= 99999 &&
          rect.width < 120 &&
          rect.bottom > 700
        ) {
          el.remove();
        }
      }
    })
    .catch(() => {});
}

/**
 * Generate an RSA public key in PEM form for a fake device's `public_key`
 * field. ShellHub's /devices/auth accepts a PKCS#1 / SubjectPublicKeyInfo PEM
 * here. Uses Node crypto only (no openssl dependency).
 */
export function genDevicePublicKeyPem() {
  const { publicKey } = crypto.generateKeyPairSync("rsa", {
    modulusLength: 2048,
  });
  return publicKey.export({ type: "spki", format: "pem" }).toString();
}

/** Generate a locally-administered random MAC address (02:xx:xx:xx:xx:xx). */
export function randomMac() {
  const octet = () =>
    Math.floor(Math.random() * 256)
      .toString(16)
      .padStart(2, "0");
  return `02:${octet()}:${octet()}:${octet()}:${octet()}:${octet()}`;
}

/** Small sleep helper. */
export const sleep = (ms) => new Promise((resolve) => setTimeout(resolve, ms));
