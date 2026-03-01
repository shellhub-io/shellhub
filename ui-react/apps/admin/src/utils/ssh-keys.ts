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
    const e = err as { name?: string };
    if (e.name === "KeyEncryptedError") {
      return { valid: true, encrypted: true };
    }
    return { valid: false, error: "Invalid private key format." };
  }
}

export function getFingerprint(pem: string, passphrase?: string): string {
  const key = sshpk.parsePrivateKey(pem, "auto", { passphrase });
  return key.fingerprint("md5").toString();
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

  const signer = parsedKey.createSign("sha512");
  signer.update(challenge);
  return signer.sign().toString();
}
