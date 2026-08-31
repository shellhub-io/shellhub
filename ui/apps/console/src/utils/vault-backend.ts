import type {
  VaultMeta,
  VaultData,
  LegacyPrivateKey,
  VaultSettings,
} from "@/types/vault";

/**
 * What a vault store has to provide. Header, body and settings are separate so the header can be
 * read — and a passphrase checked — without pulling the encrypted body.
 *
 * Implemented by LocalVaultBackend and ServerVaultBackend; the caller picks one through
 * getVaultBackend and never depends on which.
 */
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
