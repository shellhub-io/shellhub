import {
  getVault,
  saveVaultMeta,
  saveVaultData,
  saveVaultSettings,
  deleteVault,
} from "@/client";
import type { VaultResponse } from "@/client";
import type {
  VaultMeta,
  VaultData,
  LegacyPrivateKey,
  VaultSettings,
} from "@/types/vault";
import type { IVaultBackend } from "@/utils/vault-backend";
import {
  loadLegacyKeysFromStorage,
  clearLegacyKeysFromStorage,
} from "@/utils/vault-backend-local";
import {
  parseVaultMeta,
  parseVaultData,
  parseVaultSettings,
} from "@/utils/vault-parse";
import type { VaultScope } from "@/utils/vault-backend-factory";

const versionRegistry = new Map<string, number>();

function scopeKey(scope?: VaultScope): string {
  return scope ? `${scope.user}:${scope.tenant}` : "";
}

/**
 * Vault backend that stores the encrypted vault on the ShellHub server
 * (Cloud/Enterprise). The server is the single source of truth: every load
 * hits the API and every save writes through. Encryption stays in the
 * browser — the server only ever sees opaque strings.
 *
 * Writes to vault data carry the last seen `version` so a concurrent write
 * from another session is detected by the server (409) instead of silently
 * overwritten.
 */
export class ServerVaultBackend implements IVaultBackend {
  private readonly key: string;

  /**
   * Binds this backend to a scope, which is also the key its version counter is tracked under.
   */
  constructor(scope?: VaultScope) {
    this.key = scopeKey(scope);
  }

  private get version(): number {
    return versionRegistry.get(this.key) ?? 0;
  }

  private track(vault: { version?: number } | undefined): void {
    if (typeof vault?.version === "number")
      versionRegistry.set(this.key, vault.version);
  }

  private async fetch(): Promise<VaultResponse | null> {
    const { data, error, response } = await getVault();
    if (response?.status === 404) return null;
    if (error || !data)
      throw new Error("Failed to load the vault from the server.");
    this.track(data);
    return data;
  }

  /**
   * Fetches the vault header from the server, or null when the account has no vault yet.
   */
  async loadMeta(): Promise<VaultMeta | null> {
    const vault = await this.fetch();
    return parseVaultMeta(vault?.meta);
  }

  /**
   * Stores the vault header. Throws with a message fit to show the user, since a failure here
   * means the vault they just created was not saved.
   */
  async saveMeta(meta: VaultMeta): Promise<void> {
    const { data, error } = await saveVaultMeta({
      body: { meta: JSON.stringify(meta) },
    });
    if (error || !data)
      throw new Error("Failed to save the vault to the server.");
    this.track(data);
  }

  /**
   * Fetches the encrypted vault body, or null when there is none.
   */
  async loadData(): Promise<VaultData | null> {
    const vault = await this.fetch();
    return parseVaultData(vault?.data);
  }

  /**
   * Stores the encrypted vault body, guarded by the version last read. A 409 means another session
   * wrote first: the local copy is refreshed and the caller is told to reload rather than
   * overwrite, because a blind write would silently drop the other session's keys.
   */
  async saveData(data: VaultData): Promise<void> {
    const res = await saveVaultData({
      body: { data: JSON.stringify(data), version: this.version },
    });
    if (res.response?.status === 409) {
      await this.fetch().catch(() => null);
      throw new Error(
        "The vault was changed in another session. Reload the vault and try again.",
      );
    }
    if (res.error || !res.data)
      throw new Error("Failed to save the vault to the server.");
    this.track(res.data);
  }

  /**
   * Deletes the vault on the server. A 404 counts as success — the vault is gone either way — and
   * the version counter is dropped so a later write does not carry a stale one.
   */
  async clear(): Promise<void> {
    const { error, response } = await deleteVault();
    if (error && response?.status !== 404)
      throw new Error("Failed to reset the vault on the server.");
    versionRegistry.delete(this.key);
  }

  /**
   * Fetches the vault settings, falling back to defaults if the vault cannot be read. Settings are
   * not secret, and failing here would block unlocking over a preference.
   */
  async loadSettings(): Promise<VaultSettings> {
    const vault = await this.fetch().catch(() => null);
    return parseVaultSettings(vault?.settings);
  }

  /**
   * Stores the vault settings.
   */
  async saveSettings(settings: VaultSettings): Promise<void> {
    const { data, error } = await saveVaultSettings({
      body: { settings: JSON.stringify(settings) },
    });
    if (error || !data)
      throw new Error("Failed to save the vault settings to the server.");
    this.track(data);
  }

  /**
   * Reads the legacy unencrypted keys. They only ever existed in this browser, so even the server
   * backend reads them locally.
   */
  loadLegacyKeys(): Promise<LegacyPrivateKey[]> {
    return Promise.resolve(loadLegacyKeysFromStorage());
  }

  /**
   * Deletes the legacy unencrypted keys from this browser.
   */
  clearLegacyKeys(): Promise<void> {
    clearLegacyKeysFromStorage();
    return Promise.resolve();
  }
}
