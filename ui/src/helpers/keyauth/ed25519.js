const ed = require('@noble/ed25519');
const sshpk = require('sshpk');
const { bytesToHex } = require('@noble/hashes/utils');

const genSignature = async (pkText, data) => {
  const privateKey = sshpk.parsePrivateKey(pkText);
  const message = new TextEncoder().encode(data);
  const signature = await ed.sign(message, bytesToHex(privateKey.part.k.data));
  return Buffer.from(signature).toString('base64');
};

module.exports = { genSignature };
