import NodeRSA from "node-rsa";
import * as sshpk from "sshpk";

const createSignatureOfPrivateKey = (
  privateKeyData,
  username,
) => {
  const key = new NodeRSA(privateKeyData);
  key.setOptions({ signingScheme: "pkcs1-sha1" });
  const signature = encodeURIComponent(key.sign(username, "base64"));
  return signature;
};

const createKeyFingerprint = (privateKeyData) => {
  const key = sshpk.parsePrivateKey(privateKeyData);
  const fingerprint = key.fingerprint("md5").toString("hex");
  return fingerprint;
};

const parsePrivateKey = (privateKey) => {
  const key = sshpk.parsePrivateKey(privateKey);
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

const convertKeyToFingerprint = (privateKey) => {
  const fingerprint = sshpk.parsePrivateKey(privateKey).fingerprint("md5");
  return fingerprint;
};

const createSignerAndUpdate = (privateKey, username) => {
  const signer = privateKey.createSign("sha512");
  signer.update(username);
  return signer.sign().toString();
};

export default {
  createSignatureOfPrivateKey,
  createKeyFingerprint,
  parsePrivateKey,
  parseKey,
  parseCertificate,
  convertKeyToFingerprint,
  createSignerAndUpdate,
};
