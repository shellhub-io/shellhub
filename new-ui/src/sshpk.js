const NodeRSA = require("node-rsa");
const sshpk = require('sshpk');

const createSignatureOfPrivateKey = (
  privateKeyData,
  username
) => {
  let signature;
  const key = NodeRSA(privateKeyData);
  key.setOptions({ signingScheme: "pkcs1-sha1" });
  signature = encodeURIComponent(key.sign(username, "base64"));
  return signature;
};

const createKeyFingerprint = (privateKeyData) => {
  const key = sshpk.parsePrivateKey(privateKeyData);
  const fingerprint = key.fingerprint("md5").toString('hex');
  return fingerprint;
};

const parsePrivateKey = (privateKey) => {
  const key = sshpk.parsePrivateKey(privateKey);
  return key;
};

const convertKeyToFingerprint = (privateKey) => {
  const fingerprint = sshpk.parsePrivateKey(privateKey).fingerprint('md5');
  return fingerprint;
};

const createSignerAndUpdate = (privateKey, username) => {
  const signer = privateKey.createSign("sha512");
  signer.update(username);
  return signer.sign().toString();
};

window.global.createSignatureOfPrivateKey = createSignatureOfPrivateKey;
window.global.createKeyFingerprint = createKeyFingerprint;
window.global.convertKeyToFingerprint = convertKeyToFingerprint;
window.global.createSignerAndUpdate = createSignerAndUpdate;
window.global.parsePrivateKey = parsePrivateKey;
