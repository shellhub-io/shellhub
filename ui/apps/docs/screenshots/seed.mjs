// Idempotent data seeding for the docs screenshots.
//
// Seeds (or refreshes) the demo data the screenshots depend on:
//   - renames the dev namespace to `acme`
//   - registers + accepts + names a handful of fake devices
//   - (CE)         a sample public key
//   - (enterprise) firewall rules
//   - (enterprise) a web endpoint / tunnel
//
// Enterprise-gated endpoints return 402 (no license) or 404/405 (feature flag
// off). We log "skipped" and continue, so this script also works against a
// plain Community Edition install.
//
// Usage: node seed.mjs
//
// Soft dependency: `ssh-keygen` (for the public-key seed). If it is missing we
// warn and skip that step.
//
// Manual / not seeded here:
//   - device TAGS: the tags endpoints returned 404/405 in testing; treated as a
//     known-unsupported skip (TODO below).
//   - SSH SESSIONS + RECORDING: require a live PTY session against a real online
//     agent; see README "Capturing sessions and recordings".

import { execFileSync } from "node:child_process";
import fs from "node:fs";
import os from "node:os";

import { NAMESPACE, PASSWORD, TENANT, USERNAME } from "./config.mjs";
import {
  api,
  genDevicePublicKeyPem,
  login,
  randomMac,
} from "./lib.mjs";

// Devices to present in the dashboard. `id`/`pretty_name` drive the OS icon.
const DEVICES = [
  { name: "web-server-01", os: "ubuntu" },
  { name: "db-primary", os: "ubuntu" },
  { name: "edge-gateway", os: "debian" },
  { name: "rpi-edge-01", os: "raspbian" },
  { name: "build-runner-03", os: "arch" },
];

// Leave one device pending so the "pending devices" screenshot has content.
const PENDING_DEVICE = { name: "new-laptop", os: "ubuntu" };

const FIREWALL_RULES = [
  {
    priority: 1,
    action: "deny",
    active: true,
    source_ip: ".*",
    username: "root",
    filter: { hostname: ".*" },
  },
  {
    priority: 2,
    action: "allow",
    active: true,
    source_ip: "192.168.0.0/16",
    username: ".*",
    filter: { hostname: ".*" },
  },
];

function logStep(msg) {
  console.log(`seed: ${msg}`);
}

/** Treat enterprise/feature-gated responses as a soft skip. */
function isGated(status) {
  return status === 402 || status === 404 || status === 405;
}

// --- Namespace --------------------------------------------------------------

async function seedNamespace(request) {
  const res = await request("PUT", `/namespaces/${TENANT}`, {
    name: NAMESPACE,
  });
  if (res.ok) {
    logStep(`namespace renamed to "${NAMESPACE}"`);
  } else {
    logStep(`namespace rename returned ${res.status} (continuing)`);
  }
}

// --- Devices ----------------------------------------------------------------

/** Register a fake device and return its uid (or null on failure). */
async function registerDevice(request, { name, os: prettyName }) {
  const res = await request("POST", "/devices/auth", {
    sessions: [],
    info: {
      id: prettyName,
      pretty_name: prettyName,
      version: "latest",
      platform: "native",
    },
    identity: { mac: randomMac() },
    public_key: genDevicePublicKeyPem(),
    tenant_id: TENANT,
  });

  if (!res.ok || !res.data?.uid) {
    logStep(`device "${name}" registration failed (${res.status})`);
    return null;
  }
  return res.data.uid;
}

async function clearDevices(request) {
  for (const status of ["accepted", "pending", "rejected"]) {
    const res = await request("GET", `/devices?status=${status}`);
    if (!res.ok || !Array.isArray(res.data)) continue;
    for (const device of res.data) {
      if (device?.uid) await request("DELETE", `/devices/${device.uid}`);
    }
  }
  logStep("cleared existing devices for a clean slate");
}

