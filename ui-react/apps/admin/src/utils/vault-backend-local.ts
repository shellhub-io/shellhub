import type {
  VaultMeta,
  VaultData,
  LegacyPrivateKey,
} from "@/types/vault";
import type { IVaultBackend } from "@/utils/vault-backend";
import {
  VAULT_META_KEY,
  VAULT_DATA_KEY,
  LEGACY_KEYS_KEY,
} from "@/utils/vault-crypto";

function safeSetItem(key: string, value: string): void {
  try {
    localStorage.setItem(key, value);
  } catch (err) {
    if (err instanceof DOMException && err.name === "QuotaExceededError") {
      throw new Error("Storage quota exceeded. Free up space or reset the vault.");
    }
    throw err;
  }
}

function safeParse<T>(raw: string | null, fallback: T): T {
  if (!raw) return fallback;
  try {
    return JSON.parse(raw) as T;
  } catch {
    return fallback;
  }
}

export class LocalVaultBackend implements IVaultBackend {
  loadMeta(): VaultMeta | null {
    return safeParse<VaultMeta | null>(localStorage.getItem(VAULT_META_KEY), null);
  }

  saveMeta(meta: VaultMeta): void {
    safeSetItem(VAULT_META_KEY, JSON.stringify(meta));
  }

  loadData(): VaultData | null {
    return safeParse<VaultData | null>(localStorage.getItem(VAULT_DATA_KEY), null);
  }

  saveData(data: VaultData): void {
    safeSetItem(VAULT_DATA_KEY, JSON.stringify(data));
  }

  clear(): void {
    localStorage.removeItem(VAULT_META_KEY);
    localStorage.removeItem(VAULT_DATA_KEY);
  }

  loadLegacyKeys(): LegacyPrivateKey[] {
    return safeParse<LegacyPrivateKey[]>(localStorage.getItem(LEGACY_KEYS_KEY), []);
  }

  clearLegacyKeys(): void {
    localStorage.removeItem(LEGACY_KEYS_KEY);
  }
}
