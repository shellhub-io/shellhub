import NodeRSA from "node-rsa";
import * as sshpk from "sshpk";
import { Buffer } from "buffer";

export type KeyValidationResult =
  | { valid: true; encrypted: false }
  | { valid: true; encrypted: true }
  | { valid: false; error: string };

export function validatePrivateKey(pem: string): KeyValidationResult {
  try {
    sshpk.parsePrivateKey(pem, "auto");
    return { valid: true, encrypted: false };
  } catch (err) {
    const e = err instanceof Error ? err : null;
    if (e?.name === "KeyEncryptedError") {
      return { valid: true, encrypted: true };
    }
    return { valid: false, error: "Invalid private key format." };
  }
}

export function getFingerprint(pem: string, passphrase?: string): string {
  const key = sshpk.parsePrivateKey(pem, "auto", { passphrase });
  return key.fingerprint("md5").toString();
}

export function getAlgorithm(pem: string, passphrase?: string): string {
  const key = sshpk.parsePrivateKey(pem, "auto", { passphrase });
  switch (key.type) {
    case "rsa":
      return `RSA ${key.size}`;
    case "ecdsa": {
      const curve = (key as { curve?: string }).curve;
      if (curve === "nistp256") return "ECDSA P-256";
      if (curve === "nistp384") return "ECDSA P-384";
      if (curve === "nistp521") return "ECDSA P-521";
      return `ECDSA ${key.size}`;
    }
    case "ed25519":
      return "Ed25519";
    case "dsa":
      return `DSA ${key.size}`;
    default:
      return key.type.toUpperCase();
  }
}

function generateRsaKeySignature(
  privateKeyPem: string,
  challenge: Buffer,
): string {
  const key = new NodeRSA(privateKeyPem);
  key.setOptions({ signingScheme: "pkcs1-sha1" });
  return key.sign(challenge, "base64");
}

export function generateSignature(
  privateKey: string,
  challenge: Buffer,
  passphrase?: string,
): string {
  const parsedKey = sshpk.parsePrivateKey(privateKey, "auto", { passphrase });
  if (parsedKey.type === "rsa") {
    const decryptedPem = parsedKey.toString("pem");
    return generateRsaKeySignature(decryptedPem, challenge);
  }

  let hashAlgo: sshpk.AlgorithmHashType = "sha512"; // ed25519 uses sha512
  if (parsedKey.type === "ecdsa") {
    const curve = (parsedKey as { curve?: string }).curve;
    if (curve === "nistp256") hashAlgo = "sha256";
    else if (curve === "nistp384") hashAlgo = "sha384";
  }

  const signer = parsedKey.createSign(hashAlgo);
  signer.update(challenge);
  const sig = signer.sign();

  if (parsedKey.type === "ecdsa") {
    const buf = sig.toBuffer("ssh");
    const algLen = buf.readUInt32BE(0);
    const blobLen = buf.readUInt32BE(4 + algLen);
    return buf.subarray(8 + algLen, 8 + algLen + blobLen).toString("base64");
  }

  return sig.toBuffer().toString("base64");
}

function ed25519Blob(raw32: Uint8Array): Uint8Array {
  const type = new TextEncoder().encode("ssh-ed25519");
  const blob = new Uint8Array(4 + type.length + 4 + raw32.length);
  const view = new DataView(blob.buffer);
  let offset = 0;
  view.setUint32(offset, type.length);
  offset += 4;
  blob.set(type, offset);
  offset += type.length;
  view.setUint32(offset, raw32.length);
  offset += 4;
  blob.set(raw32, offset);

  return blob;
}

/** Build an OpenSSH `authorized_keys` line ("ssh-ed25519 <base64>") from a raw
 * 32-byte Ed25519 public key. */
export function ed25519PublicKeyLine(raw32: Uint8Array): string {
  return `ssh-ed25519 ${Buffer.from(ed25519Blob(raw32)).toString("base64")}`;
}

/** Compute the OpenSSH SHA256 fingerprint ("SHA256:<base64 unpadded>") of a raw
 * 32-byte Ed25519 public key, matching the gateway's identity resolution. */
export async function sha256Fingerprint(raw32: Uint8Array): Promise<string> {
  const blob = ed25519Blob(raw32);
  const digest = await crypto.subtle.digest(
    "SHA-256",
    blob.buffer as ArrayBuffer,
  );

  return `SHA256:${Buffer.from(digest).toString("base64").replace(/=+$/, "")}`;
}

const SSH_KEY_PREFIXES = [
  "ssh-rsa",
  "ssh-dss",
  "ssh-ed25519",
  "ecdsa-sha2-nistp256",
  "ecdsa-sha2-nistp384",
  "ecdsa-sha2-nistp521",
];

const PEM_BEGIN = "-----BEGIN";

/**
 * Basic browser-side validation for SSH public keys.
 * Checks for OpenSSH format (ssh-rsa ...) or PEM format (-----BEGIN ...).
 */
export function isPublicKeyValid(key: string): boolean {
  const trimmed = key.trim();
  if (!trimmed) return false;

  for (const prefix of SSH_KEY_PREFIXES) {
    if (trimmed.startsWith(prefix)) {
      const parts = trimmed.split(/\s+/);
      if (parts.length >= 2 && parts[1].length > 10) return true;
    }
  }

  if (trimmed.startsWith(PEM_BEGIN)) {
    return trimmed.includes("-----END");
  }

  return false;
}
