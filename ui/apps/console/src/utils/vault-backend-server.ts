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
    if (response.status === 404) return null;
    if (error || !data)
      throw new Error("Failed to load the vault from the server.");
    this.track(data);
    return data;
  }

  async loadMeta(): Promise<VaultMeta | null> {
    const vault = await this.fetch();
    return parseVaultMeta(vault?.meta);
  }

  async saveMeta(meta: VaultMeta): Promise<void> {
    const { data, error } = await saveVaultMeta({
      body: { meta: JSON.stringify(meta) },
    });
    if (error || !data)
      throw new Error("Failed to save the vault to the server.");
    this.track(data);
  }

  async loadData(): Promise<VaultData | null> {
    const vault = await this.fetch();
    return parseVaultData(vault?.data);
  }

  async saveData(data: VaultData): Promise<void> {
    const res = await saveVaultData({
      body: { data: JSON.stringify(data), version: this.version },
    });
    if (res.response.status === 409) {
      await this.fetch().catch(() => null);
      throw new Error(
        "The vault was changed in another session. Reload the vault and try again.",
      );
    }
    if (res.error || !res.data)
      throw new Error("Failed to save the vault to the server.");
    this.track(res.data);
  }

  async clear(): Promise<void> {
    const { error, response } = await deleteVault();
    if (error && response.status !== 404)
      throw new Error("Failed to reset the vault on the server.");
    versionRegistry.delete(this.key);
  }

  async loadSettings(): Promise<VaultSettings> {
    const vault = await this.fetch().catch(() => null);
    return parseVaultSettings(vault?.settings);
  }

  async saveSettings(settings: VaultSettings): Promise<void> {
    const { data, error } = await saveVaultSettings({
      body: { settings: JSON.stringify(settings) },
    });
    if (error || !data)
      throw new Error("Failed to save the vault settings to the server.");
    this.track(data);
  }

  loadLegacyKeys(): Promise<LegacyPrivateKey[]> {
    return Promise.resolve(loadLegacyKeysFromStorage());
  }

  clearLegacyKeys(): Promise<void> {
    clearLegacyKeysFromStorage();
    return Promise.resolve();
  }
}
