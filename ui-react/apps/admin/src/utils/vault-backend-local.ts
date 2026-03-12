import type {
  VaultMeta,
  VaultData,
  LegacyPrivateKey,
} from "@/types/vault";
import type { IVaultBackend } from "@/utils/vault-backend";

const VAULT_META_KEY = "shellhub-vault-meta";
const VAULT_DATA_KEY = "shellhub-vault-data";
const LEGACY_KEYS_KEY = "privateKeys";

function prefixKey(base: string, prefix?: string): string {
  return prefix ? `${base}:${prefix}` : base;
}

function safeSetItem(key: string, value: string): void {
  try {
    localStorage.setItem(key, value);
  } catch (err) {
    if (err instanceof DOMException && err.name === "QuotaExceededError") {
      throw new Error("Storage quota exceeded. Free up space or reset the vault.", { cause: err });
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
  private readonly prefix: string | undefined;

  constructor(scope?: { user: string; tenant: string }) {
    this.prefix = scope ? `${scope.user}:${scope.tenant}` : undefined;
  }

  loadMeta(): VaultMeta | null {
    const raw = safeParse<Record<string, unknown> | null>(
      localStorage.getItem(prefixKey(VAULT_META_KEY, this.prefix)),
      null,
    );
    if (
      !raw
      || raw.version !== 1
      || typeof raw.salt !== "string"
      || typeof raw.iterations !== "number"
      || !Number.isInteger(raw.iterations)
      || raw.iterations < 100_000
      || raw.iterations > 10_000_000
      || typeof raw.verifier !== "string"
      || typeof raw.verifierIv !== "string"
    )
      return null;
    return raw as unknown as VaultMeta;
  }

  saveMeta(meta: VaultMeta): void {
    safeSetItem(prefixKey(VAULT_META_KEY, this.prefix), JSON.stringify(meta));
  }

  loadData(): VaultData | null {
    const raw = safeParse<Record<string, unknown> | null>(
      localStorage.getItem(prefixKey(VAULT_DATA_KEY, this.prefix)),
      null,
    );
    if (!raw || typeof raw.iv !== "string" || typeof raw.ciphertext !== "string") return null;
    return raw as unknown as VaultData;
  }

  saveData(data: VaultData): void {
    safeSetItem(prefixKey(VAULT_DATA_KEY, this.prefix), JSON.stringify(data));
  }

  clear(): void {
    localStorage.removeItem(prefixKey(VAULT_META_KEY, this.prefix));
    localStorage.removeItem(prefixKey(VAULT_DATA_KEY, this.prefix));
  }

  loadLegacyKeys(): LegacyPrivateKey[] {
    const raw = safeParse<unknown[]>(localStorage.getItem(LEGACY_KEYS_KEY), []);
    if (!Array.isArray(raw)) return [];
    return raw.filter(
      (item): item is LegacyPrivateKey =>
        typeof item === "object"
        && item !== null
        && typeof (item as Record<string, unknown>).name === "string"
        && typeof (item as Record<string, unknown>).data === "string"
        && typeof (item as Record<string, unknown>).hasPassphrase === "boolean"
        && typeof (item as Record<string, unknown>).fingerprint === "string",
    );
  }

  clearLegacyKeys(): void {
    localStorage.removeItem(LEGACY_KEYS_KEY);
  }
}
