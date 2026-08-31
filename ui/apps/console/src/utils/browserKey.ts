import { ed25519PublicKeyLine, sha256Fingerprint } from "./sshKeys";

/**
 * A per-browser Ed25519 SSH identity. The private key is a non-extractable CryptoKey persisted
 * in IndexedDB (structured-cloned, never serialized), so it signs the SSH challenge in-browser
 * and can never leave it. The public half is enrolled as an ssh_identity, which is what makes
 * the web terminal a first-class identity rather than a shared credential.
 */
export interface BrowserKey {
  privateKey: CryptoKey;
  publicKeyLine: string;
  fingerprint: string;
}

const DB_NAME = "shellhub";
const DB_VERSION = 1;
const STORE = "browser-keys";

function openDB(): Promise<IDBDatabase> {
  return new Promise((resolve, reject) => {
    const req = indexedDB.open(DB_NAME, DB_VERSION);
    req.onupgradeneeded = () => req.result.createObjectStore(STORE);
    req.onsuccess = () => resolve(req.result);
    req.onerror = () => reject(req.error ?? new Error("IndexedDB open failed"));
  });
}

function idbGet(db: IDBDatabase, key: string): Promise<BrowserKey | undefined> {
  return new Promise((resolve, reject) => {
    const req = db.transaction(STORE, "readonly").objectStore(STORE).get(key);
    req.onsuccess = () => resolve(req.result as BrowserKey | undefined);
    req.onerror = () => reject(req.error ?? new Error("IndexedDB read failed"));
  });
}

function idbPut(
  db: IDBDatabase,
  key: string,
  value: BrowserKey,
): Promise<void> {
  return new Promise((resolve, reject) => {
    const tx = db.transaction(STORE, "readwrite");
    tx.objectStore(STORE).put(value, key);
    tx.oncomplete = () => resolve();
    tx.onerror = () => reject(tx.error ?? new Error("IndexedDB write failed"));
  });
}

/**
 * Return this browser's SSH identity key for the given scope: the one stored in
 * IndexedDB, or a freshly generated keypair (in memory, NOT yet persisted) when
 * none exists. Whether it needs registering is decided by the caller against the
 * server (the source of truth), not by local presence — persist it with
 * persistBrowserKey once registration is granted. The scope is "<userId>:<tenant>"
 * so a second user on the same profile never shares a key. Returns null when
 * WebCrypto Ed25519 or IndexedDB is unavailable (older browsers), so the caller
 * falls back to the keyless JWT web path.
 */
export async function ensureBrowserKey(
  scope: string,
): Promise<BrowserKey | null> {
  if (
    typeof indexedDB === "undefined" ||
    typeof crypto?.subtle?.generateKey !== "function"
  ) {
    return null;
  }

  try {
    const db = await openDB();
    const existing = await idbGet(db, scope);
    db.close();
    if (existing) {
      return existing;
    }

    const pair = await crypto.subtle.generateKey({ name: "Ed25519" }, false, [
      "sign",
      "verify",
    ]);

    const raw = new Uint8Array(
      await crypto.subtle.exportKey("raw", pair.publicKey),
    );

    return {
      privateKey: pair.privateKey,
      publicKeyLine: ed25519PublicKeyLine(raw),
      fingerprint: await sha256Fingerprint(raw),
    };
  } catch {
    return null;
  }
}

/**
 * The fingerprint of the key this browser already holds for the scope, or null
 * when it holds none. Unlike ensureBrowserKey this never generates one, so it
 * can answer "is this row my key?" without minting a key nobody asked for.
 */
export async function storedBrowserKeyFingerprint(
  scope: string,
): Promise<string | null> {
  if (typeof indexedDB === "undefined") {
    return null;
  }

  try {
    const db = await openDB();
    const existing = await idbGet(db, scope);
    db.close();

    return existing?.fingerprint ?? null;
  } catch {
    return null;
  }
}

/** Persist a freshly generated browser key so future sessions reuse it. Called
 * after the user consents to registering this browser as an identity. */
export async function persistBrowserKey(
  scope: string,
  key: BrowserKey,
): Promise<void> {
  const db = await openDB();
  try {
    await idbPut(db, scope, key);
  } finally {
    db.close();
  }
}

interface UADataBrand {
  brand: string;
  version: string;
}
interface NavigatorUAData {
  brands: UADataBrand[];
  platform: string;
}

/**
 * A human label for this browser with version ("Chrome 149 on Linux"), used as
 * the default name for the enrolled identity so the SSH Identities list
 * distinguishes one browser from another. Prefers UA client hints; falls back to
 * parsing the UA string. It is only a default — the user can rename it.
 */
export function browserLabel(): string {
  const uaData = (navigator as Navigator & { userAgentData?: NavigatorUAData })
    .userAgentData;
  if (uaData?.brands?.length) {
    const brand = uaData.brands.find(
      (b) => !/Not.?A.?Brand/i.test(b.brand) && b.brand !== "Chromium",
    );
    const name = (brand?.brand ?? "Browser")
      .replace(/^Google\s+/, "")
      .replace(/^Microsoft\s+/, "");
    const version = brand?.version ? ` ${brand.version}` : "";

    return `${name}${version} on ${uaData.platform}`;
  }

  const ua = navigator.userAgent;
  const os = /Windows/.test(ua)
    ? "Windows"
    : /Mac OS X|Macintosh/.test(ua)
      ? "macOS"
      : /Android/.test(ua)
        ? "Android"
        : /iPhone|iPad|iPod/.test(ua)
          ? "iOS"
          : /Linux/.test(ua)
            ? "Linux"
            : "Unknown OS";
  const rules: [RegExp, string][] = [
    [/Edg\/(\d+)/, "Edge"],
    [/OPR\/(\d+)/, "Opera"],
    [/Firefox\/(\d+)/, "Firefox"],
    [/Chrome\/(\d+)/, "Chrome"],
    [/Version\/(\d+)[^ ]* Safari/, "Safari"],
  ];
  let browser = "Browser";
  for (const [re, label] of rules) {
    const match = re.exec(ua);
    if (match) {
      browser = `${label} ${match[1]}`;
      break;
    }
  }

  return `${browser} on ${os}`;
}
