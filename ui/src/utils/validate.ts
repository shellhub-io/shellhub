/* eslint-disable */
declare const window: any;

import handleError from "@/utils/handleError";
import crypto from "./crypto"

export const validateKey = (keyType: string, key: string, passphrase?: string) => {
  try {
    if (keyType === "private") crypto.parsePrivateKey(key, passphrase);
    else crypto.parseKey(key);
    return true;
  } catch (err) {
    return false;
  }
};

export const convertToFingerprint = (privateKey: string, passphrase?: string) => {
  try {
    return crypto.convertKeyToFingerprint(privateKey, passphrase);
  } catch (error) {
    handleError(error);
    return false;
  }
};

export const parsePrivateKeySsh = (privateKey, passphrase?) => {
  return crypto.parsePrivateKey(privateKey, passphrase);
};

export const parseCertificate = (data: any) => {
  return crypto.parseCertificate(data);
};

export const generateSignature = (privateKey, username) => {
  try {
    return crypto.generateSignature(privateKey, username);
  } catch (err) {
    handleError(err);
    return false;
  }
};

export const validateX509Certificate = (value: string): boolean => {
  try {
    return crypto.validateCertificate(value);
  } catch (err) {
    handleError(err);
    return false;
  }
};

export const createSignatureOfPrivateKey = (
  privateKeyData: any,
  username: Buffer,
) => {
  const signature = crypto.generateRsaKeySignature(privateKeyData, username);
  return signature;
};

export const createKeyFingerprint = (privateKeyData, passphrase?) => {
  const fingerprint = crypto.createKeyFingerprint(privateKeyData, passphrase);
  return fingerprint;
};
