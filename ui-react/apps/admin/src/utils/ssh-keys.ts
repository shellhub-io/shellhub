import NodeRSA from "node-rsa";
import * as sshpk from "sshpk";
import { Buffer } from "buffer";

export type KeyValidationResult
  = | { valid: true; encrypted: false }
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
  // ShellHub server expects ssh-rsa (PKCS#1 v1.5 SHA-1) for RSA challenge-response
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

  // Choose hash based on key type/curve to match what the server expects
  // (the SSH protocol ties ECDSA curves to specific hash algorithms).
  let hashAlgo: sshpk.AlgorithmHashType = "sha512"; // ed25519 uses sha512
  if (parsedKey.type === "ecdsa") {
    const curve = (parsedKey as { curve?: string }).curve;
    if (curve === "nistp256") hashAlgo = "sha256";
    else if (curve === "nistp384") hashAlgo = "sha384";
    // nistp521 uses sha512 (default)
  }

  const signer = parsedKey.createSign(hashAlgo);
  signer.update(challenge);
  const sig = signer.sign();

  if (parsedKey.type === "ecdsa") {
    // toBuffer('ssh') = [uint32(algLen) || alg || uint32(blobLen) || blob]
    // ssh.Signature.Blob needs just `blob` = mpint(r) || mpint(s)
    const buf = sig.toBuffer("ssh");
    const algLen = buf.readUInt32BE(0);
    const blobLen = buf.readUInt32BE(4 + algLen);
    return buf.subarray(8 + algLen, 8 + algLen + blobLen).toString("base64");
  }

  // ed25519: toBuffer() returns raw 64-byte signature — correct for ssh.Signature.Blob
  return sig.toBuffer().toString("base64");
}