async function seedDevices(request) {
  await clearDevices(request);

  for (const device of DEVICES) {
    const uid = await registerDevice(request, device);
    if (!uid) continue;
    await request("PATCH", `/devices/${uid}/accept`);
    await request("PUT", `/devices/${uid}`, { name: device.name });

    // TODO(tags): the device tags endpoints returned 404/405 in testing, so
    // tagging is intentionally skipped here. Re-enable once the API supports it:
    //   await request("POST", `/devices/${uid}/tags`, { tag: "production" });

    logStep(`device "${device.name}" (${device.os}) accepted`);
  }

  // A single pending device for the "pending" screenshot.
  const pendingUid = await registerDevice(request, PENDING_DEVICE);
  if (pendingUid) {
    logStep(`device "${PENDING_DEVICE.name}" left pending`);
  }
}

// --- Public keys (CE) -------------------------------------------------------

/** Mint an OpenSSH public key line using ssh-keygen, or null if unavailable. */
function generateOpenSshPublicKey() {
  const tmp = `${os.tmpdir()}/shots-seed-key-${process.pid}`;
  try {
    execFileSync(
      "ssh-keygen",
      ["-t", "ed25519", "-N", "", "-C", "docs@acme", "-f", tmp],
      { stdio: "ignore" },
    );
    const pub = fs.readFileSync(`${tmp}.pub`, "utf8").trim();
    return pub;
  } catch {
    return null;
  } finally {
    fs.rmSync(tmp, { force: true });
    fs.rmSync(`${tmp}.pub`, { force: true });
  }
}

async function seedPublicKey(request) {
  const sshLine = generateOpenSshPublicKey();
  if (!sshLine) {
    logStep(
      "public keys skipped (ssh-keygen not found — install openssh-client)",
    );
    return;
  }

  const data = Buffer.from(sshLine).toString("base64");
  const res = await request("POST", "/sshkeys/public-keys", {
    name: "ops-laptop",
    username: ".*",
    data,
    filter: { hostname: ".*" },
  });

  if (res.ok) {
    logStep('public key "ops-laptop" created');
  } else if (res.status === 409) {
    logStep("public key already exists (skipping)");
  } else {
    logStep(`public key creation returned ${res.status} (continuing)`);
  }
}

// --- Firewall rules (enterprise) --------------------------------------------

async function seedFirewallRules(request) {
  let gated = false;
  for (const rule of FIREWALL_RULES) {
    const res = await request("POST", "/firewall/rules", rule);
    if (res.ok) {
      logStep(`firewall rule "${rule.action}" (priority ${rule.priority}) created`);
    } else if (isGated(res.status)) {
      gated = true;
      break;
    } else {
      logStep(`firewall rule returned ${res.status} (continuing)`);
    }
  }
  if (gated) {
    logStep("firewall rules skipped (needs enterprise license / feature flag)");
  }
}

// --- Web endpoints / tunnels (enterprise) -----------------------------------

async function seedWebEndpoint(request) {
  // host MUST be a valid IP (not "localhost").
  const accepted = await request("GET", "/devices?status=accepted");
  const uid = Array.isArray(accepted.data) ? accepted.data[0]?.uid : null;
  if (!uid) {
    logStep("web endpoint skipped (no accepted device available)");
    return;
  }

  const res = await request("POST", `/devices/${uid}/tunnels`, {
    host: "127.0.0.1",
    port: 8080,
    ttl: 3600,
  });

  if (res.ok) {
    logStep("web endpoint / tunnel created");
  } else if (isGated(res.status)) {
    logStep(
      "web endpoints skipped (needs enterprise license + SHELLHUB_WEB_ENDPOINTS=true)",
    );
  } else {
    logStep(`web endpoint returned ${res.status} (continuing)`);
  }
}

// --- Members (documented manual step) ---------------------------------------
//
// Members must reference EXISTING users. Create them first, e.g.:
//   ./bin/cli user create alice <password> alice@acme.test
// then add to the namespace via:
//   POST /api/namespaces/<tenant>/members { email, role }   role: administrator|operator|observer
//
// We do not create users from here (that requires the CLI), so member seeding
// is left as a documented manual step in the README.

// --- Main -------------------------------------------------------------------

async function main() {
  const request = api(null);
  const token = await login(request, USERNAME, PASSWORD);
  const authed = api(token);

  await seedNamespace(authed);
  await seedDevices(authed);
  await seedPublicKey(authed);
  await seedFirewallRules(authed);
  await seedWebEndpoint(authed);

  logStep("done");
}

await main();
