export type VaultStatus = "uninitialized" | "locked" | "unlocked";

export interface VaultMeta {
  version: 1;
  salt: string;
  iterations: number;
  verifier: string;
  verifierIv: string;
}

export interface VaultData {
  iv: string;
  ciphertext: string;
}

export interface VaultKeyEntry {
  id: string;
  name: string;
  data: string;
  hasPassphrase: boolean;
  fingerprint: string;
  createdAt: string;
  updatedAt: string;
}

export interface LegacyPrivateKey {
  id: number;
  name: string;
  data: string;
  hasPassphrase: boolean;
  fingerprint: string;
}
