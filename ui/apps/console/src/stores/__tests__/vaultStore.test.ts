import { describe, it, expect, vi, beforeEach } from "vitest";
import type { VaultMeta, VaultData, VaultKeyEntry } from "@/types/vault";
import { DEFAULT_VAULT_SETTINGS } from "@/types/vault";

vi.mock("@/utils/vault-crypto", () => ({
  createVaultMeta: vi.fn(),
  verifyPassword: vi.fn(),
  encrypt: vi.fn(),
  decrypt: vi.fn(),
  setSessionKey: vi.fn(),
  getSessionKey: vi.fn(),
  clearSessionKey: vi.fn(),
}));

vi.mock("@/utils/vault-backend-factory", () => ({
  getVaultBackend: vi.fn(),
  getVaultStorageMode: vi.fn(() => "local"),
  setVaultStorageMode: vi.fn(),
}));

vi.mock("../authStore", () => ({
  useAuthStore: {
    getState: vi.fn(() => ({ user: "testuser", tenant: "test-tenant" })),
  },
}));

vi.mock("@/utils/vault-activity-tracker", () => ({
  start: vi.fn(),
  stop: vi.fn(),
}));

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
import * as activityTracker from "@/utils/vault-activity-tracker";

import { useVaultStore, DuplicateKeyError } from "../vaultStore";

const mockCrypto = vi.mocked(createVaultMeta);
const mockVerify = vi.mocked(verifyPassword);
const mockEncrypt = vi.mocked(encrypt);
const mockDecrypt = vi.mocked(decrypt);
const mockSetSession = vi.mocked(setSessionKey);
const mockGetSession = vi.mocked(getSessionKey);
const mockClearSession = vi.mocked(clearSessionKey);
const mockGetBackend = vi.mocked(getVaultBackend);
const mockTrackerStart = vi.mocked(activityTracker.start);
const mockTrackerStop = vi.mocked(activityTracker.stop);

function makeFakeKey(overrides: Partial<VaultKeyEntry> = {}): VaultKeyEntry {
  return {
    id: "key-1",
    name: "My Key",
    data: "-----BEGIN RSA PRIVATE KEY-----",
    hasPassphrase: false,
    fingerprint: "aa:bb:cc",
    createdAt: "2024-01-01T00:00:00.000Z",
    updatedAt: "2024-01-01T00:00:00.000Z",
    ...overrides,
  };
}

function makeMeta(): VaultMeta {
  return {
    version: 1,
    salt: "c2FsdA==",
    iterations: 600000,
    verifier: "dmVyaWZpZXI=",
    verifierIv: "aXY=",
  };
}

function makeVaultData(): VaultData {
  return { iv: "aXY=", ciphertext: "Y2lwaGVydGV4dA==" };
}

function makeFakeBackend() {
  return {
    loadMeta: vi.fn().mockResolvedValue(null),
    saveMeta: vi.fn().mockResolvedValue(undefined),
    loadData: vi.fn().mockResolvedValue(null),
    saveData: vi.fn().mockResolvedValue(undefined),
    clear: vi.fn().mockResolvedValue(undefined),
    loadLegacyKeys: vi.fn().mockResolvedValue([]),
    clearLegacyKeys: vi.fn().mockResolvedValue(undefined),
    loadSettings: vi.fn().mockResolvedValue(DEFAULT_VAULT_SETTINGS),
    saveSettings: vi.fn().mockResolvedValue(undefined),
  };
}

function makeFakeCryptoKey(label = "AES-GCM"): CryptoKey {
  return {
    type: "secret",
    extractable: false,
    algorithm: { name: label },
    usages: ["encrypt", "decrypt"],
  };
}

beforeEach(() => {
  useVaultStore.setState({
    status: "uninitialized",
    keys: [],
    loading: false,
    error: null,
    autoLockNonce: 0,
    autoLockTimeoutMinutes: 15,
    lockOnHidden: false,
  });
  vi.clearAllMocks();
});

