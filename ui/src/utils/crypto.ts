import NodeRSA from "node-rsa";
import * as sshpk from "sshpk";

const generateRsaKeySignature = (
  privateKeyData,
  challenge,
) => {
  const key = new NodeRSA(privateKeyData);
  key.setOptions({ signingScheme: "pkcs1-sha1" });
  const signature = key.sign(challenge, "base64");
  return signature;
};

const generateSignature = (privateKey, challenge) => {
  if (privateKey.type === "rsa") {
    const decryptedPem = privateKey.toString("pem");
    return generateRsaKeySignature(decryptedPem, challenge);
  }

  const signer = privateKey.createSign("sha512");
  signer.update(challenge);
  return signer.sign().toString();
};

const createKeyFingerprint = (privateKeyData: string, passphrase?: string) => {
  const key = sshpk.parsePrivateKey(privateKeyData, "auto", { passphrase });
  const fingerprint = key.fingerprint("md5").toString("hex");
  return fingerprint;
};

const parsePrivateKey = (privateKey, passphrase) => {
  const key = sshpk.parsePrivateKey(privateKey, "auto", { passphrase });
  return key;
};

const parseKey = (key) => {
  const parsedKey = sshpk.parseKey(key);
  return parsedKey;
};

const parseCertificate = (data) => {
  const certBase64 = btoa(data);
  const cert = sshpk.parseCertificate(Buffer.from(certBase64, "base64"), "pem");
  return cert;
};

const validateCertificate = (certificate: string): boolean => {
  try {
    sshpk.parseCertificate(certificate, "pem");
    return true;
  } catch {
    return false;
  }
};

const convertKeyToFingerprint = (privateKey, passphrase) => {
  const fingerprint = sshpk.parsePrivateKey(privateKey, "auto", { passphrase }).fingerprint("md5");
  return fingerprint;
};

export default {
  generateRsaKeySignature,
  createKeyFingerprint,
  parsePrivateKey,
  parseKey,
  parseCertificate,
  validateCertificate,
  convertKeyToFingerprint,
  generateSignature,
};
