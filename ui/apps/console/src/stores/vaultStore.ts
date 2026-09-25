import { create } from "zustand";
import type {
  VaultStatus,
  VaultKeyEntry,
  VaultMeta,
  LegacyPrivateKey,
  VaultSettings,
} from "@/types/vault";
import { DEFAULT_VAULT_SETTINGS, HIDDEN_GRACE_MS } from "@/types/vault";
import {
  createVaultMeta,
  verifyPassword,
  encrypt,
  decrypt,
  setSessionKey,
  getSessionKey,
  clearSessionKey,
} from "@/utils/vault-crypto";
import {
  getVaultBackend,
  getVaultStorageMode,
  setVaultStorageMode,
  type VaultStorageMode,
} from "@/utils/vault-backend-factory";
import type { IVaultBackend } from "@/utils/vault-backend";
import * as activityTracker from "@/utils/vault-activity-tracker";
import { useAuthStore } from "@/stores/authStore";
import { generateRandomUUID } from "@/utils/random-uuid";
import { nullOnFailure } from "@/utils/failure";

function getScope() {
  const { user, tenant } = useAuthStore.getState();
  return user && tenant ? { user, tenant } : undefined;
}

function getBackend() {
  return getVaultBackend(getScope());
}

/**
 * Which part of an existing key a new one collides with. Both means the same key under the same
 * name, which is a re-add rather than a conflict to resolve.
 */
export type DuplicateField = "name" | "private_key" | "both";

/**
 * Raised when a key being added is already in the vault. It carries the colliding field so the
 * form can mark the input at fault instead of showing a general error.
 */
export class DuplicateKeyError extends Error {
  field: DuplicateField;
  /**
   * Builds the error for a collision on field.
   */
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
  autoLockTimeoutMinutes: number;
  lockOnHidden: boolean;
  autoLockNonce: number;
  storageMode: VaultStorageMode;

  refreshStatus: () => Promise<void>;
  initialize: (
    masterPassword: string,
    mode?: VaultStorageMode,
  ) => Promise<void>;
  unlock: (masterPassword: string) => Promise<void>;
  lock: () => void;
  addKey: (
    entry: Pick<
      VaultKeyEntry,
      "name" | "data" | "hasPassphrase" | "fingerprint" | "algorithm"
    >,
  ) => Promise<void>;
  updateKey: (
    id: string,
    updates: Partial<
      Pick<
        VaultKeyEntry,
        "name" | "data" | "hasPassphrase" | "fingerprint" | "algorithm"
      >
    >,
  ) => Promise<void>;
  removeKey: (id: string) => Promise<void>;
  changeMasterPassword: (
    currentPassword: string,
    newPassword: string,
  ) => Promise<void>;
  resetVault: () => Promise<void>;
  clearError: () => void;
  updateAutoLockSettings: (updates: Partial<VaultSettings>) => Promise<void>;
}

async function persistKeys(
  keys: VaultKeyEntry[],
  be?: IVaultBackend,
  sessionKey?: CryptoKey,
): Promise<void> {
  const key = sessionKey ?? getSessionKey();
  if (!key) throw new Error("Vault is locked");

  const backend = be ?? getBackend();
  const data = await encrypt(key, JSON.stringify(keys));
  await backend.saveData(data);
}

async function ownsUnadoptedVault(
  meta: VaultMeta | null,
  backend: IVaultBackend,
): Promise<boolean> {
  if (!meta) return false;

  const stored = await backend.loadMeta().catch(nullOnFailure);
  if (
    !stored ||
    stored.salt !== meta.salt ||
    stored.verifier !== meta.verifier
  ) {
    return false;
  }

  return backend
    .loadData()
    .then((body) => body === null)
    .catch(() => false);
}

function checkDuplicates(
  existing: VaultKeyEntry[],
  entry: Partial<Pick<VaultKeyEntry, "name" | "fingerprint">>,
  excludeId?: string,
): void {
  const { name, fingerprint } = entry;
  const hasDuplicateName =
    name != null &&
    existing.some((key) => key.id !== excludeId && key.name === name);
  const hasDuplicateKey =
    fingerprint != null &&
    existing.some(
      (key) => key.id !== excludeId && key.fingerprint === fingerprint,
    );
  if (hasDuplicateName && hasDuplicateKey) throw new DuplicateKeyError("both");
  if (hasDuplicateName) throw new DuplicateKeyError("name");
  if (hasDuplicateKey) throw new DuplicateKeyError("private_key");
}

function migrateLegacyKeys(legacy: LegacyPrivateKey[]): VaultKeyEntry[] {
  const now = new Date().toISOString();
  return legacy.map((entry) => ({
    id: generateRandomUUID(),
    name: entry.name,
    data: entry.data,
    hasPassphrase: entry.hasPassphrase,
    fingerprint: entry.fingerprint,
    createdAt: now,
    updatedAt: now,
  }));
}

/**
 * The private-key vault: its lock state, its keys, and the operations that move between them.
 * The decrypted keys live here and nowhere else, and are dropped when the vault locks, so a
 * component must read them through this store rather than holding its own copy.
 */
