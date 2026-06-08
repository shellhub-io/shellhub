export type VaultStatus = "uninitialized" | "locked" | "unlocked";

export const ALLOWED_TIMEOUT_MINUTES = [0, 5, 15, 30, 60] as const;
export type AllowedTimeoutMinutes = (typeof ALLOWED_TIMEOUT_MINUTES)[number];

export interface VaultSettings {
  autoLockTimeoutMinutes: number;
  lockOnHidden: boolean;
}

export const DEFAULT_VAULT_SETTINGS: VaultSettings = {
  autoLockTimeoutMinutes: 15,
  lockOnHidden: false,
};

export const HIDDEN_GRACE_MS = 60000;

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
  algorithm?: string;
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
