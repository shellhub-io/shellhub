import type {
  VaultMeta,
  VaultData,
  LegacyPrivateKey,
  VaultSettings,
} from "@/types/vault";
import type { IVaultBackend } from "@/utils/vault-backend";
import {
  parseVaultMeta,
  parseVaultData,
  parseVaultSettings,
} from "@/utils/vault-parse";

const VAULT_META_KEY = "shellhub-vault-meta";
const VAULT_DATA_KEY = "shellhub-vault-data";
const VAULT_SETTINGS_KEY = "shellhub-vault-settings";
const LEGACY_KEYS_KEY = "privateKeys";

function prefixKey(base: string, prefix?: string): string {
  return prefix ? `${base}:${prefix}` : base;
}

/**
 * Whether an initialized vault exists in this browser's localStorage for the
 * given scope. Vault meta presence is what defines an initialized vault.
 */
export function localVaultExists(scope?: {
  user: string;
  tenant: string;
}): boolean {
  const prefix = scope ? `${scope.user}:${scope.tenant}` : undefined;
  return localStorage.getItem(prefixKey(VAULT_META_KEY, prefix)) !== null;
}

function safeSetItem(key: string, value: string): void {
  try {
    localStorage.setItem(key, value);
  } catch (err) {
    if (err instanceof DOMException && err.name === "QuotaExceededError") {
      throw new Error(
        "Storage quota exceeded. Free up space or reset the vault.",
        { cause: err },
      );
    }
    throw err;
  }
}

function settle(write: () => void): Promise<void> {
  try {
    write();
    return Promise.resolve();
  } catch (err) {
    return Promise.reject(err instanceof Error ? err : new Error(String(err)));
  }
}

/**
 * Reads the keys written by the pre-vault UI, which stored private keys unencrypted. Returns an
 * empty list on anything unparseable rather than throwing, since this runs on the migration
 * path where failing would strand the user with keys they can no longer see.
 */
export function loadLegacyKeysFromStorage(): LegacyPrivateKey[] {
  let raw: unknown[];
  try {
    raw = JSON.parse(
      localStorage.getItem(LEGACY_KEYS_KEY) ?? "[]",
    ) as unknown[];
  } catch {
    return [];
  }
  if (!Array.isArray(raw)) return [];
  return raw.filter(
    (item): item is LegacyPrivateKey =>
      typeof item === "object" &&
      item !== null &&
      typeof (item as Record<string, unknown>).name === "string" &&
      typeof (item as Record<string, unknown>).data === "string" &&
      typeof (item as Record<string, unknown>).hasPassphrase === "boolean" &&
      typeof (item as Record<string, unknown>).fingerprint === "string",
  );
}

/**
 * Deletes the legacy unencrypted keys. Called only after they have been migrated into the vault:
 * this is the point where the plaintext copies stop existing.
 */
export function clearLegacyKeysFromStorage(): void {
  localStorage.removeItem(LEGACY_KEYS_KEY);
}

/**
 * Keeps the vault in this browser's localStorage. Nothing leaves the machine, so the vault does
 * not follow the user to another browser and is lost with the profile.
 */
export class LocalVaultBackend implements IVaultBackend {
  private readonly prefix: string | undefined;

  /**
   * Scopes every key this backend touches to a user and tenant. Without a scope it reads the
   * unscoped keys, which is what the migration path needs.
   */
  constructor(scope?: { user: string; tenant: string }) {
    this.prefix = scope ? `${scope.user}:${scope.tenant}` : undefined;
  }

  /**
   * Reads the vault header — salt and parameters — or null when no vault exists here.
   */
  loadMeta(): Promise<VaultMeta | null> {
    return Promise.resolve(
      parseVaultMeta(
        localStorage.getItem(prefixKey(VAULT_META_KEY, this.prefix)),
      ),
    );
  }

  /**
   * Writes the vault header. Rejects when storage refuses the write, which is how a full or
   * blocked localStorage reaches the caller instead of silently losing the vault.
   */
  saveMeta(meta: VaultMeta): Promise<void> {
    return settle(() =>
      safeSetItem(prefixKey(VAULT_META_KEY, this.prefix), JSON.stringify(meta)),
    );
  }

  /**
   * Reads the encrypted vault body, or null when there is none.
   */
  loadData(): Promise<VaultData | null> {
    return Promise.resolve(
      parseVaultData(
        localStorage.getItem(prefixKey(VAULT_DATA_KEY, this.prefix)),
      ),
    );
  }

  /**
   * Writes the encrypted vault body, rejecting if storage refuses.
   */
  saveData(data: VaultData): Promise<void> {
    return settle(() =>
      safeSetItem(prefixKey(VAULT_DATA_KEY, this.prefix), JSON.stringify(data)),
    );
  }

  /**
   * Removes the vault entirely — header, body and settings. Irreversible: without the header the
   * body could not be decrypted even if it were recovered.
   */
  clear(): Promise<void> {
    localStorage.removeItem(prefixKey(VAULT_META_KEY, this.prefix));
    localStorage.removeItem(prefixKey(VAULT_DATA_KEY, this.prefix));
    localStorage.removeItem(prefixKey(VAULT_SETTINGS_KEY, this.prefix));
    return Promise.resolve();
  }

  /**
   * Reads the vault settings, falling back to defaults when none are stored.
   */
  loadSettings(): Promise<VaultSettings> {
    return Promise.resolve(
      parseVaultSettings(
        localStorage.getItem(prefixKey(VAULT_SETTINGS_KEY, this.prefix)),
      ),
    );
  }

  /**
   * Writes the vault settings.
   */
  saveSettings(settings: VaultSettings): Promise<void> {
    return settle(() =>
      safeSetItem(
        prefixKey(VAULT_SETTINGS_KEY, this.prefix),
        JSON.stringify(settings),
      ),
    );
  }

  /**
   * Reads the legacy unencrypted keys. Not scoped: they predate scoping, so there is only one set.
   */
  loadLegacyKeys(): Promise<LegacyPrivateKey[]> {
    return Promise.resolve(loadLegacyKeysFromStorage());
  }

  /**
   * Deletes the legacy unencrypted keys.
   */
  clearLegacyKeys(): Promise<void> {
    clearLegacyKeysFromStorage();
    return Promise.resolve();
  }
}
