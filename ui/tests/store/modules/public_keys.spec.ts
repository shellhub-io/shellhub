import { describe, expect, it, beforeEach, afterEach } from "vitest";
import MockAdapter from "axios-mock-adapter";
import { createPinia, setActivePinia } from "pinia";
import { sshApi } from "@/api/http";
import { IPublicKey } from "@/interfaces/IPublicKey";
import usePublicKeysStore from "@/store/modules/public_keys";
import { buildUrl } from "../../utils/url";

const mockPublicKeyBase: IPublicKey = {
  data: "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQC...",
  fingerprint: "SHA256:abcd1234efgh5678",
  created_at: "2020-05-01T00:00:00.000Z",
  tenant_id: "fake-tenant-id",
  name: "my-public-key",
  filter: {
    hostname: ".*",
  },
  username: ".*",
};

describe("PublicKeys Store", () => {
  let publicKeysStore: ReturnType<typeof usePublicKeysStore>;
  let mockSshApi: MockAdapter;

  beforeEach(() => {
    setActivePinia(createPinia());
    publicKeysStore = usePublicKeysStore();
    mockSshApi = new MockAdapter(sshApi.getAxios());
  });

  afterEach(() => { mockSshApi.reset(); });

  describe("Initial State", () => {
    it("should have empty public keys array", () => {
      expect(publicKeysStore.publicKeys).toEqual([]);
    });

    it("should have zero public key count", () => {
      expect(publicKeysStore.publicKeyCount).toBe(0);
    });
  });

  describe("fetchPublicKeyList", () => {
    const baseUrl = "http://localhost:3000/api/sshkeys/public-keys";

    it("should fetch public keys successfully with default pagination", async () => {
      const publicKeyList = [mockPublicKeyBase];

      mockSshApi
        .onGet(buildUrl(baseUrl, { page: "1", per_page: "10" }))
        .reply(200, publicKeyList, { "x-total-count": "1" });

      await expect(publicKeysStore.fetchPublicKeyList()).resolves.not.toThrow();

      expect(publicKeysStore.publicKeys).toEqual(publicKeyList);
      expect(publicKeysStore.publicKeyCount).toBe(1);
    });

    it("should fetch public keys successfully with custom pagination", async () => {
      const publicKeyList = [mockPublicKeyBase];

      mockSshApi
        .onGet(buildUrl(baseUrl, { page: "2", per_page: "20" }))
        .reply(200, publicKeyList, { "x-total-count": "5" });

      await expect(publicKeysStore.fetchPublicKeyList({ page: 2, perPage: 20 })).resolves.not.toThrow();

      expect(publicKeysStore.publicKeys).toEqual(publicKeyList);
      expect(publicKeysStore.publicKeyCount).toBe(5);
    });

    it("should fetch public keys successfully with filter", async () => {
      const publicKeyList = [mockPublicKeyBase];
      const filter = "my-key";

      mockSshApi
        .onGet(buildUrl(baseUrl, { filter: filter, page: "1", per_page: "10" }))
        .reply(200, publicKeyList, { "x-total-count": "1" });

      await expect(publicKeysStore.fetchPublicKeyList({ page: 1, perPage: 10, filter })).resolves.not.toThrow();

      expect(publicKeysStore.publicKeys).toEqual(publicKeyList);
    });

    it("should fetch empty list successfully", async () => {
      mockSshApi
        .onGet(buildUrl(baseUrl, { page: "1", per_page: "10" }))
        .reply(200, [], { "x-total-count": "0" });

      await expect(publicKeysStore.fetchPublicKeyList()).resolves.not.toThrow();

      expect(publicKeysStore.publicKeys).toEqual([]);
      expect(publicKeysStore.publicKeyCount).toBe(0);
    });

    it("should handle not found error when fetching public keys", async () => {
      mockSshApi
        .onGet(buildUrl(baseUrl, { page: "1", per_page: "10" }))
        .reply(404, { message: "Public keys not found" });

      await expect(publicKeysStore.fetchPublicKeyList()).rejects.toBeAxiosErrorWithStatus(404);
    });

    it("should handle server error when fetching public keys", async () => {
      mockSshApi
        .onGet(buildUrl(baseUrl, { page: "1", per_page: "10" }))
        .reply(500);

      await expect(publicKeysStore.fetchPublicKeyList()).rejects.toBeAxiosErrorWithStatus(500);
    });

    it("should handle network error when fetching public keys", async () => {
      mockSshApi
        .onGet(buildUrl(baseUrl, { page: "1", per_page: "10" }))
        .networkError();

      await expect(publicKeysStore.fetchPublicKeyList()).rejects.toThrow("Network Error");
    });
  });

  describe("createPublicKey", () => {
    const createUrl = "http://localhost:3000/api/sshkeys/public-keys";

    it("should create public key successfully", async () => {
      const newKey = {
        data: mockPublicKeyBase.data,
        name: mockPublicKeyBase.name,
        filter: mockPublicKeyBase.filter,
        username: mockPublicKeyBase.username,
      };

      mockSshApi
        .onPost(createUrl)
        .reply(200);

      await expect(publicKeysStore.createPublicKey(newKey)).resolves.not.toThrow();
    });

    it("should create public key with custom filter successfully", async () => {
      const newKey = {
        data: "ssh-rsa AAAAB3NzaC1...",
        name: "custom-key",
        filter: {
          hostname: "server-.*",
        },
        username: "admin",
      };

      mockSshApi
        .onPost(createUrl)
        .reply(200);

      await expect(publicKeysStore.createPublicKey(newKey)).resolves.not.toThrow();
    });

    it("should handle validation error when creating public key", async () => {
      const newKey = {
        data: "invalid-key-data",
        name: "invalid",
        filter: { hostname: ".*" },
        username: ".*",
      };

      mockSshApi
        .onPost(createUrl)
        .reply(400, { message: "Invalid public key format" });

      await expect(publicKeysStore.createPublicKey(newKey)).rejects.toBeAxiosErrorWithStatus(400);
    });

    it("should handle server error when creating public key", async () => {
      const newKey = {
        data: mockPublicKeyBase.data,
        name: mockPublicKeyBase.name,
        filter: mockPublicKeyBase.filter,
        username: mockPublicKeyBase.username,
      };

      mockSshApi
        .onPost(createUrl)
        .reply(500);

      await expect(publicKeysStore.createPublicKey(newKey)).rejects.toBeAxiosErrorWithStatus(500);
    });

    it("should handle network error when creating public key", async () => {
      const newKey = {
        data: mockPublicKeyBase.data,
        name: mockPublicKeyBase.name,
        filter: mockPublicKeyBase.filter,
        username: mockPublicKeyBase.username,
      };

      mockSshApi
        .onPost(createUrl)
        .networkError();

      await expect(publicKeysStore.createPublicKey(newKey)).rejects.toThrow("Network Error");
    });
  });

  describe("updatePublicKey", () => {
    const encodedFingerprint = encodeURIComponent(mockPublicKeyBase.fingerprint);
    const updateUrl = `http://localhost:3000/api/sshkeys/public-keys/${encodedFingerprint}`;

    it("should update public key successfully", async () => {
      const updatedKey = {
        ...mockPublicKeyBase,
        name: "updated-key-name",
      };

      mockSshApi
        .onPut(updateUrl)
        .reply(200);

      await expect(publicKeysStore.updatePublicKey(updatedKey)).resolves.not.toThrow();
    });

    it("should update public key filter successfully", async () => {
      const updatedKey = {
        ...mockPublicKeyBase,
        filter: {
          hostname: "production-.*",
        },
      };

      mockSshApi
        .onPut(updateUrl)
        .reply(200);

      await expect(publicKeysStore.updatePublicKey(updatedKey)).resolves.not.toThrow();
    });

    it("should update public key username successfully", async () => {
      const updatedKey = {
        ...mockPublicKeyBase,
        username: "admin",
      };

      mockSshApi
        .onPut(updateUrl)
        .reply(200);

      await expect(publicKeysStore.updatePublicKey(updatedKey)).resolves.not.toThrow();
    });

    it("should handle not found error when updating public key", async () => {
      const updatedKey = {
        ...mockPublicKeyBase,
        fingerprint: "non-existent-fingerprint",
      };

      mockSshApi
        .onPut(updateUrl)
        .reply(404, { message: "Public key not found" });

      await expect(publicKeysStore.updatePublicKey(updatedKey)).rejects.toBeAxiosErrorWithStatus(404);
    });

    it("should handle server error when updating public key", async () => {
      mockSshApi
        .onPut(updateUrl)
        .reply(500);

      await expect(publicKeysStore.updatePublicKey(mockPublicKeyBase)).rejects.toBeAxiosErrorWithStatus(500);
    });

    it("should handle network error when updating public key", async () => {
      mockSshApi
        .onPut(updateUrl)
        .networkError();

      await expect(publicKeysStore.updatePublicKey(mockPublicKeyBase)).rejects.toThrow("Network Error");
    });
  });

  describe("deletePublicKey", () => {
    beforeEach(() => {
      publicKeysStore.publicKeys = [mockPublicKeyBase];
      publicKeysStore.publicKeyCount = 1;
    });

    it("should delete public key successfully and remove from list", async () => {
      const { fingerprint } = mockPublicKeyBase;

      mockSshApi
        .onDelete(`http://localhost:3000/api/sshkeys/public-keys/${encodeURIComponent(fingerprint)}`)
        .reply(200);

      await expect(publicKeysStore.deletePublicKey(fingerprint)).resolves.not.toThrow();

      expect(publicKeysStore.publicKeys).toEqual([]);
      expect(publicKeysStore.publicKeyCount).toBe(0);
    });

    it("should delete only specified key when multiple keys exist", async () => {
      const secondKey = {
        ...mockPublicKeyBase,
        fingerprint: "SHA256:different",
        name: "second-key",
      };
      publicKeysStore.publicKeys = [mockPublicKeyBase, secondKey];
      publicKeysStore.publicKeyCount = 2;

      mockSshApi
        .onDelete(`http://localhost:3000/api/sshkeys/public-keys/${encodeURIComponent(mockPublicKeyBase.fingerprint)}`)
        .reply(200);

      await expect(publicKeysStore.deletePublicKey(mockPublicKeyBase.fingerprint)).resolves.not.toThrow();

      expect(publicKeysStore.publicKeys).toEqual([secondKey]);
      expect(publicKeysStore.publicKeyCount).toBe(1);
    });

    it("should handle deleting non-existent key gracefully", async () => {
      const nonExistentFingerprint = "non-existent";

      mockSshApi
        .onDelete(`http://localhost:3000/api/sshkeys/public-keys/${encodeURIComponent(nonExistentFingerprint)}`)
        .reply(200);

      await expect(publicKeysStore.deletePublicKey(nonExistentFingerprint)).resolves.not.toThrow();

      expect(publicKeysStore.publicKeys).toEqual([mockPublicKeyBase]);
      expect(publicKeysStore.publicKeyCount).toBe(1);
    });

    it("should handle not found error when deleting public key", async () => {
      const { fingerprint } = mockPublicKeyBase;

      mockSshApi
        .onDelete(`http://localhost:3000/api/sshkeys/public-keys/${encodeURIComponent(fingerprint)}`)
        .reply(404, { message: "Public key not found" });

      await expect(publicKeysStore.deletePublicKey(fingerprint)).rejects.toBeAxiosErrorWithStatus(404);
    });

    it("should handle server error when deleting public key", async () => {
      const { fingerprint } = mockPublicKeyBase;

      mockSshApi
        .onDelete(`http://localhost:3000/api/sshkeys/public-keys/${encodeURIComponent(fingerprint)}`)
        .reply(500);

      await expect(publicKeysStore.deletePublicKey(fingerprint)).rejects.toBeAxiosErrorWithStatus(500);
    });

    it("should handle network error when deleting public key", async () => {
      const { fingerprint } = mockPublicKeyBase;

      mockSshApi
        .onDelete(`http://localhost:3000/api/sshkeys/public-keys/${encodeURIComponent(fingerprint)}`)
        .networkError();

      await expect(publicKeysStore.deletePublicKey(fingerprint)).rejects.toThrow("Network Error");
    });
  });
});
