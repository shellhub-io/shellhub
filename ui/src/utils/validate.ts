/* eslint-disable */
declare const window: any;

import crypto from "./crypto"

export const validateKey = (typeKey: string, value: string) => {
  try {
    let x;
    if (typeKey === "private") {
      x = crypto.parsePrivateKey(value);
    } else {
      x = crypto.parseKey(value);
    }
    return true;
  } catch (err) {
    return false;
  }
};

export const convertToFingerprint = (privateKey: string) => {
  try {
    return crypto.convertKeyToFingerprint(privateKey);
  } catch (err) {
    console.error("err", err);
    return false;
  }
};

export const parsePrivateKeySsh = (privateKey: any) => {
    return crypto.parsePrivateKey(privateKey);
};

export const createSignerPrivateKey = (privateKey: any, username: string) => {
  try {
    return crypto.createSignerAndUpdate(privateKey, username);
  } catch (err) {
    console.error("err", err);
    return false;
  }
};

export const createSignatureOfPrivateKey = async (
  privateKeyData: any,
  username: string,
) => {
  const signature = await crypto.createSignatureOfPrivateKey(privateKeyData, username);
  return signature;
};

export const createKeyFingerprint = async (privateKeyData: any) => {
  const fingerprint = await crypto.createKeyFingerprint(privateKeyData);
  return fingerprint;
};
