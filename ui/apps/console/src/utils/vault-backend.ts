import type {
  VaultMeta,
  VaultData,
  LegacyPrivateKey,
  VaultSettings,
} from "@/types/vault";

export interface IVaultBackend {
  loadMeta(): VaultMeta | null;
  saveMeta(meta: VaultMeta): void;
  loadData(): VaultData | null;
  saveData(data: VaultData): void;
  clear(): void;
  loadLegacyKeys(): LegacyPrivateKey[];
  clearLegacyKeys(): void;
  loadSettings(): VaultSettings;
  saveSettings(settings: VaultSettings): void;
}
