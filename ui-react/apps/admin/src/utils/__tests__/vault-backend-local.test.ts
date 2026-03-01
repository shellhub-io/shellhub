import { describe, it, expect, beforeEach, vi } from "vitest";
import { LocalVaultBackend } from "../vault-backend-local";
import {
  VAULT_META_KEY,
  VAULT_DATA_KEY,
  LEGACY_KEYS_KEY,
} from "../vault-crypto";
import type { VaultMeta, VaultData, LegacyPrivateKey } from "../../types/vault";

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
    it("returns null when no meta is stored", () => {
      expect(backend.loadMeta()).toBeNull();
    });

    it("saves and loads meta correctly", () => {
      backend.saveMeta(META);
      expect(backend.loadMeta()).toEqual(META);
    });

    it("persists meta under the correct localStorage key", () => {
      backend.saveMeta(META);
      expect(localStorage.getItem(VAULT_META_KEY)).toBe(JSON.stringify(META));
    });

    it("returns null after meta is cleared via clear()", () => {
      backend.saveMeta(META);
      backend.clear();
      expect(backend.loadMeta()).toBeNull();
    });

    it("returns null when stored meta is invalid JSON", () => {
      localStorage.setItem(VAULT_META_KEY, "not-json{{{");
      expect(backend.loadMeta()).toBeNull();
    });

    it("overwrites existing meta on repeated saves", () => {
      backend.saveMeta(META);
      const updated: VaultMeta = { ...META, iterations: 100000 };
      backend.saveMeta(updated);
      expect(backend.loadMeta()).toEqual(updated);
    });
  });

  describe("data", () => {
    it("returns null when no data is stored", () => {
      expect(backend.loadData()).toBeNull();
    });

    it("saves and loads data correctly", () => {
      backend.saveData(DATA);
      expect(backend.loadData()).toEqual(DATA);
    });

    it("persists data under the correct localStorage key", () => {
      backend.saveData(DATA);
      expect(localStorage.getItem(VAULT_DATA_KEY)).toBe(JSON.stringify(DATA));
    });

    it("returns null after data is cleared via clear()", () => {
      backend.saveData(DATA);
      backend.clear();
      expect(backend.loadData()).toBeNull();
    });

    it("returns null when stored data is invalid JSON", () => {
      localStorage.setItem(VAULT_DATA_KEY, "!!!bad json!!!");
      expect(backend.loadData()).toBeNull();
    });

    it("overwrites existing data on repeated saves", () => {
      backend.saveData(DATA);
      const updated: VaultData = { iv: "bmV3SXY=", ciphertext: "bmV3Q2lwaA==" };
      backend.saveData(updated);
      expect(backend.loadData()).toEqual(updated);
    });
  });

  describe("clear", () => {
    it("removes both meta and data from localStorage", () => {
      backend.saveMeta(META);
      backend.saveData(DATA);
      backend.clear();
      expect(localStorage.getItem(VAULT_META_KEY)).toBeNull();
      expect(localStorage.getItem(VAULT_DATA_KEY)).toBeNull();
    });

    it("does not remove legacy keys", () => {
      localStorage.setItem(LEGACY_KEYS_KEY, JSON.stringify(LEGACY_KEYS));
      backend.clear();
      expect(localStorage.getItem(LEGACY_KEYS_KEY)).not.toBeNull();
    });

    it("is idempotent when nothing is stored", () => {
      expect(() => backend.clear()).not.toThrow();
    });
  });

  describe("legacy keys", () => {
    it("returns an empty array when no legacy keys are stored", () => {
      expect(backend.loadLegacyKeys()).toEqual([]);
    });

    it("loads legacy keys stored by another mechanism", () => {
      localStorage.setItem(LEGACY_KEYS_KEY, JSON.stringify(LEGACY_KEYS));
      expect(backend.loadLegacyKeys()).toEqual(LEGACY_KEYS);
    });

    it("returns an empty array when legacy keys storage contains invalid JSON", () => {
      localStorage.setItem(LEGACY_KEYS_KEY, "oops{");
      expect(backend.loadLegacyKeys()).toEqual([]);
    });

    it("removes legacy keys from localStorage on clearLegacyKeys()", () => {
      localStorage.setItem(LEGACY_KEYS_KEY, JSON.stringify(LEGACY_KEYS));
      backend.clearLegacyKeys();
      expect(localStorage.getItem(LEGACY_KEYS_KEY)).toBeNull();
    });

    it("clearLegacyKeys is idempotent when nothing is stored", () => {
      expect(() => backend.clearLegacyKeys()).not.toThrow();
    });
  });

  describe("storage quota exceeded error", () => {
    it("throws a descriptive error when saving meta exceeds quota", () => {
      const quotaError = new DOMException("QuotaExceededError", "QuotaExceededError");
      vi.spyOn(Storage.prototype, "setItem").mockImplementationOnce(() => {
        throw quotaError;
      });
      expect(() => backend.saveMeta(META)).toThrow(
        "Storage quota exceeded. Free up space or reset the vault.",
      );
    });

    it("throws a descriptive error when saving data exceeds quota", () => {
      const quotaError = new DOMException("QuotaExceededError", "QuotaExceededError");
      vi.spyOn(Storage.prototype, "setItem").mockImplementationOnce(() => {
        throw quotaError;
      });
      expect(() => backend.saveData(DATA)).toThrow(
        "Storage quota exceeded. Free up space or reset the vault.",
      );
    });

    it("re-throws non-quota errors as-is", () => {
      const unknownError = new Error("disk failure");
      vi.spyOn(Storage.prototype, "setItem").mockImplementationOnce(() => {
        throw unknownError;
      });
      expect(() => backend.saveMeta(META)).toThrow("disk failure");
    });
  });
});
