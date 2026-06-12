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

// Runs a synchronous storage write and surfaces any throw as a rejected
// promise, so the async IVaultBackend contract holds (a synchronous throw
// would escape `await` and bypass `.catch`).
function settle(write: () => void): Promise<void> {
  try {
    write();
    return Promise.resolve();
  } catch (err) {
    return Promise.reject(err instanceof Error ? err : new Error(String(err)));
  }
}

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

export function clearLegacyKeysFromStorage(): void {
  localStorage.removeItem(LEGACY_KEYS_KEY);
}

export class LocalVaultBackend implements IVaultBackend {
  private readonly prefix: string | undefined;

  constructor(scope?: { user: string; tenant: string }) {
    this.prefix = scope ? `${scope.user}:${scope.tenant}` : undefined;
  }

  loadMeta(): Promise<VaultMeta | null> {
    return Promise.resolve(
      parseVaultMeta(
        localStorage.getItem(prefixKey(VAULT_META_KEY, this.prefix)),
      ),
    );
  }

  saveMeta(meta: VaultMeta): Promise<void> {
    return settle(() =>
      safeSetItem(prefixKey(VAULT_META_KEY, this.prefix), JSON.stringify(meta)),
    );
  }

  loadData(): Promise<VaultData | null> {
    return Promise.resolve(
      parseVaultData(
        localStorage.getItem(prefixKey(VAULT_DATA_KEY, this.prefix)),
      ),
    );
  }

  saveData(data: VaultData): Promise<void> {
    return settle(() =>
      safeSetItem(prefixKey(VAULT_DATA_KEY, this.prefix), JSON.stringify(data)),
    );
  }

  clear(): Promise<void> {
    localStorage.removeItem(prefixKey(VAULT_META_KEY, this.prefix));
    localStorage.removeItem(prefixKey(VAULT_DATA_KEY, this.prefix));
    localStorage.removeItem(prefixKey(VAULT_SETTINGS_KEY, this.prefix));
    return Promise.resolve();
  }

  loadSettings(): Promise<VaultSettings> {
    return Promise.resolve(
      parseVaultSettings(
        localStorage.getItem(prefixKey(VAULT_SETTINGS_KEY, this.prefix)),
      ),
    );
  }

  saveSettings(settings: VaultSettings): Promise<void> {
    return settle(() =>
      safeSetItem(
        prefixKey(VAULT_SETTINGS_KEY, this.prefix),
        JSON.stringify(settings),
      ),
    );
  }

  loadLegacyKeys(): Promise<LegacyPrivateKey[]> {
    return Promise.resolve(loadLegacyKeysFromStorage());
  }

  clearLegacyKeys(): Promise<void> {
    clearLegacyKeysFromStorage();
    return Promise.resolve();
  }
}
