import { create } from "zustand";
import type { VaultStatus, VaultKeyEntry, LegacyPrivateKey } from "@/types/vault";
import {
  createVaultMeta,
  verifyPassword,
  encrypt,
  decrypt,
  setSessionKey,
  getSessionKey,
  clearSessionKey,
} from "@/utils/vault-crypto";
import { getVaultBackend } from "@/utils/vault-backend-factory";

export type DuplicateField = "name" | "private_key" | "both";

export class DuplicateKeyError extends Error {
  field: DuplicateField;
  constructor(field: DuplicateField) {
    super(`Duplicate key: ${field}`);
    this.field = field;
  }
}

interface VaultState {
  status: VaultStatus;
  keys: VaultKeyEntry[];
  loading: boolean;
  error: string | null;

  refreshStatus: () => void;
  initialize: (masterPassword: string) => Promise<void>;
  unlock: (masterPassword: string) => Promise<void>;
  lock: () => void;
  addKey: (entry: Pick<VaultKeyEntry, "name" | "data" | "hasPassphrase" | "fingerprint">) => Promise<void>;
  updateKey: (id: string, updates: Partial<Pick<VaultKeyEntry, "name" | "data" | "hasPassphrase" | "fingerprint">>) => Promise<void>;
  removeKey: (id: string) => Promise<void>;
  changeMasterPassword: (currentPassword: string, newPassword: string) => Promise<void>;
  resetVault: () => void;
}

async function persistKeys(keys: VaultKeyEntry[]): Promise<void> {
  const key = getSessionKey();
  if (!key) throw new Error("Vault is locked");

  const backend = getVaultBackend();
  const data = await encrypt(key, JSON.stringify(keys));
  backend.saveData(data);
}

function checkDuplicates(
  existing: VaultKeyEntry[],
  entry: Partial<Pick<VaultKeyEntry, "name" | "fingerprint">>,
  excludeId?: string,
): void {
  const { name, fingerprint } = entry;
  const hasDuplicateName = name != null && existing.some((key) => key.id !== excludeId && key.name === name);
  const hasDuplicateKey = fingerprint != null && existing.some((key) => key.id !== excludeId && key.fingerprint === fingerprint);
  if (hasDuplicateName && hasDuplicateKey) throw new DuplicateKeyError("both");
  if (hasDuplicateName) throw new DuplicateKeyError("name");
  if (hasDuplicateKey) throw new DuplicateKeyError("private_key");
}

function migrateLegacyKeys(legacy: LegacyPrivateKey[]): VaultKeyEntry[] {
  const now = new Date().toISOString();
  return legacy.map((entry) => ({
    id: crypto.randomUUID(),
    name: entry.name,
    data: entry.data,
    hasPassphrase: entry.hasPassphrase,
    fingerprint: entry.fingerprint,
    createdAt: now,
    updatedAt: now,
  }));
}

export const useVaultStore = create<VaultState>((set, get) => ({
  status: "uninitialized",
  keys: [],
  loading: false,
  error: null,

  refreshStatus: () => {
    const backend = getVaultBackend();
    const meta = backend.loadMeta();

    if (!meta) {
      set({ status: "uninitialized" });
      return;
    }

    set({ status: getSessionKey() ? "unlocked" : "locked" });
  },

  initialize: async (masterPassword) => {
    set({ loading: true, error: null });
    try {
      const backend = getVaultBackend();
      const { meta, derivedKey } = await createVaultMeta(masterPassword);
      backend.saveMeta(meta);

      setSessionKey(derivedKey);

      const legacyKeys = backend.loadLegacyKeys();
      const keys = legacyKeys.length > 0
        ? migrateLegacyKeys(legacyKeys)
        : [];

      await persistKeys(keys);

      if (legacyKeys.length > 0) {
        backend.clearLegacyKeys();
      }

      set({ status: "unlocked", keys, loading: false });
    } catch (err) {
      // Rollback saved meta so the vault returns to "uninitialized"
      const backend = getVaultBackend();
      backend.clear();
      clearSessionKey();
      const msg = err instanceof Error ? err.message : "Failed to create vault";
      set({ loading: false, error: msg });
    }
  },

  unlock: async (masterPassword) => {
    set({ loading: true, error: null });
    try {
      const backend = getVaultBackend();
      const meta = backend.loadMeta();
      if (!meta) {
        set({ loading: false, error: "No vault found" });
        return;
      }

      let derivedKey: CryptoKey;
      try {
        derivedKey = await verifyPassword(masterPassword, meta);
      } catch {
        set({ loading: false, error: "Incorrect master password" });
        return;
      }

      setSessionKey(derivedKey);

      try {
        const vaultData = backend.loadData();
        const keys: VaultKeyEntry[] = vaultData
          ? JSON.parse(await decrypt(derivedKey, vaultData))
          : [];

        set({ status: "unlocked", keys, loading: false });
      } catch {
        clearSessionKey();
        set({ loading: false, error: "Vault data is corrupted" });
      }
    } catch (err) {
      clearSessionKey();
      const msg = err instanceof Error ? err.message : "Failed to unlock vault";
      set({ loading: false, error: msg });
    }
  },

  lock: () => {
    clearSessionKey();
    set({ status: "locked", keys: [], error: null });
  },

  addKey: async (entry) => {
    const existing = get().keys;
    checkDuplicates(existing, entry);

    const now = new Date().toISOString();
    const newKey: VaultKeyEntry = {
      ...entry,
      id: crypto.randomUUID(),
      createdAt: now,
      updatedAt: now,
    };
    const keys = [...existing, newKey];
    await persistKeys(keys);
    set({ keys });
  },

  updateKey: async (id, updates) => {
    const existing = get().keys;
    if (!existing.some((k) => k.id === id)) throw new Error("Key not found");
    checkDuplicates(existing, updates, id);

    const keys = existing.map((key) =>
      key.id === id
        ? { ...key, ...updates, updatedAt: new Date().toISOString() }
        : key,
    );
    await persistKeys(keys);
    set({ keys });
  },

  removeKey: async (id) => {
    const keys = get().keys.filter((key) => key.id !== id);
    await persistKeys(keys);
    set({ keys });
  },

  changeMasterPassword: async (currentPassword, newPassword) => {
    set({ loading: true, error: null });
    try {
      const backend = getVaultBackend();
      const meta = backend.loadMeta();
      if (!meta) {
        set({ loading: false, error: "No vault found" });
        return;
      }

      if (get().status !== "unlocked") throw new Error("Vault is locked");
      const oldKey = getSessionKey()!;

      await verifyPassword(currentPassword, meta);

      const { meta: newMeta, derivedKey: newKey } = await createVaultMeta(newPassword);

      // Save current encrypted data for rollback before re-encrypting.
      const oldData = backend.loadData();

      setSessionKey(newKey);
      try {
        await persistKeys(get().keys);
        backend.saveMeta(newMeta);
      } catch (err) {
        // Restore old session key and old encrypted data (no re-encryption needed)
        setSessionKey(oldKey);
        if (oldData) backend.saveData(oldData);
        throw err;
      }

      set({ loading: false });
    } catch (err) {
      const msg = err instanceof Error ? err.message : "";
      const isWrongPassword = msg === "Vault verifier mismatch";
      set({
        loading: false,
        error: isWrongPassword
          ? "Current password is incorrect"
          : "Failed to change master password",
      });
    }
  },

  resetVault: () => {
    const backend = getVaultBackend();
    backend.clear();
    clearSessionKey();
    set({ status: "uninitialized", keys: [], error: null });
  },
}));
