/* eslint-disable */
export const validateKey = (typeKey: string, value: string) => {
  try {
    let x;
    if (typeKey === "private") {
      x = window.global.parsePrivateKey(value);
    } else {
      x = window.global.parseKey(value);
    }
    return true;
  } catch (err) {
    return false;
  }
};

export const convertToFingerprint = (privateKey: string) => {
  try {
    return window.global.convertKeyToFingerprint(privateKey);
  } catch (err) {
    console.error("err", err);
    return false;
  }
};

export const parsePrivateKeySsh = (privateKey: any) => {
  try {
    return window.global.parsePrivateKey(privateKey);
  } catch (err) {
    console.error("err", err);
    return false;
  }
};

export const createSignerPrivateKey = (privateKey: any, username: string) => {
  try {
    return window.global.createSignerAndUpdate(privateKey, username);
  } catch (err) {
    console.error("err", err);
    return false;
  }
};

export const createSignatureOfPrivateKey = async (
  privateKeyData: any,
  username: string,
) => {
  const signature = await window.global.createSignatureOfPrivateKey(privateKeyData, username);
  return signature;
};

export const createKeyFingerprint = async (privateKeyData: any) => {
  const fingerprint = await window.global.createKeyFingerprint(privateKeyData);
  return fingerprint;
};
