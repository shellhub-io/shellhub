import type { VaultMeta, VaultData } from "@/types/vault";

export const VAULT_META_KEY = "shellhub-vault-meta";
export const VAULT_DATA_KEY = "shellhub-vault-data";
export const LEGACY_KEYS_KEY = "privateKeys";

const VERIFIER_PLAINTEXT = "shellhub-vault-ok";
const PBKDF2_ITERATIONS = 600_000;
const SALT_BYTES = 16;
const IV_BYTES = 12;

function toArrayBuffer(arr: Uint8Array): ArrayBuffer {
  return arr.buffer.slice(arr.byteOffset, arr.byteOffset + arr.byteLength) as ArrayBuffer;
}

let sessionKey: CryptoKey | null = null;

export function setSessionKey(key: CryptoKey): void {
  sessionKey = key;
}

export function getSessionKey(): CryptoKey | null {
  return sessionKey;
}

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
    // AES-GCM appends the 128-bit authentication tag to the ciphertext
    ciphertext: toBase64(ciphertext),
  };
}

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
