import { describe, it, expect, beforeEach, vi } from "vitest";
import { LocalVaultBackend, localVaultExists } from "../vault-backend-local";
import type { VaultMeta, VaultData, LegacyPrivateKey } from "@/types/vault";
import { DEFAULT_VAULT_SETTINGS } from "@/types/vault";

const VAULT_META_KEY = "shellhub-vault-meta";
const VAULT_DATA_KEY = "shellhub-vault-data";
const VAULT_SETTINGS_KEY = "shellhub-vault-settings";
const LEGACY_KEYS_KEY = "privateKeys";

const META: VaultMeta = {
  version: 1,
  salt: "c2FsdA==",
  iterations: 600000,
  verifier: "dmVyaWZpZXI=",
  verifierIv: "aXY=",
};

const DATA: VaultData = {
  iv: "aXY=",
  ciphertext: "Y2lwaGVydGV4dA==",
};

const LEGACY_KEYS: LegacyPrivateKey[] = [
  {
    id: 1,
    name: "my-key",
    data: "-----BEGIN RSA PRIVATE KEY-----\n...\n-----END RSA PRIVATE KEY-----",
    hasPassphrase: false,
    fingerprint: "SHA256:abc123",
  },
];

describe("LocalVaultBackend", () => {
  let backend: LocalVaultBackend;

  beforeEach(() => {
    localStorage.clear();
    backend = new LocalVaultBackend();
  });

  describe("meta", () => {
    it("returns null when no meta is stored", async () => {
      expect(await backend.loadMeta()).toBeNull();
    });

    it("saves and loads meta correctly", async () => {
      await backend.saveMeta(META);
      expect(await backend.loadMeta()).toEqual(META);
    });

    it("persists meta under the correct localStorage key", async () => {
      await backend.saveMeta(META);
      expect(localStorage.getItem(VAULT_META_KEY)).toBe(JSON.stringify(META));
    });

    it("returns null after meta is cleared via clear()", async () => {
      await backend.saveMeta(META);
      await backend.clear();
      expect(await backend.loadMeta()).toBeNull();
    });

    it("returns null when stored meta is invalid JSON", async () => {
      localStorage.setItem(VAULT_META_KEY, "not-json{{{");
      expect(await backend.loadMeta()).toBeNull();
    });

    it("overwrites existing meta on repeated saves", async () => {
      await backend.saveMeta(META);
      const updated: VaultMeta = { ...META, iterations: 100000 };
      await backend.saveMeta(updated);
      expect(await backend.loadMeta()).toEqual(updated);
    });

    it("returns null when version is not 1", async () => {
      localStorage.setItem(
        VAULT_META_KEY,
        JSON.stringify({ ...META, version: 2 }),
      );
      expect(await backend.loadMeta()).toBeNull();
    });

    it("returns null when iterations is below minimum", async () => {
      localStorage.setItem(
        VAULT_META_KEY,
        JSON.stringify({ ...META, iterations: 99_999 }),
      );
      expect(await backend.loadMeta()).toBeNull();
    });

    it("returns null when iterations exceeds maximum", async () => {
      localStorage.setItem(
        VAULT_META_KEY,
        JSON.stringify({ ...META, iterations: 10_000_001 }),
      );
      expect(await backend.loadMeta()).toBeNull();
    });

    it("returns null when iterations is not an integer", async () => {
      localStorage.setItem(
        VAULT_META_KEY,
        JSON.stringify({ ...META, iterations: 600000.5 }),
      );
      expect(await backend.loadMeta()).toBeNull();
    });

    it("returns null when salt is not a string", async () => {
      localStorage.setItem(
        VAULT_META_KEY,
        JSON.stringify({ ...META, salt: 42 }),
      );
      expect(await backend.loadMeta()).toBeNull();
    });

    it("returns null when verifier is not a string", async () => {
      localStorage.setItem(
        VAULT_META_KEY,
        JSON.stringify({ ...META, verifier: null }),
      );
      expect(await backend.loadMeta()).toBeNull();
    });
  });

  describe("data", () => {
    it("returns null when no data is stored", async () => {
      expect(await backend.loadData()).toBeNull();
    });

    it("saves and loads data correctly", async () => {
      await backend.saveData(DATA);
      expect(await backend.loadData()).toEqual(DATA);
    });

    it("persists data under the correct localStorage key", async () => {
      await backend.saveData(DATA);
      expect(localStorage.getItem(VAULT_DATA_KEY)).toBe(JSON.stringify(DATA));
    });

    it("returns null after data is cleared via clear()", async () => {
      await backend.saveData(DATA);
      await backend.clear();
      expect(await backend.loadData()).toBeNull();
    });

    it("returns null when stored data is invalid JSON", async () => {
      localStorage.setItem(VAULT_DATA_KEY, "!!!bad json!!!");
      expect(await backend.loadData()).toBeNull();
    });

    it("overwrites existing data on repeated saves", async () => {
      await backend.saveData(DATA);
      const updated: VaultData = { iv: "bmV3SXY=", ciphertext: "bmV3Q2lwaA==" };
      await backend.saveData(updated);
      expect(await backend.loadData()).toEqual(updated);
    });
  });

  describe("clear", () => {
    it("removes both meta and data from localStorage", async () => {
      await backend.saveMeta(META);
      await backend.saveData(DATA);
      await backend.clear();
      expect(localStorage.getItem(VAULT_META_KEY)).toBeNull();
      expect(localStorage.getItem(VAULT_DATA_KEY)).toBeNull();
    });

    it("does not remove legacy keys", async () => {
      localStorage.setItem(LEGACY_KEYS_KEY, JSON.stringify(LEGACY_KEYS));
      await backend.clear();
      expect(localStorage.getItem(LEGACY_KEYS_KEY)).not.toBeNull();
    });

    it("is idempotent when nothing is stored", async () => {
      await expect(backend.clear()).resolves.toBeUndefined();
    });
  });

  describe("legacy keys", () => {
    it("returns an empty array when no legacy keys are stored", async () => {
      expect(await backend.loadLegacyKeys()).toEqual([]);
    });

    it("loads legacy keys stored by another mechanism", async () => {
      localStorage.setItem(LEGACY_KEYS_KEY, JSON.stringify(LEGACY_KEYS));
      expect(await backend.loadLegacyKeys()).toEqual(LEGACY_KEYS);
    });

    it("returns an empty array when legacy keys storage contains invalid JSON", async () => {
      localStorage.setItem(LEGACY_KEYS_KEY, "oops{");
      expect(await backend.loadLegacyKeys()).toEqual([]);
    });

    it("removes legacy keys from localStorage on clearLegacyKeys()", async () => {
      localStorage.setItem(LEGACY_KEYS_KEY, JSON.stringify(LEGACY_KEYS));
      await backend.clearLegacyKeys();
      expect(localStorage.getItem(LEGACY_KEYS_KEY)).toBeNull();
    });

    it("clearLegacyKeys is idempotent when nothing is stored", async () => {
      await expect(backend.clearLegacyKeys()).resolves.toBeUndefined();
    });
  });

  describe("storage quota exceeded error", () => {
    it("throws a descriptive error when saving meta exceeds quota", async () => {
      const quotaError = new DOMException(
        "QuotaExceededError",
        "QuotaExceededError",
      );
      vi.spyOn(Storage.prototype, "setItem").mockImplementationOnce(() => {
        throw quotaError;
      });
      await expect(backend.saveMeta(META)).rejects.toThrow(
        "Storage quota exceeded. Free up space or reset the vault.",
      );
    });

    it("throws a descriptive error when saving data exceeds quota", async () => {
      const quotaError = new DOMException(
        "QuotaExceededError",
        "QuotaExceededError",
      );
      vi.spyOn(Storage.prototype, "setItem").mockImplementationOnce(() => {
        throw quotaError;
      });
      await expect(backend.saveData(DATA)).rejects.toThrow(
        "Storage quota exceeded. Free up space or reset the vault.",
      );
    });

    it("re-throws non-quota errors as-is", async () => {
      const unknownError = new Error("disk failure");
      vi.spyOn(Storage.prototype, "setItem").mockImplementationOnce(() => {
        throw unknownError;
      });
      await expect(backend.saveMeta(META)).rejects.toThrow("disk failure");
    });
  });

  describe("settings", () => {
    it("returns DEFAULT_VAULT_SETTINGS when nothing is stored", async () => {
      expect(await backend.loadSettings()).toEqual(DEFAULT_VAULT_SETTINGS);
    });

    it("save->load round-trip returns the same settings", async () => {
      const settings = { autoLockTimeoutMinutes: 30, lockOnHidden: true };
      await backend.saveSettings(settings);
      expect(await backend.loadSettings()).toEqual(settings);
    });

    it("returns DEFAULT_VAULT_SETTINGS when stored JSON is invalid", async () => {
      localStorage.setItem(VAULT_SETTINGS_KEY, "not-json{{{");
      expect(await backend.loadSettings()).toEqual(DEFAULT_VAULT_SETTINGS);
    });

    it("autoLockTimeoutMinutes:0 round-trips as 0 (valid member, not coerced)", async () => {
      await backend.saveSettings({
        autoLockTimeoutMinutes: 0,
        lockOnHidden: false,
      });
      expect((await backend.loadSettings()).autoLockTimeoutMinutes).toBe(0);
    });

    it("autoLockTimeoutMinutes:7 (not in allowed list) falls back to 15", async () => {
      localStorage.setItem(
        VAULT_SETTINGS_KEY,
        JSON.stringify({ autoLockTimeoutMinutes: 7, lockOnHidden: false }),
      );
      expect((await backend.loadSettings()).autoLockTimeoutMinutes).toBe(15);
    });

    it("autoLockTimeoutMinutes:999999 falls back to 15", async () => {
      localStorage.setItem(
        VAULT_SETTINGS_KEY,
        JSON.stringify({ autoLockTimeoutMinutes: 999999, lockOnHidden: false }),
      );
      expect((await backend.loadSettings()).autoLockTimeoutMinutes).toBe(15);
    });

    it("autoLockTimeoutMinutes:15.5 (float, not a member) falls back to 15", async () => {
      localStorage.setItem(
        VAULT_SETTINGS_KEY,
        JSON.stringify({ autoLockTimeoutMinutes: 15.5, lockOnHidden: false }),
      );
      expect((await backend.loadSettings()).autoLockTimeoutMinutes).toBe(15);
    });

    it("autoLockTimeoutMinutes as string '15' falls back to 15 (no coercion)", async () => {
      localStorage.setItem(
        VAULT_SETTINGS_KEY,
        JSON.stringify({ autoLockTimeoutMinutes: "15", lockOnHidden: false }),
      );
      expect((await backend.loadSettings()).autoLockTimeoutMinutes).toBe(15);
    });

    it("autoLockTimeoutMinutes:null falls back to 15", async () => {
      localStorage.setItem(
        VAULT_SETTINGS_KEY,
        JSON.stringify({ autoLockTimeoutMinutes: null, lockOnHidden: false }),
      );
      expect((await backend.loadSettings()).autoLockTimeoutMinutes).toBe(15);
    });

    it("missing autoLockTimeoutMinutes falls back to 15", async () => {
      localStorage.setItem(
        VAULT_SETTINGS_KEY,
        JSON.stringify({ lockOnHidden: false }),
      );
      expect((await backend.loadSettings()).autoLockTimeoutMinutes).toBe(15);
    });

    it("non-boolean lockOnHidden falls back to false", async () => {
      localStorage.setItem(
        VAULT_SETTINGS_KEY,
        JSON.stringify({ autoLockTimeoutMinutes: 15, lockOnHidden: "yes" }),
      );
      expect((await backend.loadSettings()).lockOnHidden).toBe(false);
    });

    it("clear() removes the settings key", async () => {
      await backend.saveSettings({
        autoLockTimeoutMinutes: 30,
        lockOnHidden: true,
      });
      await backend.clear();
      expect(localStorage.getItem(VAULT_SETTINGS_KEY)).toBeNull();
    });

    it("scoped backend uses prefixed settings key", async () => {
      const scoped = new LocalVaultBackend({ user: "alice", tenant: "t1" });
      await scoped.saveSettings({
        autoLockTimeoutMinutes: 60,
        lockOnHidden: true,
      });
      expect(localStorage.getItem(`${VAULT_SETTINGS_KEY}:alice:t1`)).toBe(
        JSON.stringify({ autoLockTimeoutMinutes: 60, lockOnHidden: true }),
      );
      expect(localStorage.getItem(VAULT_SETTINGS_KEY)).toBeNull();
    });
  });

  describe("localVaultExists", () => {
    it("is false when no vault meta is stored", () => {
      expect(localVaultExists()).toBe(false);
    });

    it("is true after meta is saved", async () => {
      await backend.saveMeta(META);
      expect(localVaultExists()).toBe(true);
    });

    it("respects scoping", async () => {
      const scoped = new LocalVaultBackend({ user: "alice", tenant: "t1" });
      await scoped.saveMeta(META);
      expect(localVaultExists({ user: "alice", tenant: "t1" })).toBe(true);
      expect(localVaultExists({ user: "bob", tenant: "t1" })).toBe(false);
      expect(localVaultExists()).toBe(false);
    });
  });
});
