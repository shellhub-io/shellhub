import type {
  VaultMeta,
  VaultData,
  LegacyPrivateKey,
  VaultSettings,
} from "@/types/vault";

export interface IVaultBackend {
  loadMeta(): Promise<VaultMeta | null>;
  saveMeta(meta: VaultMeta): Promise<void>;
  loadData(): Promise<VaultData | null>;
  saveData(data: VaultData): Promise<void>;
  clear(): Promise<void>;
  loadLegacyKeys(): Promise<LegacyPrivateKey[]>;
  clearLegacyKeys(): Promise<void>;
  loadSettings(): Promise<VaultSettings>;
  saveSettings(settings: VaultSettings): Promise<void>;
}
