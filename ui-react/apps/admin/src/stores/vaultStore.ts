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
import type { IVaultBackend } from "@/utils/vault-backend";
import { useAuthStore } from "@/stores/authStore";

function getBackend() {
  const { user, tenant } = useAuthStore.getState();
  return getVaultBackend(user && tenant ? { user, tenant } : undefined);
}

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
  addKey: (entry: Pick<VaultKeyEntry, "name" | "data" | "hasPassphrase" | "fingerprint" | "algorithm">) => Promise<void>;
  updateKey: (id: string, updates: Partial<Pick<VaultKeyEntry, "name" | "data" | "hasPassphrase" | "fingerprint" | "algorithm">>) => Promise<void>;
  removeKey: (id: string) => Promise<void>;
  changeMasterPassword: (currentPassword: string, newPassword: string) => Promise<void>;
  resetVault: () => void;
  clearError: () => void;
}

async function persistKeys(keys: VaultKeyEntry[], be?: IVaultBackend): Promise<void> {
  const key = getSessionKey();
  if (!key) throw new Error("Vault is locked");

  const backend = be ?? getBackend();
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
    const backend = getBackend();
    const meta = backend.loadMeta();

    if (!meta) {
      set({ status: "uninitialized" });
      return;
    }

    set({ status: getSessionKey() ? "unlocked" : "locked" });
  },

  initialize: async (masterPassword) => {
    set({ loading: true, error: null });
    const backend = getBackend();
    try {
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
      backend.clear();
      clearSessionKey();
      const msg = err instanceof Error ? err.message : "Failed to create vault";
      set({ loading: false, error: msg });
    }
  },

  unlock: async (masterPassword) => {
    set({ loading: true, error: null });
    try {
      const backend = getBackend();
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
        const parsed: unknown = vaultData
          ? JSON.parse(await decrypt(derivedKey, vaultData))
          : [];
        if (!Array.isArray(parsed)) throw new Error("Vault data is corrupted");
        const isValid = parsed.every(
          (item: unknown) =>
            typeof item === "object" &&
            item !== null &&
            typeof (item as Record<string, unknown>).id === "string" &&
            typeof (item as Record<string, unknown>).name === "string" &&
            typeof (item as Record<string, unknown>).data === "string" &&
            typeof (item as Record<string, unknown>).fingerprint === "string" &&
            typeof (item as Record<string, unknown>).hasPassphrase === "boolean" &&
            typeof (item as Record<string, unknown>).createdAt === "string" &&
            typeof (item as Record<string, unknown>).updatedAt === "string",
        );
        if (!isValid) throw new Error("Vault data is corrupted");
        const keys = parsed as VaultKeyEntry[];

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

  // addKey, updateKey, and removeKey intentionally throw raw errors to the caller
  // rather than writing to the store's loading/error fields. Their callers (KeyDrawer,
  // KeyDeleteDialog) manage their own local error UI. This keeps vault-wide loading/error
  // state reserved for operations that affect the whole vault (initialize, unlock, changeMasterPassword).
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
    const existing = get().keys;
    if (!existing.some((k) => k.id === id)) throw new Error("Key not found");
    const keys = existing.filter((key) => key.id !== id);
    await persistKeys(keys);
    set({ keys });
  },

  changeMasterPassword: async (currentPassword, newPassword) => {
    set({ loading: true, error: null });
    try {
      const backend = getBackend();
      const meta = backend.loadMeta();
      if (!meta) {
        set({ loading: false, error: "No vault found" });
        return;
      }

      if (get().status !== "unlocked") throw new Error("Vault is locked");
      const oldKey = getSessionKey();
      if (!oldKey) throw new Error("Session key missing while vault is unlocked");

      try {
        await verifyPassword(currentPassword, meta);
      } catch {
        set({ loading: false, error: "Current password is incorrect" });
        return;
      }

      const { meta: newMeta, derivedKey: newKey } = await createVaultMeta(newPassword);

      // Save current encrypted data and meta for rollback before re-encrypting.
      const oldData = backend.loadData();
      const oldMeta = meta;

      setSessionKey(newKey);
      try {
        await persistKeys(get().keys, backend);
        backend.saveMeta(newMeta);
      } catch (err) {
        // Restore old session key, encrypted data, and meta
        setSessionKey(oldKey);
        if (oldData) backend.saveData(oldData);
        backend.saveMeta(oldMeta);
        throw err;
      }

      set({ loading: false });
    } catch (err) {
      set({
        loading: false,
        error: err instanceof Error ? err.message : "Failed to change master password",
      });
    }
  },

  resetVault: () => {
    const backend = getBackend();
    backend.clear();
    clearSessionKey();
    set({ status: "uninitialized", keys: [], error: null });
  },

  clearError: () => set({ error: null }),
}));
