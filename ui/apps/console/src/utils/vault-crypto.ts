import type { VaultMeta, VaultData } from "@/types/vault";

const VERIFIER_PLAINTEXT = "shellhub-vault-ok";
const PBKDF2_ITERATIONS = 600_000;
const SALT_BYTES = 16;
const IV_BYTES = 12;

function toArrayBuffer(arr: Uint8Array): ArrayBuffer {
  return arr.buffer.slice(arr.byteOffset, arr.byteOffset + arr.byteLength) as ArrayBuffer;
}

let sessionKey: CryptoKey | null = null;

/**
 * Holds the unwrapped vault key for the life of the tab. Module state rather than storage: the
 * key must not survive a reload, or locking the vault would not lock anything.
 */
export function setSessionKey(key: CryptoKey): void {
  sessionKey = key;
}

/**
 * The current vault key, or null when the vault is locked.
 */
export function getSessionKey(): CryptoKey | null {
  return sessionKey;
}

/**
 * Drops the vault key. This is what locking is — every later read has to derive the key from the
 * passphrase again.
 */
export function clearSessionKey(): void {
  sessionKey = null;
}

function toBase64(buffer: ArrayBuffer | Uint8Array): string {
  const bytes = buffer instanceof Uint8Array ? buffer : new Uint8Array(buffer);
  let binary = "";
  for (let i = 0; i < bytes.byteLength; i++) {
    binary += String.fromCharCode(bytes[i]);
  }
  return btoa(binary);
}

function fromBase64(base64: string): Uint8Array {
  let binary: string;
  try {
    binary = atob(base64);
  } catch {
    throw new Error("Invalid base64 encoding");
  }
  const bytes = new Uint8Array(binary.length);
  for (let i = 0; i < binary.length; i++) {
    bytes[i] = binary.charCodeAt(i);
  }
  return bytes;
}

/**
 * Derives the vault key from a passphrase with PBKDF2. iterations is a parameter rather than a
 * constant because a vault records the count it was created with — raising the default must
 * not lock existing vaults out.
 */
export async function deriveKey(
  password: string,
  salt: Uint8Array,
  iterations: number = PBKDF2_ITERATIONS,
): Promise<CryptoKey> {
  const encoder = new TextEncoder();
  const keyMaterial = await crypto.subtle.importKey(
    "raw",
    encoder.encode(password),
    "PBKDF2",
    false,
    ["deriveKey"],
  );

  return crypto.subtle.deriveKey(
    {
      name: "PBKDF2",
      salt: toArrayBuffer(salt),
      iterations,
      hash: "SHA-256",
    },
    keyMaterial,
    { name: "AES-GCM", length: 256 },
    false,
    ["encrypt", "decrypt"],
  );
}

/**
 * Encrypts the vault body with AES-GCM under a fresh random IV, returned alongside the
 * ciphertext. A new IV every time is required: reusing one under the same key breaks GCM.
 */
export async function encrypt(
  key: CryptoKey,
  plaintext: string,
): Promise<VaultData> {
  const encoder = new TextEncoder();
  const iv = crypto.getRandomValues(new Uint8Array(IV_BYTES));

  const ciphertext = await crypto.subtle.encrypt(
    { name: "AES-GCM", iv: toArrayBuffer(iv) },
    key,
    encoder.encode(plaintext),
  );

  return {
    iv: toBase64(iv),
    ciphertext: toBase64(ciphertext),
  };
}

/**
 * Decrypts a vault body. Rejects if the key is wrong or the ciphertext was altered — GCM
 * authenticates, so a failure here means tampering or a bad key, not corrupt output.
 */
export async function decrypt(
  key: CryptoKey,
  data: VaultData,
): Promise<string> {
  const iv = fromBase64(data.iv);
  const ciphertext = fromBase64(data.ciphertext);

  const plaintext = await crypto.subtle.decrypt(
    { name: "AES-GCM", iv: toArrayBuffer(iv) },
    key,
    toArrayBuffer(ciphertext),
  );

  return new TextDecoder().decode(plaintext);
}

/**
 * Creates a vault header for a new passphrase: a random salt, the iteration count, and a short
 * verifier encrypted under the derived key. The verifier is what lets a passphrase be checked
 * without the vault body, and is why an empty vault can still be unlocked.
 */
export async function createVaultMeta(
  password: string,
): Promise<{ meta: VaultMeta; derivedKey: CryptoKey }> {
  const salt = crypto.getRandomValues(new Uint8Array(SALT_BYTES));
  const derivedKey = await deriveKey(password, salt);

  const encoder = new TextEncoder();
  const verifierIv = crypto.getRandomValues(new Uint8Array(IV_BYTES));
  const verifierCiphertext = await crypto.subtle.encrypt(
    { name: "AES-GCM", iv: toArrayBuffer(verifierIv) },
    derivedKey,
    encoder.encode(VERIFIER_PLAINTEXT),
  );

  const meta: VaultMeta = {
    version: 1,
    salt: toBase64(salt),
    iterations: PBKDF2_ITERATIONS,
    verifier: toBase64(verifierCiphertext),
    verifierIv: toBase64(verifierIv),
  };

  return { meta, derivedKey };
}

/**
 * Checks a passphrase against a vault header and returns the derived key on success. Rejects on
 * a wrong passphrase, which is the verifier failing to decrypt — the vault body is never
 * touched, so this is safe to call on every unlock attempt.
 */
export async function verifyPassword(
  password: string,
  meta: VaultMeta,
): Promise<CryptoKey> {
  const salt = fromBase64(meta.salt);
  const derivedKey = await deriveKey(password, salt, meta.iterations);

  const verifierIv = fromBase64(meta.verifierIv);
  const verifierCiphertext = fromBase64(meta.verifier);

  const plaintext = await crypto.subtle.decrypt(
    { name: "AES-GCM", iv: toArrayBuffer(verifierIv) },
    derivedKey,
    toArrayBuffer(verifierCiphertext),
  );

  const decoded = new TextDecoder().decode(plaintext);
  if (decoded !== VERIFIER_PLAINTEXT) {
    throw new Error("Vault verifier mismatch");
  }

  return derivedKey;
}
