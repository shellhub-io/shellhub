import { parsePrivateKey, parseKey } from "sshpk";

export const validateKey = (typeKey: string, value: string) => {
  try {
    let x;
    if (typeKey === "private") {
      x = parsePrivateKey(value, "ssh");
    } else {
      x = parseKey(value, "ssh");
    }
    return true;
  } catch (err) {
    return false;
  }
};

export const convertToFingerprint = (privateKey: string) => {
  try {
    // @ts-ignore
    return window.global.convertKeyToFingerprint(privateKey);
  } catch (err) {
    console.error("err", err);
    return false;
  }
};

export const parsePrivateKeySsh = (privateKey: any) => {
  try {
    // @ts-ignore
    return window.global.parsePrivateKey(privateKey);
  } catch (err) {
    console.error("err", err);
    return false;
  }
};

export const createSignerPrivateKey = (privateKey: any, username: string) => {
  try {
    // @ts-ignore
    return window.global.createSignerAndUpdate(privateKey, username);
  } catch (err) {
    console.error("err", err);
    return false;
  }
};

export const createSignatureOfPrivateKey = async (
  privateKeyData: any,
  username: string
) => {
  // @ts-ignore
  let signature = await window.global.createSignatureOfPrivateKey(privateKeyData, username);
  return signature;
};

export const createKeyFingerprint = async (privateKeyData: any) => {
  // @ts-ignore
  let fingerprint = await window.global.createKeyFingerprint(privateKeyData);
  return fingerprint;
};
