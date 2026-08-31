/**
 * Where the vault stands. Uninitialized is not locked: there is nothing to unlock yet, and the UI
 * offers to create a vault rather than asking for a passphrase.
 */
export type VaultStatus = "uninitialized" | "locked" | "unlocked";

/**
 * The auto-lock intervals offered, in minutes. Zero is "never", and is deliberately in the list
 * rather than expressed as a separate switch.
 */
export const ALLOWED_TIMEOUT_MINUTES = [0, 5, 15, 30, 60] as const;
/**
 * One of the offered auto-lock intervals.
 */
export type AllowedTimeoutMinutes = (typeof ALLOWED_TIMEOUT_MINUTES)[number];

/**
 * The user's vault preferences. Not secret, and readable without unlocking — the vault has to
 * know when to lock before it is open.
 */
export interface VaultSettings {
  autoLockTimeoutMinutes: number;
  lockOnHidden: boolean;
}

/**
 * What a vault with no stored preferences uses: locks after fifteen minutes idle, and stays open
 * when the tab is hidden.
 */
export const DEFAULT_VAULT_SETTINGS: VaultSettings = {
  autoLockTimeoutMinutes: 15,
  lockOnHidden: false,
};

/**
 * How long a hidden tab is tolerated before lock-on-hidden fires. Switching windows hides the
 * tab for a moment, and locking on that would make the vault unusable.
 */
export const HIDDEN_GRACE_MS = 60000;

/**
 * The vault header: the KDF parameters and a verifier for checking a passphrase. Not secret — it
 * is what makes the ciphertext readable to whoever knows the passphrase, and to nobody else.
 */
export interface VaultMeta {
  version: 1;
  salt: string;
  iterations: number;
  verifier: string;
  verifierIv: string;
}

/**
 * The encrypted vault body and the IV it was sealed with.
 */
export interface VaultData {
  iv: string;
  ciphertext: string;
}

/**
 * One key in the vault, as held after decryption. data is the private key itself, so a value of
 * this type must never be logged, persisted outside the vault, or put in a URL.
 */
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

/**
 * A key as the pre-vault UI stored it, unencrypted in localStorage. Read only on the migration
 * path, and deleted once the keys are inside the vault.
 */
export interface LegacyPrivateKey {
  id: number;
  name: string;
  data: string;
  hasPassphrase: boolean;
  fingerprint: string;
}
