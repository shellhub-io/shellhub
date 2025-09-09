import NodeRSA from "node-rsa";
import * as sshpk from "sshpk";
import handleError from "@/utils/handleError";

export const parsePrivateKey = (privateKey: string, passphrase?: string) => sshpk.parsePrivateKey(privateKey, "auto", { passphrase });

export const parsePublicKey = (publicKey: string) => sshpk.parseKey(publicKey);

export const isKeyValid = (type: "private" | "public", key: string, passphrase?: string) => {
  try {
    if (type === "private") return !!parsePrivateKey(key, passphrase);
    return !!parsePublicKey(key);
  } catch (error) { return false; }
};

const generateRsaKeySignature = (
  privateKey: string,
  challenge: Buffer,
) => {
  const key = new NodeRSA(privateKey);
  key.setOptions({ signingScheme: "pkcs1-sha1" });
  return key.sign(challenge, "base64");
};

export const generateSignature = (privateKey: string, challenge: Buffer, passphrase?: string) => {
  try {
    const parsedPrivateKey = parsePrivateKey(privateKey, passphrase);
    if (parsedPrivateKey.type === "rsa") {
      const decryptedPem = parsedPrivateKey.toString("pem");
      return generateRsaKeySignature(decryptedPem, challenge);
    }

    const signer = parsedPrivateKey.createSign("sha512");
    signer.update(challenge);
    return signer.sign().toString();
  } catch (error) {
    handleError(error);
    return false;
  }
};

export const parseCertificate = (certificate: string) => sshpk.parseCertificate(Buffer.from(btoa(certificate), "base64"), "pem");

export const isX509CertificateValid = (certificate: string) => {
  try {
    return !!sshpk.parseCertificate(certificate, "pem");
  } catch (error) { return false; }
};

export const convertToFingerprint = (privateKey: string, passphrase?: string) => (
  sshpk.parsePrivateKey(privateKey, "auto", { passphrase }).fingerprint("md5").toString()
);
