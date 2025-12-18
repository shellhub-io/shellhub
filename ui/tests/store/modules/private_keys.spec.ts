import { describe, expect, it, beforeEach, afterEach } from "vitest";
import { createPinia, setActivePinia } from "pinia";
import { IPrivateKey } from "@/interfaces/IPrivateKey";
import usePrivateKeysStore from "@/store/modules/private_keys";

const mockPrivateKeyBase: Omit<IPrivateKey, "id"> = {
  data: "-----BEGIN RSA PRIVATE KEY-----\nMIIEpAIBAAKCAQEA...",
  name: "my-private-key",
  hasPassphrase: false,
  fingerprint: "SHA256:abcd1234efgh5678",
};

describe("PrivateKeys Store", () => {
  let privateKeysStore: ReturnType<typeof usePrivateKeysStore>;

  beforeEach(() => {
    localStorage.clear();
    setActivePinia(createPinia());
    privateKeysStore = usePrivateKeysStore();
  });

  afterEach(() => { localStorage.clear(); });

  describe("Initial State", () => {
    it("should have empty private keys array", () => {
      expect(privateKeysStore.privateKeys).toEqual([]);
    });

    it("should have empty array when localStorage is empty", () => {
      privateKeysStore.getPrivateKeyList();

      expect(privateKeysStore.privateKeys).toEqual([]);
    });
  });

  describe("getPrivateKeyList", () => {
    it("should load private keys from localStorage", () => {
      const privateKeys: IPrivateKey[] = [
        { ...mockPrivateKeyBase, id: 1 },
        { ...mockPrivateKeyBase, id: 2, name: "another-key", fingerprint: "SHA256:xyz9876" },
      ];

      localStorage.setItem("privateKeys", JSON.stringify(privateKeys));

      privateKeysStore.getPrivateKeyList();

      expect(privateKeysStore.privateKeys).toEqual(privateKeys);
    });

    it("should handle malformed localStorage data gracefully", () => {
      localStorage.setItem("privateKeys", "invalid-json");

      expect(() => privateKeysStore.getPrivateKeyList()).toThrow();
    });
  });

  describe("addPrivateKey", () => {
    it("should add private key successfully with id 1 when list is empty", () => {
      privateKeysStore.addPrivateKey(mockPrivateKeyBase);

      expect(privateKeysStore.privateKeys).toHaveLength(1);
      expect(privateKeysStore.privateKeys[0]).toEqual({ ...mockPrivateKeyBase, id: 1 });
    });

    it("should persist private key to localStorage", () => {
      privateKeysStore.addPrivateKey(mockPrivateKeyBase);

      const storedKeys = JSON.parse(localStorage.getItem("privateKeys") || "[]");
      expect(storedKeys).toEqual([{ ...mockPrivateKeyBase, id: 1 }]);
    });

    it("should add private key with auto-incremented id", () => {
      privateKeysStore.addPrivateKey(mockPrivateKeyBase);
      privateKeysStore.addPrivateKey({ ...mockPrivateKeyBase, name: "key-2", data: "data-2", fingerprint: "SHA256:xyz" });

      expect(privateKeysStore.privateKeys).toHaveLength(2);
      expect(privateKeysStore.privateKeys[1].id).toBe(2);
    });

    it("should generate correct id after max existing id", () => {
      localStorage.setItem("privateKeys", JSON.stringify([
        { ...mockPrivateKeyBase, id: 5 },
        { ...mockPrivateKeyBase, id: 3, name: "key-3", data: "data-3" },
      ]));
      privateKeysStore.getPrivateKeyList();

      privateKeysStore.addPrivateKey({ ...mockPrivateKeyBase, name: "new-key", data: "new-data", fingerprint: "SHA256:new" });

      expect(privateKeysStore.privateKeys).toHaveLength(3);
      expect(privateKeysStore.privateKeys[2].id).toBe(6);
    });

    it("should add private key with passphrase", () => {
      const keyWithPassphrase = { ...mockPrivateKeyBase, hasPassphrase: true };

      privateKeysStore.addPrivateKey(keyWithPassphrase);

      expect(privateKeysStore.privateKeys[0].hasPassphrase).toBe(true);
    });

    it("should throw error when adding key with duplicate name", () => {
      privateKeysStore.addPrivateKey(mockPrivateKeyBase);

      expect(() => {
        privateKeysStore.addPrivateKey({ ...mockPrivateKeyBase, data: "different-data" });
      }).toThrow("name");
    });

    it("should throw error when adding key with duplicate data", () => {
      privateKeysStore.addPrivateKey(mockPrivateKeyBase);

      expect(() => {
        privateKeysStore.addPrivateKey({ ...mockPrivateKeyBase, name: "different-name" });
      }).toThrow("private_key");
    });

    it("should throw error when adding key with both duplicate name and data", () => {
      privateKeysStore.addPrivateKey(mockPrivateKeyBase);

      expect(() => {
        privateKeysStore.addPrivateKey(mockPrivateKeyBase);
      }).toThrow("both");
    });
  });

  describe("editPrivateKey", () => {
    beforeEach(() => {
      privateKeysStore.addPrivateKey(mockPrivateKeyBase);
    });

    it("should edit private key successfully", () => {
      const updatedKey: IPrivateKey = {
        id: 1,
        data: "updated-data",
        name: "updated-name",
        hasPassphrase: true,
        fingerprint: "SHA256:updated",
      };

      privateKeysStore.editPrivateKey(updatedKey);

      expect(privateKeysStore.privateKeys[0]).toEqual(updatedKey);
    });

    it("should persist edited private key to localStorage", () => {
      const updatedKey: IPrivateKey = {
        id: 1,
        data: "updated-data",
        name: "updated-name",
        hasPassphrase: true,
        fingerprint: "SHA256:updated",
      };

      privateKeysStore.editPrivateKey(updatedKey);

      const storedKeys = JSON.parse(localStorage.getItem("privateKeys") || "[]");
      expect(storedKeys[0]).toEqual(updatedKey);
    });

    it("should edit only specified key when multiple keys exist", () => {
      privateKeysStore.addPrivateKey({ ...mockPrivateKeyBase, name: "key-2", data: "data-2", fingerprint: "SHA256:key2" });

      const updatedKey: IPrivateKey = {
        id: 1,
        ...mockPrivateKeyBase,
        name: "updated-first-key",
      };

      privateKeysStore.editPrivateKey(updatedKey);

      expect(privateKeysStore.privateKeys[0].name).toBe("updated-first-key");
      expect(privateKeysStore.privateKeys[1].name).toBe("key-2");
    });

    it("should allow editing key without changing name or data", () => {
      const updatedKey: IPrivateKey = {
        id: 1,
        ...mockPrivateKeyBase,
        hasPassphrase: true,
      };

      privateKeysStore.editPrivateKey(updatedKey);

      expect(privateKeysStore.privateKeys[0].hasPassphrase).toBe(true);
    });

    it("should throw error when editing non-existent key", () => {
      const nonExistentKey: IPrivateKey = {
        id: 999,
        ...mockPrivateKeyBase,
      };

      expect(() => {
        privateKeysStore.editPrivateKey(nonExistentKey);
      }).toThrow("Key not found");
    });

    it("should throw error when editing to duplicate name", () => {
      privateKeysStore.addPrivateKey({ ...mockPrivateKeyBase, name: "key-2", fingerprint: "SHA256:key2", data: "data-2" });

      const updatedKey: IPrivateKey = {
        id: 1,
        ...mockPrivateKeyBase,
        name: "key-2",
      };

      expect(() => {
        privateKeysStore.editPrivateKey(updatedKey);
      }).toThrow("name");
    });

    it("should throw error when editing to duplicate data", () => {
      privateKeysStore.addPrivateKey({ ...mockPrivateKeyBase, name: "key-2", fingerprint: "SHA256:key2", data: "data-2" });

      const updatedKey: IPrivateKey = {
        id: 1,
        data: "data-2",
        name: "unique-name",
        hasPassphrase: false,
        fingerprint: "SHA256:unique",
      };

      expect(() => {
        privateKeysStore.editPrivateKey(updatedKey);
      }).toThrow("private_key");
    });

    it("should throw error when editing to both duplicate name and data", () => {
      const secondKey = { ...mockPrivateKeyBase, name: "key-2", fingerprint: "SHA256:key2", data: "data-2" };
      privateKeysStore.addPrivateKey(secondKey);

      const updatedKey: IPrivateKey = {
        id: 1,
        ...secondKey,
      };

      expect(() => {
        privateKeysStore.editPrivateKey(updatedKey);
      }).toThrow("both");
    });

    it("should allow editing key to same name and data", () => {
      const updatedKey: IPrivateKey = {
        id: 1,
        ...mockPrivateKeyBase,
        hasPassphrase: true,
      };

      expect(() => {
        privateKeysStore.editPrivateKey(updatedKey);
      }).not.toThrow();

      expect(privateKeysStore.privateKeys[0].hasPassphrase).toBe(true);
    });
  });

  describe("deletePrivateKey", () => {
    it("should delete private key successfully", () => {
      privateKeysStore.addPrivateKey(mockPrivateKeyBase);

      privateKeysStore.deletePrivateKey(1);

      expect(privateKeysStore.privateKeys).toHaveLength(0);
    });

    it("should persist deletion to localStorage", () => {
      privateKeysStore.addPrivateKey(mockPrivateKeyBase);

      privateKeysStore.deletePrivateKey(1);

      const storedKeys = JSON.parse(localStorage.getItem("privateKeys") || "[]");
      expect(storedKeys).toHaveLength(0);
    });

    it("should delete only specified key when multiple keys exist", () => {
      privateKeysStore.addPrivateKey(mockPrivateKeyBase);
      privateKeysStore.addPrivateKey({ ...mockPrivateKeyBase, name: "key-2", data: "data-2", fingerprint: "SHA256:key2" });

      privateKeysStore.deletePrivateKey(1);

      expect(privateKeysStore.privateKeys).toHaveLength(1);
      expect(privateKeysStore.privateKeys[0].id).toBe(2);
    });

    it("should handle deleting non-existent key gracefully", () => {
      privateKeysStore.addPrivateKey(mockPrivateKeyBase);

      privateKeysStore.deletePrivateKey(999);

      expect(privateKeysStore.privateKeys).toHaveLength(1);
    });

    it("should handle deleting from empty list gracefully", () => {
      privateKeysStore.deletePrivateKey(1);

      expect(privateKeysStore.privateKeys).toHaveLength(0);
    });

    it("should update localStorage after deleting last key", () => {
      privateKeysStore.addPrivateKey(mockPrivateKeyBase);

      privateKeysStore.deletePrivateKey(1);

      const storedKeys = JSON.parse(localStorage.getItem("privateKeys") || "[]");
      expect(storedKeys).toEqual([]);
    });
  });
});