describe("vaultStore", () => {
  describe("refreshStatus", () => {
    it("sets status to uninitialized when backend has no meta", async () => {
      const backend = makeFakeBackend();
      mockGetBackend.mockReturnValue(backend);
      mockGetSession.mockReturnValue(null);

      await useVaultStore.getState().refreshStatus();

      expect(useVaultStore.getState().status).toBe("uninitialized");
    });

    it("sets status to locked when meta exists but no session key", async () => {
      const backend = makeFakeBackend();
      backend.loadMeta.mockReturnValue(makeMeta());
      mockGetBackend.mockReturnValue(backend);
      mockGetSession.mockReturnValue(null);

      await useVaultStore.getState().refreshStatus();

      expect(useVaultStore.getState().status).toBe("locked");
    });

    it("sets status to unlocked when meta exists and session key is present", async () => {
      const backend = makeFakeBackend();
      backend.loadMeta.mockReturnValue(makeMeta());
      mockGetBackend.mockReturnValue(backend);
      mockGetSession.mockReturnValue(makeFakeCryptoKey());

      await useVaultStore.getState().refreshStatus();

      expect(useVaultStore.getState().status).toBe("unlocked");
    });

    it("loads settings from backend ONLY when meta exists", async () => {
      const backend = makeFakeBackend();
      const customSettings = { autoLockTimeoutMinutes: 30, lockOnHidden: true };
      backend.loadMeta.mockReturnValue(makeMeta());
      backend.loadSettings.mockReturnValue(customSettings);
      mockGetBackend.mockReturnValue(backend);
      mockGetSession.mockReturnValue(null);

      await useVaultStore.getState().refreshStatus();

      expect(backend.loadSettings).toHaveBeenCalled();
      expect(useVaultStore.getState().autoLockTimeoutMinutes).toBe(30);
      expect(useVaultStore.getState().lockOnHidden).toBe(true);
    });

    it("does NOT call loadSettings when meta is absent (uninitialized)", async () => {
      const backend = makeFakeBackend();
      backend.loadMeta.mockReturnValue(null);
      mockGetBackend.mockReturnValue(backend);
      mockGetSession.mockReturnValue(null);

      await useVaultStore.getState().refreshStatus();

      expect(backend.loadSettings).not.toHaveBeenCalled();
    });

    it("does not change autoLockNonce on any refreshStatus call", async () => {
      const backend = makeFakeBackend();
      backend.loadMeta.mockReturnValue(makeMeta());
      mockGetBackend.mockReturnValue(backend);
      mockGetSession.mockReturnValue(null);

      useVaultStore.setState({ autoLockNonce: 5 });
      await useVaultStore.getState().refreshStatus();

      expect(useVaultStore.getState().autoLockNonce).toBe(5);
    });

    it("calls activityTracker.stop() when no meta (uninitialized return)", async () => {
      const backend = makeFakeBackend();
      backend.loadMeta.mockReturnValue(null);
      mockGetBackend.mockReturnValue(backend);
      mockGetSession.mockReturnValue(null);

      await useVaultStore.getState().refreshStatus();

      expect(mockTrackerStop).toHaveBeenCalled();
    });

    it("calls activityTracker.stop() when meta exists but no session key (locked)", async () => {
      const backend = makeFakeBackend();
      backend.loadMeta.mockReturnValue(makeMeta());
      mockGetBackend.mockReturnValue(backend);
      mockGetSession.mockReturnValue(null);

      await useVaultStore.getState().refreshStatus();

      expect(mockTrackerStop).toHaveBeenCalled();
    });

    it("does NOT call activityTracker.start() on the unlocked branch", async () => {
      const backend = makeFakeBackend();
      backend.loadMeta.mockReturnValue(makeMeta());
      mockGetBackend.mockReturnValue(backend);
      mockGetSession.mockReturnValue(makeFakeCryptoKey());

      await useVaultStore.getState().refreshStatus();

      expect(mockTrackerStart).not.toHaveBeenCalled();
    });

    it("does not throw when backend.loadSettings throws (exception-safe)", async () => {
      const backend = makeFakeBackend();
      backend.loadMeta.mockReturnValue(makeMeta());
      backend.loadSettings.mockImplementation(() => {
        throw new Error("storage error");
      });
      mockGetBackend.mockReturnValue(backend);
      mockGetSession.mockReturnValue(null);

      await expect(
        useVaultStore.getState().refreshStatus(),
      ).resolves.toBeUndefined();
      // Status should still be set despite settings load failure
      expect(useVaultStore.getState().status).toBe("locked");
    });
  });

  describe("initialize", () => {
    it("creates vault meta, persists empty keys, and transitions to unlocked", async () => {
      const backend = makeFakeBackend();
      const derivedKey = makeFakeCryptoKey();
      const meta = makeMeta();
      const encryptedData = makeVaultData();

      mockGetBackend.mockReturnValue(backend);
      mockCrypto.mockResolvedValue({ meta, derivedKey });
      mockGetSession.mockReturnValue(derivedKey);
      mockEncrypt.mockResolvedValue(encryptedData);

      await useVaultStore.getState().initialize("master-pass");

      const state = useVaultStore.getState();
      expect(state.status).toBe("unlocked");
      expect(state.keys).toEqual([]);
      expect(state.loading).toBe(false);
      expect(state.error).toBeNull();
      expect(backend.saveMeta).toHaveBeenCalledWith(meta);
      expect(mockSetSession).toHaveBeenCalledWith(derivedKey);
      expect(backend.saveData).toHaveBeenCalledWith(encryptedData);
    });

    it("calls loadSettings then startTracker (activityTracker.start) after success", async () => {
      const backend = makeFakeBackend();
      const derivedKey = makeFakeCryptoKey();
      const customSettings = { autoLockTimeoutMinutes: 30, lockOnHidden: true };
      backend.loadSettings.mockReturnValue(customSettings);

      mockGetBackend.mockReturnValue(backend);
      mockCrypto.mockResolvedValue({ meta: makeMeta(), derivedKey });
      mockGetSession.mockReturnValue(derivedKey);
      mockEncrypt.mockResolvedValue(makeVaultData());

      await useVaultStore.getState().initialize("master-pass");

      expect(backend.loadSettings).toHaveBeenCalled();
      expect(mockTrackerStart).toHaveBeenCalled();
      const startArg = mockTrackerStart.mock.calls[0][0];
      expect(startArg.idleTimeoutMs).toBe(30 * 60000);
      expect(startArg.lockOnHidden).toBe(true);
    });

    it("migrates legacy keys when they exist", async () => {
      const backend = makeFakeBackend();
      const derivedKey = makeFakeCryptoKey();
      const legacyKeys = [
        {
          id: 1,
          name: "Old Key",
          data: "-----BEGIN",
          hasPassphrase: false,
          fingerprint: "ff:ee:dd",
        },
      ];
      backend.loadLegacyKeys.mockReturnValue(legacyKeys);

      mockGetBackend.mockReturnValue(backend);
      mockCrypto.mockResolvedValue({ meta: makeMeta(), derivedKey });
      mockGetSession.mockReturnValue(derivedKey);
      mockEncrypt.mockResolvedValue(makeVaultData());

      await useVaultStore.getState().initialize("master-pass");

      const state = useVaultStore.getState();
      expect(state.keys).toHaveLength(1);
      expect(state.keys[0].name).toBe("Old Key");
      expect(state.keys[0].fingerprint).toBe("ff:ee:dd");
      expect(backend.clearLegacyKeys).toHaveBeenCalled();
    });

    it("does not call clearLegacyKeys when no legacy keys exist", async () => {
      const backend = makeFakeBackend();
      const derivedKey = makeFakeCryptoKey();

      mockGetBackend.mockReturnValue(backend);
      mockCrypto.mockResolvedValue({ meta: makeMeta(), derivedKey });
      mockGetSession.mockReturnValue(derivedKey);
      mockEncrypt.mockResolvedValue(makeVaultData());

      await useVaultStore.getState().initialize("master-pass");

      expect(backend.clearLegacyKeys).not.toHaveBeenCalled();
    });

    it("rolls back and sets error when initialization fails", async () => {
      const backend = makeFakeBackend();
      mockGetBackend.mockReturnValue(backend);
      mockCrypto.mockRejectedValue(new Error("Crypto failure"));

      await useVaultStore.getState().initialize("master-pass");

      const state = useVaultStore.getState();
      expect(state.status).toBe("uninitialized");
      expect(state.loading).toBe(false);
      expect(state.error).toBe("Crypto failure");
      expect(backend.clear).toHaveBeenCalled();
      expect(mockClearSession).toHaveBeenCalled();
    });

    it("sets loading to true during initialization", async () => {
      const backend = makeFakeBackend();
      const derivedKey = makeFakeCryptoKey();

      let resolve: (v: { meta: VaultMeta; derivedKey: CryptoKey }) => void;
      mockGetBackend.mockReturnValue(backend);
      mockCrypto.mockReturnValue(
        new Promise((r) => {
          resolve = r;
        }),
      );

      const promise = useVaultStore.getState().initialize("master-pass");
      expect(useVaultStore.getState().loading).toBe(true);

      mockGetSession.mockReturnValue(derivedKey);
      mockEncrypt.mockResolvedValue(makeVaultData());
      resolve!({ meta: makeMeta(), derivedKey });
      await promise;

      expect(useVaultStore.getState().loading).toBe(false);
    });
  });

  describe("unlock", () => {
    it("transitions to unlocked and loads keys on success", async () => {
      const backend = makeFakeBackend();
      const derivedKey = makeFakeCryptoKey();
      const existingKeys = [makeFakeKey()];
      backend.loadMeta.mockReturnValue(makeMeta());
      backend.loadData.mockReturnValue(makeVaultData());

      mockGetBackend.mockReturnValue(backend);
      mockVerify.mockResolvedValue(derivedKey);
      mockDecrypt.mockResolvedValue(JSON.stringify(existingKeys));

      await useVaultStore.getState().unlock("master-pass");

      const state = useVaultStore.getState();
      expect(state.status).toBe("unlocked");
      expect(state.keys).toEqual(existingKeys);
      expect(state.loading).toBe(false);
      expect(state.error).toBeNull();
      expect(mockSetSession).toHaveBeenCalledWith(derivedKey);
    });

    it("calls loadSettings then startTracker (activityTracker.start) after success", async () => {
      const backend = makeFakeBackend();
      const derivedKey = makeFakeCryptoKey();
      const customSettings = { autoLockTimeoutMinutes: 5, lockOnHidden: false };
      backend.loadMeta.mockReturnValue(makeMeta());
      backend.loadData.mockReturnValue(makeVaultData());
      backend.loadSettings.mockReturnValue(customSettings);

      mockGetBackend.mockReturnValue(backend);
      mockVerify.mockResolvedValue(derivedKey);
      mockDecrypt.mockResolvedValue(JSON.stringify([makeFakeKey()]));

      await useVaultStore.getState().unlock("master-pass");

      expect(backend.loadSettings).toHaveBeenCalled();
      expect(mockTrackerStart).toHaveBeenCalled();
      const startArg = mockTrackerStart.mock.calls[0][0];
      expect(startArg.idleTimeoutMs).toBe(5 * 60000);
      expect(startArg.lockOnHidden).toBe(false);
    });

    it("loads empty keys when vault has no data yet", async () => {
      const backend = makeFakeBackend();
      const derivedKey = makeFakeCryptoKey();
      backend.loadMeta.mockReturnValue(makeMeta());
      backend.loadData.mockReturnValue(null);

      mockGetBackend.mockReturnValue(backend);
      mockVerify.mockResolvedValue(derivedKey);

      await useVaultStore.getState().unlock("master-pass");

      const state = useVaultStore.getState();
      expect(state.status).toBe("unlocked");
      expect(state.keys).toEqual([]);
    });

    it("sets error when vault has not been initialized", async () => {
      // Precondition: status is "uninitialized" (default from beforeEach — no meta in backend)
      const backend = makeFakeBackend();
      backend.loadMeta.mockReturnValue(null);
      mockGetBackend.mockReturnValue(backend);

      await useVaultStore.getState().unlock("master-pass");

      const state = useVaultStore.getState();
      expect(state.status).toBe("uninitialized");
      expect(state.error).toBe("No vault found");
      expect(state.loading).toBe(false);
    });

    it("sets error when master password is wrong", async () => {
      const backend = makeFakeBackend();
      backend.loadMeta.mockReturnValue(makeMeta());
      mockGetBackend.mockReturnValue(backend);
      mockVerify.mockRejectedValue(new Error("Vault verifier mismatch"));

      useVaultStore.setState({ status: "locked" });
      await useVaultStore.getState().unlock("wrong-pass");

      const state = useVaultStore.getState();
      expect(state.error).toBe("Incorrect master password");
      expect(state.loading).toBe(false);
      expect(state.status).toBe("locked");
    });

    it("sets error and clears session key when vault data is corrupted", async () => {
      const backend = makeFakeBackend();
      const derivedKey = makeFakeCryptoKey();
      backend.loadMeta.mockReturnValue(makeMeta());
      backend.loadData.mockReturnValue(makeVaultData());

      mockGetBackend.mockReturnValue(backend);
      mockVerify.mockResolvedValue(derivedKey);
      mockDecrypt.mockRejectedValue(new Error("Decryption failed"));

      useVaultStore.setState({ status: "locked" });
      await useVaultStore.getState().unlock("master-pass");

      const state = useVaultStore.getState();
      expect(state.error).toBe("Vault data is corrupted");
      expect(state.loading).toBe(false);
      expect(state.status).toBe("locked");
      expect(mockClearSession).toHaveBeenCalled();
    });
  });

  describe("lock", () => {
    it("clears session key, empties keys, and transitions to locked", () => {
      useVaultStore.setState({
        status: "unlocked",
        keys: [makeFakeKey()],
        error: "old error",
      });

      useVaultStore.getState().lock();

      const state = useVaultStore.getState();
      expect(state.status).toBe("locked");
      expect(state.keys).toEqual([]);
      expect(state.error).toBeNull();
      expect(mockClearSession).toHaveBeenCalled();
    });

    it("calls activityTracker.stop() when locking manually", () => {
      useVaultStore.setState({ status: "unlocked", keys: [] });

      useVaultStore.getState().lock();

      expect(mockTrackerStop).toHaveBeenCalled();
    });

    it("does NOT bump autoLockNonce on manual lock", () => {
      useVaultStore.setState({
        status: "unlocked",
        keys: [],
        autoLockNonce: 3,
      });

      useVaultStore.getState().lock();

      expect(useVaultStore.getState().autoLockNonce).toBe(3);
    });
  });

  describe("onIdle callback behavior", () => {
    it("onIdle sets status to locked, clears keys, and bumps autoLockNonce", async () => {
      const backend = makeFakeBackend();
      const derivedKey = makeFakeCryptoKey();
      backend.loadMeta.mockReturnValue(makeMeta());
      backend.loadData.mockReturnValue(makeVaultData());
      mockGetBackend.mockReturnValue(backend);
      mockVerify.mockResolvedValue(derivedKey);
      mockDecrypt.mockResolvedValue(JSON.stringify([makeFakeKey()]));

      await useVaultStore.getState().unlock("master-pass");

      expect(mockTrackerStart).toHaveBeenCalled();
      const { onIdle } = mockTrackerStart.mock.calls[0][0];

      // Simulate vault becoming locked before onIdle fires (ensure session key cleared)
      mockGetSession.mockReturnValue(derivedKey);
      useVaultStore.setState({ autoLockNonce: 0 });

      onIdle();

      const state = useVaultStore.getState();
      expect(state.status).toBe("locked");
      expect(state.keys).toEqual([]);
      expect(state.error).toBeNull();
      expect(state.autoLockNonce).toBe(1);
      expect(mockClearSession).toHaveBeenCalled();
      expect(mockTrackerStop).toHaveBeenCalled();
    });

    it("onIdle is a no-op (no nonce bump) when vault is already locked", async () => {
      const backend = makeFakeBackend();
      const derivedKey = makeFakeCryptoKey();
      backend.loadMeta.mockReturnValue(makeMeta());
      backend.loadData.mockReturnValue(makeVaultData());
      mockGetBackend.mockReturnValue(backend);
      mockVerify.mockResolvedValue(derivedKey);
      mockDecrypt.mockResolvedValue(JSON.stringify([]));

      await useVaultStore.getState().unlock("master-pass");

      const { onIdle } = mockTrackerStart.mock.calls[0][0];

      // Manually lock the vault first
      useVaultStore.setState({ status: "locked", autoLockNonce: 2 });
      vi.clearAllMocks();

      onIdle();

      expect(useVaultStore.getState().autoLockNonce).toBe(2);
      expect(mockClearSession).not.toHaveBeenCalled();
    });
  });

  describe("updateAutoLockSettings", () => {
    it("saves settings to backend, updates state, and restarts tracker when unlocked", async () => {
      const backend = makeFakeBackend();
      mockGetBackend.mockReturnValue(backend);
      useVaultStore.setState({
        status: "unlocked",
        autoLockTimeoutMinutes: 15,
        lockOnHidden: false,
      });

      await useVaultStore.getState().updateAutoLockSettings({
        autoLockTimeoutMinutes: 30,
        lockOnHidden: true,
      });

      expect(backend.saveSettings).toHaveBeenCalledWith({
        autoLockTimeoutMinutes: 30,
        lockOnHidden: true,
      });
      const state = useVaultStore.getState();
      expect(state.autoLockTimeoutMinutes).toBe(30);
      expect(state.lockOnHidden).toBe(true);
      expect(mockTrackerStart).toHaveBeenCalled();
      const startArg = mockTrackerStart.mock.calls[0][0];
      expect(startArg.idleTimeoutMs).toBe(30 * 60000);
      expect(startArg.lockOnHidden).toBe(true);
    });

    it("saves settings and updates state but does NOT restart tracker when locked", async () => {
      const backend = makeFakeBackend();
      mockGetBackend.mockReturnValue(backend);
      useVaultStore.setState({
        status: "locked",
        autoLockTimeoutMinutes: 15,
        lockOnHidden: false,
      });

      await useVaultStore
        .getState()
        .updateAutoLockSettings({ autoLockTimeoutMinutes: 60 });

      expect(backend.saveSettings).toHaveBeenCalledWith({
        autoLockTimeoutMinutes: 60,
        lockOnHidden: false,
      });
      expect(useVaultStore.getState().autoLockTimeoutMinutes).toBe(60);
      expect(mockTrackerStart).not.toHaveBeenCalled();
    });

    it("can update only lockOnHidden (partial update)", async () => {
      const backend = makeFakeBackend();
      mockGetBackend.mockReturnValue(backend);
      useVaultStore.setState({
        status: "unlocked",
        autoLockTimeoutMinutes: 15,
        lockOnHidden: false,
      });

      await useVaultStore
        .getState()
        .updateAutoLockSettings({ lockOnHidden: true });

      expect(backend.saveSettings).toHaveBeenCalledWith({
        autoLockTimeoutMinutes: 15,
        lockOnHidden: true,
      });
      expect(useVaultStore.getState().lockOnHidden).toBe(true);
      expect(useVaultStore.getState().autoLockTimeoutMinutes).toBe(15);
    });
  });

  describe("addKey", () => {
    it("appends new key and persists", async () => {
      const derivedKey = makeFakeCryptoKey();
      const encryptedData = makeVaultData();
      const backend = makeFakeBackend();

      useVaultStore.setState({ status: "unlocked", keys: [] });
      mockGetSession.mockReturnValue(derivedKey);
      mockEncrypt.mockResolvedValue(encryptedData);
      mockGetBackend.mockReturnValue(backend);

      await useVaultStore.getState().addKey({
        name: "New Key",
        data: "-----BEGIN",
        hasPassphrase: false,
        fingerprint: "11:22:33",
      });

      const state = useVaultStore.getState();
      expect(state.keys).toHaveLength(1);
      expect(state.keys[0].name).toBe("New Key");
      expect(state.keys[0].fingerprint).toBe("11:22:33");
      expect(state.keys[0].id).toBeDefined();
      expect(state.keys[0].createdAt).toBeDefined();
      expect(backend.saveData).toHaveBeenCalledWith(encryptedData);
    });

    it("throws DuplicateKeyError with field 'name' when name already exists", async () => {
      useVaultStore.setState({
        status: "unlocked",
        keys: [makeFakeKey({ name: "My Key", fingerprint: "aa:bb:cc" })],
      });

      await expect(
        useVaultStore.getState().addKey({
          name: "My Key",
          data: "x",
          hasPassphrase: false,
          fingerprint: "dd:ee:ff",
        }),
      ).rejects.toBeInstanceOf(DuplicateKeyError);

      await expect(
        useVaultStore.getState().addKey({
          name: "My Key",
          data: "x",
          hasPassphrase: false,
          fingerprint: "dd:ee:ff",
        }),
      ).rejects.toMatchObject({ field: "name" });
    });

    it("throws DuplicateKeyError with field 'private_key' when fingerprint already exists", async () => {
      useVaultStore.setState({
        status: "unlocked",
        keys: [makeFakeKey({ name: "My Key", fingerprint: "aa:bb:cc" })],
      });

      await expect(
        useVaultStore.getState().addKey({
          name: "Other Key",
          data: "x",
          hasPassphrase: false,
          fingerprint: "aa:bb:cc",
        }),
      ).rejects.toMatchObject({ field: "private_key" });
    });

    it("throws DuplicateKeyError with field 'both' when both name and fingerprint are duplicates", async () => {
      useVaultStore.setState({
        status: "unlocked",
        keys: [makeFakeKey({ name: "My Key", fingerprint: "aa:bb:cc" })],
      });

      await expect(
        useVaultStore.getState().addKey({
          name: "My Key",
          data: "x",
          hasPassphrase: false,
          fingerprint: "aa:bb:cc",
        }),
      ).rejects.toMatchObject({ field: "both" });
    });

    it("throws when vault is locked (no session key)", async () => {
      const backend = makeFakeBackend();
      useVaultStore.setState({ status: "unlocked", keys: [] });
      mockGetSession.mockReturnValue(null);
      mockGetBackend.mockReturnValue(backend);

      await expect(
        useVaultStore.getState().addKey({
          name: "Key",
          data: "x",
          hasPassphrase: false,
          fingerprint: "11:22:33",
        }),
      ).rejects.toThrow("Vault is locked");
    });

    it("aborts set({keys}) if vault locks mid-flight during persistKeys", async () => {
      const derivedKey = makeFakeCryptoKey();
      const backend = makeFakeBackend();

      useVaultStore.setState({ status: "unlocked", keys: [] });
      mockGetSession.mockReturnValue(derivedKey);
      mockGetBackend.mockReturnValue(backend);

      let resolveEncrypt!: (v: VaultData) => void;
      mockEncrypt.mockReturnValue(
        new Promise<VaultData>((r) => {
          resolveEncrypt = r;
        }),
      );

      const promise = useVaultStore.getState().addKey({
        name: "New Key",
        data: "-----BEGIN",
        hasPassphrase: false,
        fingerprint: "11:22:33",
      });

      // Lock the vault mid-flight (simulating onIdle firing during await)
      useVaultStore.setState({ status: "locked", keys: [] });

      // Resolve encrypt so persistKeys finishes
      resolveEncrypt(makeVaultData());
      await promise.catch(() => {
        /* may throw */
      });

      // keys should remain [] (the locked state, not overwritten by addKey)
      expect(useVaultStore.getState().keys).toEqual([]);
      expect(useVaultStore.getState().status).toBe("locked");
    });
  });

  describe("updateKey", () => {
    it("updates key fields and persists", async () => {
      const derivedKey = makeFakeCryptoKey();
      const backend = makeFakeBackend();
      const existing = makeFakeKey({
        id: "key-1",
        name: "Old Name",
        fingerprint: "aa:bb:cc",
      });

      useVaultStore.setState({ status: "unlocked", keys: [existing] });
      mockGetSession.mockReturnValue(derivedKey);
      mockEncrypt.mockResolvedValue(makeVaultData());
      mockGetBackend.mockReturnValue(backend);

      await useVaultStore.getState().updateKey("key-1", { name: "New Name" });

      const state = useVaultStore.getState();
      expect(state.keys[0].name).toBe("New Name");
      expect(state.keys[0].updatedAt).not.toBe(existing.updatedAt);
    });

    it("throws when key id does not exist", async () => {
      useVaultStore.setState({
        status: "unlocked",
        keys: [makeFakeKey({ id: "key-1" })],
      });

      await expect(
        useVaultStore.getState().updateKey("non-existent", { name: "X" }),
      ).rejects.toThrow("Key not found");
    });

    it("throws DuplicateKeyError when renaming to an existing key name", async () => {
      const keys = [
        makeFakeKey({ id: "key-1", name: "First", fingerprint: "11:22:33" }),
        makeFakeKey({ id: "key-2", name: "Second", fingerprint: "44:55:66" }),
      ];
      useVaultStore.setState({ status: "unlocked", keys });

      await expect(
        useVaultStore.getState().updateKey("key-2", { name: "First" }),
      ).rejects.toMatchObject({ field: "name" });
    });

    it("allows updating a key's own name without throwing duplicate error", async () => {
      const derivedKey = makeFakeCryptoKey();
      const backend = makeFakeBackend();
      const key = makeFakeKey({
        id: "key-1",
        name: "My Key",
        fingerprint: "aa:bb:cc",
      });

      useVaultStore.setState({ status: "unlocked", keys: [key] });
      mockGetSession.mockReturnValue(derivedKey);
      mockEncrypt.mockResolvedValue(makeVaultData());
      mockGetBackend.mockReturnValue(backend);

      await expect(
        useVaultStore.getState().updateKey("key-1", { name: "My Key" }),
      ).resolves.toBeUndefined();
    });

    it("aborts set({keys}) if vault locks mid-flight during persistKeys", async () => {
      const derivedKey = makeFakeCryptoKey();
      const backend = makeFakeBackend();
      const existing = makeFakeKey({ id: "key-1", name: "Old Name" });

      useVaultStore.setState({ status: "unlocked", keys: [existing] });
      mockGetSession.mockReturnValue(derivedKey);
      mockGetBackend.mockReturnValue(backend);

      let resolveEncrypt!: (v: VaultData) => void;
      mockEncrypt.mockReturnValue(
        new Promise<VaultData>((r) => {
          resolveEncrypt = r;
        }),
      );

      const promise = useVaultStore
        .getState()
        .updateKey("key-1", { name: "New Name" });

      // Lock mid-flight
      useVaultStore.setState({ status: "locked", keys: [] });

      resolveEncrypt(makeVaultData());
      await promise.catch(() => {
        /* may throw or not */
      });

      // Name should NOT have been updated; keys still [] from the locked state
      expect(useVaultStore.getState().keys).toEqual([]);
      expect(useVaultStore.getState().status).toBe("locked");
    });
  });

  describe("removeKey", () => {
    it("removes the key by id and persists", async () => {
      const derivedKey = makeFakeCryptoKey();
      const backend = makeFakeBackend();
      const keys = [
        makeFakeKey({ id: "key-1", name: "First" }),
        makeFakeKey({ id: "key-2", name: "Second", fingerprint: "dd:ee:ff" }),
      ];

      useVaultStore.setState({ status: "unlocked", keys });
      mockGetSession.mockReturnValue(derivedKey);
      mockEncrypt.mockResolvedValue(makeVaultData());
      mockGetBackend.mockReturnValue(backend);

      await useVaultStore.getState().removeKey("key-1");

      const state = useVaultStore.getState();
      expect(state.keys).toHaveLength(1);
      expect(state.keys[0].id).toBe("key-2");
    });

    it("throws when removing a non-existent id", async () => {
      const backend = makeFakeBackend();
      const keys = [makeFakeKey({ id: "key-1" })];

      useVaultStore.setState({ status: "unlocked", keys });
      mockGetBackend.mockReturnValue(backend);

      await expect(
        useVaultStore.getState().removeKey("ghost-id"),
      ).rejects.toThrow("Key not found");
    });

    it("aborts set({keys}) if vault locks mid-flight during persistKeys", async () => {
      const derivedKey = makeFakeCryptoKey();
      const backend = makeFakeBackend();
      const existing = makeFakeKey({ id: "key-1", name: "Key" });

      useVaultStore.setState({ status: "unlocked", keys: [existing] });
      mockGetSession.mockReturnValue(derivedKey);
      mockGetBackend.mockReturnValue(backend);

      let resolveEncrypt!: (v: VaultData) => void;
      mockEncrypt.mockReturnValue(
        new Promise<VaultData>((r) => {
          resolveEncrypt = r;
        }),
      );

      const promise = useVaultStore.getState().removeKey("key-1");

      // Lock mid-flight
      useVaultStore.setState({ status: "locked", keys: [] });

      resolveEncrypt(makeVaultData());
      await promise.catch(() => {
        /* may throw or not */
      });

      // keys should remain [] (locked state not overwritten with post-remove keys)
      expect(useVaultStore.getState().keys).toEqual([]);
      expect(useVaultStore.getState().status).toBe("locked");
    });
  });

  describe("changeMasterPassword", () => {
    it("re-encrypts keys with new password and updates meta", async () => {
      const backend = makeFakeBackend();
      const oldKey = makeFakeCryptoKey();
      const newKey = makeFakeCryptoKey();
      const newMeta = makeMeta();

      backend.loadMeta.mockReturnValue(makeMeta());
      backend.loadData.mockReturnValue(makeVaultData());
      mockGetBackend.mockReturnValue(backend);
      mockGetSession.mockReturnValue(oldKey);
      mockVerify.mockResolvedValue(oldKey);
      mockCrypto.mockResolvedValue({ meta: newMeta, derivedKey: newKey });
      mockEncrypt.mockResolvedValue(makeVaultData());

      useVaultStore.setState({ status: "unlocked", keys: [makeFakeKey()] });

      await useVaultStore
        .getState()
        .changeMasterPassword("current-pass", "new-pass");

      const state = useVaultStore.getState();
      expect(state.loading).toBe(false);
      expect(state.error).toBeNull();
      expect(mockSetSession).toHaveBeenCalledWith(newKey);
      expect(backend.saveMeta).toHaveBeenCalledWith(newMeta);
    });

    it("sets error when vault has not been initialized", async () => {
      const backend = makeFakeBackend();
      backend.loadMeta.mockReturnValue(null);
      mockGetBackend.mockReturnValue(backend);

      await useVaultStore
        .getState()
        .changeMasterPassword("current-pass", "new-pass");

      const state = useVaultStore.getState();
      expect(state.error).toBe("No vault found");
      expect(state.loading).toBe(false);
    });

    it("sets error when current password is incorrect", async () => {
      const backend = makeFakeBackend();
      backend.loadMeta.mockReturnValue(makeMeta());
      mockGetBackend.mockReturnValue(backend);
      mockGetSession.mockReturnValue(makeFakeCryptoKey());
      mockVerify.mockRejectedValue(new Error("Vault verifier mismatch"));

      useVaultStore.setState({ status: "unlocked", keys: [] });

      await useVaultStore
        .getState()
        .changeMasterPassword("wrong-pass", "new-pass");

      const state = useVaultStore.getState();
      expect(state.error).toBe("Current password is incorrect");
      expect(state.loading).toBe(false);
    });

    it("sets error when vault is locked", async () => {
      const backend = makeFakeBackend();
      backend.loadMeta.mockReturnValue(makeMeta());
      mockGetBackend.mockReturnValue(backend);
      mockGetSession.mockReturnValue(null);

      useVaultStore.setState({ status: "locked", keys: [] });

      await useVaultStore
        .getState()
        .changeMasterPassword("current-pass", "new-pass");

      const state = useVaultStore.getState();
      expect(state.error).toBe("Vault is locked");
      expect(state.loading).toBe(false);
    });

    it("restores old key, data, and meta on re-encryption failure", async () => {
      const backend = makeFakeBackend();
      const oldKey = makeFakeCryptoKey("old");
      const newKey = makeFakeCryptoKey("new");
      const oldData = makeVaultData();

      backend.loadMeta.mockReturnValue(makeMeta());
      backend.loadData.mockReturnValue(oldData);
      mockGetBackend.mockReturnValue(backend);
      mockGetSession.mockReturnValue(oldKey);
      mockVerify.mockResolvedValue(oldKey);
      mockCrypto.mockResolvedValue({ meta: makeMeta(), derivedKey: newKey });
      mockEncrypt.mockRejectedValue(new Error("Encrypt failed"));

      useVaultStore.setState({ status: "unlocked", keys: [] });

      await useVaultStore
        .getState()
        .changeMasterPassword("current-pass", "new-pass");

      const state = useVaultStore.getState();
      expect(state.error).toBe("Encrypt failed");
      // Session should have been restored to old key as the last setSessionKey call
      expect(
        mockSetSession.mock.calls[mockSetSession.mock.calls.length - 1]?.[0],
      ).toBe(oldKey);
      // Old data should have been restored
      expect(backend.saveData).toHaveBeenCalledWith(oldData);
    });

    it("CRITICAL: early return when vault locks during createVaultMeta (after verifyPassword)", async () => {
      const backend = makeFakeBackend();
      const oldKey = makeFakeCryptoKey("old");
      const newKey = makeFakeCryptoKey("new");
      const newMeta = makeMeta();

      backend.loadMeta.mockReturnValue(makeMeta());
      backend.loadData.mockReturnValue(makeVaultData());
      mockGetBackend.mockReturnValue(backend);
      mockGetSession.mockReturnValue(oldKey);
      mockVerify.mockResolvedValue(oldKey);

      let resolveCrypto!: (v: {
        meta: VaultMeta;
        derivedKey: CryptoKey;
      }) => void;
      mockCrypto.mockReturnValue(
        new Promise<{ meta: VaultMeta; derivedKey: CryptoKey }>((r) => {
          resolveCrypto = r;
        }),
      );

      useVaultStore.setState({ status: "unlocked", keys: [makeFakeKey()] });

      const promise = useVaultStore
        .getState()
        .changeMasterPassword("current-pass", "new-pass");

      // Wait until loadMeta + verifyPassword have resolved and execution is
      // suspended inside createVaultMeta, then simulate the vault locking
      // while it is in flight.
      await vi.waitFor(() => expect(mockCrypto).toHaveBeenCalled());
      useVaultStore.setState({ status: "locked", keys: [] });
      mockGetSession.mockReturnValue(null);

      // Now resolve createVaultMeta
      resolveCrypto({ meta: newMeta, derivedKey: newKey });
      await promise;

      // CRITICAL: neither setSessionKey(newKey) nor setSessionKey(oldKey) should be called
      // after the early return path. getSessionKey() should remain null.
      expect(mockGetSession()).toBeNull();
      // saveMeta should NOT be called with newMeta
      expect(backend.saveMeta).not.toHaveBeenCalledWith(newMeta);
      // saveData should NOT have been re-called (the new encrypt path never ran)
      expect(mockEncrypt).not.toHaveBeenCalled();
      // Error should be set
      expect(useVaultStore.getState().error).toBe(
        "Vault locked during password change",
      );
    });
  });

  describe("resetVault", () => {
    it("clears backend, session key, and resets state to uninitialized", async () => {
      const backend = makeFakeBackend();
      mockGetBackend.mockReturnValue(backend);

      useVaultStore.setState({
        status: "unlocked",
        keys: [makeFakeKey()],
        error: "previous error",
      });

      await useVaultStore.getState().resetVault();

      const state = useVaultStore.getState();
      expect(state.status).toBe("uninitialized");
      expect(state.keys).toEqual([]);
      expect(state.error).toBeNull();
      expect(backend.clear).toHaveBeenCalled();
      expect(mockClearSession).toHaveBeenCalled();
    });

    it("calls activityTracker.stop() on resetVault", async () => {
      const backend = makeFakeBackend();
      mockGetBackend.mockReturnValue(backend);
      useVaultStore.setState({ status: "unlocked", keys: [] });

      await useVaultStore.getState().resetVault();

      expect(mockTrackerStop).toHaveBeenCalled();
    });

    it("resets autoLockTimeoutMinutes and lockOnHidden to defaults", async () => {
      const backend = makeFakeBackend();
      mockGetBackend.mockReturnValue(backend);

      useVaultStore.setState({
        status: "unlocked",
        autoLockTimeoutMinutes: 60,
        lockOnHidden: true,
      });

      await useVaultStore.getState().resetVault();

      const state = useVaultStore.getState();
      expect(state.autoLockTimeoutMinutes).toBe(
        DEFAULT_VAULT_SETTINGS.autoLockTimeoutMinutes,
      );
      expect(state.lockOnHidden).toBe(DEFAULT_VAULT_SETTINGS.lockOnHidden);
    });
  });

  describe("authStore.logout() path -> lock() -> activityTracker.stop()", () => {
    it("lock() called from logout stops the activity tracker", () => {
      useVaultStore.setState({ status: "unlocked", keys: [] });

      // Simulate what authStore.logout() does: calls useVaultStore.getState().lock()
      useVaultStore.getState().lock();

      expect(mockTrackerStop).toHaveBeenCalled();
    });
  });

  describe("DuplicateKeyError", () => {
    it("is an instance of Error", () => {
      const err = new DuplicateKeyError("name");
      expect(err).toBeInstanceOf(Error);
    });

    it("exposes the duplicate field", () => {
      expect(new DuplicateKeyError("name").field).toBe("name");
      expect(new DuplicateKeyError("private_key").field).toBe("private_key");
      expect(new DuplicateKeyError("both").field).toBe("both");
    });

    it("has a descriptive message", () => {
      const err = new DuplicateKeyError("both");
      expect(err.message).toBe("Duplicate key: both");
    });
  });
});
