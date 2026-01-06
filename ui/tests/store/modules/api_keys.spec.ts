import { describe, expect, it, beforeEach, afterEach } from "vitest";
import MockAdapter from "axios-mock-adapter";
import { createPinia, setActivePinia } from "pinia";
import { apiKeysApi } from "@/api/http";
import useApiKeysStore from "@/store/modules/api_keys";
import { buildUrl } from "../../utils/url";

describe("API Keys Store", () => {
  let mockApiKeysApi: MockAdapter;
  let store: ReturnType<typeof useApiKeysStore>;

  beforeEach(() => {
    setActivePinia(createPinia());
    mockApiKeysApi = new MockAdapter(apiKeysApi.getAxios());
    store = useApiKeysStore();
  });

  afterEach(() => { mockApiKeysApi.reset(); });

  describe("Initial State", () => {
    it("should have empty API keys array", () => {
      expect(store.apiKeys).toEqual([]);
    });

    it("should have zero API keys count", () => {
      expect(store.apiKeysCount).toBe(0);
    });
  });

  describe("fetchApiKeys", () => {
    const baseUrl = "http://localhost:3000/api/namespaces/api-key";

    it("should fetch API keys successfully with pagination", async () => {
      const mockApiKeys = [
        {
          id: "3e5a5194-test-4a32-fake-7434c6d49df1",
          tenant_id: "fake-tenant",
          name: "my api key",
          expires_in: 1707958989,
        },
        {
          id: "a1b2c3d4-5678-90ef-ghij-klmnopqrstuv",
          tenant_id: "fake-tenant",
          name: "production key",
          expires_in: 1709958989,
        },
      ];

      mockApiKeysApi
        .onGet(buildUrl(baseUrl, { page: "1", per_page: "10" }))
        .reply(200, mockApiKeys, {
          "x-total-count": "2",
        });

      await store.fetchApiKeys({ page: 1, perPage: 10 });

      expect(store.apiKeys).toEqual(mockApiKeys);
      expect(store.apiKeysCount).toBe(2);
    });

    it("should handle empty API keys response", async () => {
      mockApiKeysApi
        .onGet(buildUrl(baseUrl, { page: "1", per_page: "10" }))
        .reply(200, [], {
          "x-total-count": "0",
        });

      await store.fetchApiKeys({ page: 1, perPage: 10 });

      expect(store.apiKeys).toEqual([]);
      expect(store.apiKeysCount).toBe(0);
    });

    it("should fetch API keys with sorting parameters", async () => {
      const mockApiKeys = [
        {
          id: "key-1",
          tenant_id: "tenant",
          name: "api key 1",
          expires_in: 1707958989,
        },
      ];

      mockApiKeysApi
        .onGet(buildUrl(baseUrl, { page: "1", per_page: "10", order_by: "asc", sort_by: "name" }))
        .reply(200, mockApiKeys, {
          "x-total-count": "1",
        });

      await store.fetchApiKeys({
        page: 1,
        perPage: 10,
        sortField: "name",
        sortOrder: "asc",
      });

      expect(store.apiKeys).toEqual(mockApiKeys);
      expect(store.apiKeysCount).toBe(1);
    });

    it("should handle network errors", async () => {
      mockApiKeysApi
        .onGet(buildUrl(baseUrl, { page: "1", per_page: "10" }))
        .networkError();

      await expect(
        store.fetchApiKeys({ page: 1, perPage: 10 }),
      ).rejects.toThrow();
    });
  });

  describe("generateApiKey", () => {
    const generateApiKeyUrl = "http://localhost:3000/api/namespaces/api-key";

    it("should generate API key successfully", async () => {
      const mockGeneratedId = "c629572a-b643-4301-90fe-4572b00d007e";
      const requestData = {
        name: "dev",
        expires_in: 30,
        role: "administrator",
      };

      mockApiKeysApi
        .onPost(generateApiKeyUrl)
        .reply(200, { id: mockGeneratedId });

      const result = await store.generateApiKey(requestData);

      expect(result).toBe(mockGeneratedId);
    });

    it("should handle validation errors when generating", async () => {
      const requestData = {
        name: "",
        expires_in: -1,
        role: "invalid",
      };

      mockApiKeysApi
        .onPost(generateApiKeyUrl)
        .reply(400, { message: "Invalid request data" });

      await expect(
        store.generateApiKey(requestData),
      ).rejects.toBeAxiosErrorWithStatus(400);
    });
  });

  describe("editApiKey", () => {
    const generateEditApiKeyUrl = (key: string) => `http://localhost:3000/api/namespaces/api-key/${key}`;
    it("should edit API key name successfully", async () => {
      const editData = {
        key: "test-key-123",
        name: "updated name",
        role: "administrator",
      };

      mockApiKeysApi
        .onPatch(generateEditApiKeyUrl(editData.key))
        .reply(200);

      await expect(store.editApiKey(editData)).resolves.not.toThrow();
    });

    it("should edit API key role successfully", async () => {
      const editData = {
        key: "test-key-456",
        name: "my key",
        role: "observer",
      };

      mockApiKeysApi
        .onPatch(generateEditApiKeyUrl(editData.key))
        .reply(200);

      await expect(store.editApiKey(editData)).resolves.not.toThrow();
    });

    it("should handle not found error when editing", async () => {
      const editData = {
        key: "non-existent-key",
        name: "test",
        role: "administrator",
      };

      mockApiKeysApi
        .onPatch(generateEditApiKeyUrl(editData.key))
        .reply(404, { message: "API key not found" });

      await expect(
        store.editApiKey(editData),
      ).rejects.toBeAxiosErrorWithStatus(404);
    });
  });

  describe("removeApiKey", () => {
    const generateRemoveApiKeyUrl = (key: string) => `http://localhost:3000/api/namespaces/api-key/${key}`;

    it("should remove API key successfully", async () => {
      const removeData = {
        key: "test-key-to-remove",
      };

      mockApiKeysApi
        .onDelete(generateRemoveApiKeyUrl(removeData.key))
        .reply(200);

      await expect(store.removeApiKey(removeData)).resolves.not.toThrow();
    });

    it("should handle not found error when removing", async () => {
      const removeData = {
        key: "non-existent-key",
      };

      mockApiKeysApi
        .onDelete(generateRemoveApiKeyUrl(removeData.key))
        .reply(404, { message: "API key not found" });

      await expect(
        store.removeApiKey(removeData),
      ).rejects.toBeAxiosErrorWithStatus(404);
    });
  });
});