export const useVaultStore = create<VaultState>((set, get) => {
  let lockGeneration = 0;
  let refreshGeneration = 0;
  let storageGeneration = 0;

  async function loadSettingsIntoState(): Promise<void> {
    const settings = await Promise.resolve()
      .then(() => getBackend().loadSettings())
      .catch(nullOnFailure);

    if (settings) {
      set({
        autoLockTimeoutMinutes: settings.autoLockTimeoutMinutes,
        lockOnHidden: settings.lockOnHidden,
      });
    }
  }

  function startTracker(): void {
    const { autoLockTimeoutMinutes, lockOnHidden } = get();

    activityTracker.start({
      idleTimeoutMs: autoLockTimeoutMinutes * 60000,
      lockOnHidden,
      hiddenGraceMs: HIDDEN_GRACE_MS,
      onIdle: () => {
        if (get().status !== "unlocked") return;
        lockGeneration++;
        clearSessionKey();
        activityTracker.stop();
        set({
          status: "locked",
          keys: [],
          loading: false,
          error: null,
          autoLockNonce: get().autoLockNonce + 1,
        });
      },
    });
  }

  return {
    status: "uninitialized",
    keys: [],
    loading: false,
    error: null,
    autoLockTimeoutMinutes: DEFAULT_VAULT_SETTINGS.autoLockTimeoutMinutes,
    lockOnHidden: DEFAULT_VAULT_SETTINGS.lockOnHidden,
    autoLockNonce: 0,
    storageMode: "local",

    refreshStatus: async () => {
      const generation = ++refreshGeneration;
      const lockedAt = lockGeneration;
      const superseded = () =>
        generation !== refreshGeneration || lockedAt !== lockGeneration;

      let meta;
      try {
        set({ storageMode: getVaultStorageMode(getScope()) });
        const backend = getBackend();
        meta = await backend.loadMeta();
      } catch (err) {
        if (superseded()) return;
        const msg =
          err instanceof Error ? err.message : "Failed to load the vault";
        set({ error: msg });
        return;
      }
      if (superseded()) return;

      if (!meta) {
        activityTracker.stop();
        set({ status: "uninitialized" });
        return;
      }

      await loadSettingsIntoState();
      if (superseded()) return;

      if (!getSessionKey()) {
        activityTracker.stop();
        set({ status: "locked" });
        return;
      }

      set({ status: "unlocked" });
    },

    initialize: async (masterPassword, mode) => {
      const generation = ++lockGeneration;
      const superseded = () => generation !== lockGeneration;

      set({ loading: true, error: null });
      if (mode) {
        setVaultStorageMode(mode, getScope());
        set({ storageMode: mode });
      }
      const backend = getBackend();
      let createdMeta: VaultMeta | null = null;
      try {
        const { meta, derivedKey } = await createVaultMeta(masterPassword);
        await backend.saveMeta(meta);
        createdMeta = meta;

        const legacyKeys = await backend.loadLegacyKeys();
        const keys = legacyKeys.length > 0 ? migrateLegacyKeys(legacyKeys) : [];

        await persistKeys(keys, backend, derivedKey);

        if (legacyKeys.length > 0) {
          await backend.clearLegacyKeys();
        }

        if (superseded()) return;

        setSessionKey(derivedKey);
        set({ status: "unlocked", keys, loading: false });
        await loadSettingsIntoState();
        if (superseded()) return;
        startTracker();
      } catch (err) {
        if (await ownsUnadoptedVault(createdMeta, backend)) {
          await backend
            .clear()
            .then(() => {
              storageGeneration++;
            })
            .catch(() => undefined);
        }
        if (superseded()) return;
        clearSessionKey();
        const msg =
          err instanceof Error ? err.message : "Failed to create vault";
        set({ loading: false, error: msg });
      }
    },

    unlock: async (masterPassword) => {
      const generation = ++lockGeneration;
      const superseded = () => generation !== lockGeneration;

      set({ loading: true, error: null });
      try {
        const backend = getBackend();
        const meta = await backend.loadMeta();
        if (superseded()) return;
        if (!meta) {
          set({ loading: false, error: "No vault found" });
          return;
        }

        let derivedKey: CryptoKey;
        try {
          derivedKey = await verifyPassword(masterPassword, meta);
        } catch {
          if (superseded()) return;
          set({ loading: false, error: "Incorrect master password" });
          return;
        }
        if (superseded()) return;

        setSessionKey(derivedKey);

        try {
          const vaultData = await backend.loadData();
          const parsed: unknown = vaultData
            ? JSON.parse(await decrypt(derivedKey, vaultData))
            : [];
          if (superseded()) return;
          if (!Array.isArray(parsed))
            throw new Error("Vault data is corrupted");
          const isValid = parsed.every(
            (item: unknown) =>
              typeof item === "object" &&
              item !== null &&
              typeof (item as Record<string, unknown>).id === "string" &&
              typeof (item as Record<string, unknown>).name === "string" &&
              typeof (item as Record<string, unknown>).data === "string" &&
              typeof (item as Record<string, unknown>).fingerprint ===
                "string" &&
              typeof (item as Record<string, unknown>).hasPassphrase ===
                "boolean" &&
              typeof (item as Record<string, unknown>).createdAt === "string" &&
              typeof (item as Record<string, unknown>).updatedAt === "string",
          );
          if (!isValid) throw new Error("Vault data is corrupted");
          const keys = parsed as VaultKeyEntry[];

          set({ status: "unlocked", keys, loading: false });
          await loadSettingsIntoState();
          if (superseded()) return;
          startTracker();
        } catch {
          if (superseded()) return;
          clearSessionKey();
          set({ loading: false, error: "Vault data is corrupted" });
        }
      } catch (err) {
        if (superseded()) return;
        clearSessionKey();
        const msg =
          err instanceof Error ? err.message : "Failed to unlock vault";
        set({ loading: false, error: msg });
      }
    },

    lock: () => {
      lockGeneration++;
      clearSessionKey();
      activityTracker.stop();
      set({ status: "locked", keys: [], loading: false, error: null });
    },

    addKey: async (entry) => {
      const existing = get().keys;
      checkDuplicates(existing, entry);

      const now = new Date().toISOString();
      const newKey: VaultKeyEntry = {
        ...entry,
        id: generateRandomUUID(),
        createdAt: now,
        updatedAt: now,
      };
      const keys = [...existing, newKey];
      await persistKeys(keys);
      if (get().status !== "unlocked") return;
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
      if (get().status !== "unlocked") return;
      set({ keys });
    },

    removeKey: async (id) => {
      const existing = get().keys;
      if (!existing.some((k) => k.id === id)) throw new Error("Key not found");
      const keys = existing.filter((key) => key.id !== id);
      await persistKeys(keys);
      if (get().status !== "unlocked") return;
      set({ keys });
    },

    changeMasterPassword: async (currentPassword, newPassword) => {
      const generation = lockGeneration;
      const storage = storageGeneration;
      const superseded = () => generation !== lockGeneration;
      const storageDiscarded = () => storage !== storageGeneration;

      set({ loading: true, error: null });
      try {
        const backend = getBackend();
        const meta = await backend.loadMeta();
        if (!meta) {
          set({ loading: false, error: "No vault found" });
          return;
        }

        if (get().status !== "unlocked") throw new Error("Vault is locked");
        const oldKey = getSessionKey();
        if (!oldKey)
          throw new Error("Session key missing while vault is unlocked");

        try {
          await verifyPassword(currentPassword, meta);
        } catch {
          set({ loading: false, error: "Current password is incorrect" });
          return;
        }

        const { meta: newMeta, derivedKey: newKey } =
          await createVaultMeta(newPassword);

        if (get().status !== "unlocked") {
          set({ loading: false, error: "Vault locked during password change" });
          return;
        }

        const oldData = await backend.loadData();
        const oldMeta = meta;

        if (superseded() || get().status !== "unlocked") {
          set({ loading: false, error: "Vault locked during password change" });
          return;
        }

        setSessionKey(newKey);
        try {
          await persistKeys(get().keys, backend);
          await backend.saveMeta(newMeta);
        } catch (err) {
          if (!superseded() && get().status === "unlocked") {
            setSessionKey(oldKey);
          }
          if (!storageDiscarded()) {
            if (oldData) await backend.saveData(oldData).catch(() => undefined);
            await backend.saveMeta(oldMeta).catch(() => undefined);
          }
          throw err;
        }

        set({ loading: false });
      } catch (err) {
        if (superseded()) return;
        set({
          loading: false,
          error:
            err instanceof Error
              ? err.message
              : "Failed to change master password",
        });
      }
    },

    resetVault: async () => {
      const backend = getBackend();
      try {
        await backend.clear();
      } catch (err) {
        const msg =
          err instanceof Error ? err.message : "Failed to reset the vault";
        set({ error: msg });
        return;
      }
      lockGeneration++;
      storageGeneration++;
      clearSessionKey();
      activityTracker.stop();
      set({
        status: "uninitialized",
        keys: [],
        loading: false,
        error: null,
        autoLockTimeoutMinutes: DEFAULT_VAULT_SETTINGS.autoLockTimeoutMinutes,
        lockOnHidden: DEFAULT_VAULT_SETTINGS.lockOnHidden,
      });
    },

    clearError: () => set({ error: null }),

    updateAutoLockSettings: async (updates) => {
      const current = get();
      const newSettings: VaultSettings = {
        autoLockTimeoutMinutes:
          updates.autoLockTimeoutMinutes ?? current.autoLockTimeoutMinutes,
        lockOnHidden: updates.lockOnHidden ?? current.lockOnHidden,
      };

      const backend = getBackend();
      try {
        await backend.saveSettings(newSettings);
      } catch (err) {
        const msg =
          err instanceof Error ? err.message : "Failed to save vault settings";
        set({ error: msg });
        return;
      }
      set(newSettings);

      if (get().status === "unlocked") {
        startTracker();
      }
    },
  };
});
