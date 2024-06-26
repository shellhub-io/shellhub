import NodeRSA from "node-rsa";
import * as sshpk from "sshpk";

const createSignatureOfPrivateKey = (
  privateKeyData,
  username,
) => {
  const key = NodeRSA(privateKeyData);
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
  convertKeyToFingerprint,
  createSignerAndUpdate,
};
