import type {
  VaultMeta,
  VaultData,
  LegacyPrivateKey,
} from "@/types/vault";

export interface IVaultBackend {
  loadMeta(): VaultMeta | null;
  saveMeta(meta: VaultMeta): void;
  loadData(): VaultData | null;
  saveData(data: VaultData): void;
  clear(): void;
  loadLegacyKeys(): LegacyPrivateKey[];
  clearLegacyKeys(): void;
}
