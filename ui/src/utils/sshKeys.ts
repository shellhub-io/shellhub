import NodeRSA from "node-rsa";
import * as sshpk from "sshpk";

export const parsePrivateKey = (privateKey: string, passphrase?: string) => sshpk.parsePrivateKey(privateKey, "auto", { passphrase });

export const parsePublicKey = (publicKey: string) => sshpk.parseKey(publicKey, "ssh");

export const isKeyValid = (type: "private" | "public", key: string, passphrase?: string) => {
  try {
    if (type === "private") return !!parsePrivateKey(key, passphrase);
    return !!parsePublicKey(key);
  } catch { return false; }
};

const generateRsaKeySignature = (
  privateKey: string,
  challenge: Buffer,
) => {
  const key = new NodeRSA(privateKey);
  key.setOptions({ signingScheme: "pkcs1-sha1" });
  return key.sign(challenge, "base64");
};

export const generateSignature = (
  privateKey: string,
  challenge: Buffer,
  passphrase?: string,
): string => {
  const parsedPrivateKey = parsePrivateKey(privateKey, passphrase);
  if (parsedPrivateKey.type === "rsa") {
    const decryptedPem = parsedPrivateKey.toString("pem");
    return generateRsaKeySignature(decryptedPem, challenge);
  }

  const signer = parsedPrivateKey.createSign("sha512");
  signer.update(challenge);
  return signer.sign().toString();
};

export const isX509CertificateValid = (certificate: string) => {
  try {
    return !!sshpk.parseCertificate(certificate, "pem");
  } catch { return false; }
};

export const convertToFingerprint = (privateKey: string, passphrase?: string) => (
  sshpk.parsePrivateKey(privateKey, "auto", { passphrase }).fingerprint("md5").toString()
);
