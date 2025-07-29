import { describe, expect, it } from "vitest";
import MockAdapter from "axios-mock-adapter";
import { flushPromises } from "@vue/test-utils";
import { createPinia, setActivePinia } from "pinia";
import { apiKeysApi } from "@/api/http";
import useApiKeysStore from "@/store/modules/api_keys";

describe("apiKeys Pinia Store", () => {
  setActivePinia(createPinia());
  const mockApiKeysApi = new MockAdapter(apiKeysApi.getAxios());
  const apiKeysStore = useApiKeysStore();

  it("should return the default API keys variables", () => {
    expect(apiKeysStore.apiKeys).toEqual([]);
    expect(apiKeysStore.apiKeysCount).toEqual(0);
  });

  it("should generate API key", async () => {
    const generateApiResponse = { id: "c629572a-b643-4301-90fe-4572b00d007e" };
    const generateApiData = {
      name: "dev",
      expires_in: 30,
      role: "owner",
    };

    mockApiKeysApi.onPost("http://localhost:3000/api/namespaces/api-key").reply(200, generateApiResponse);

    const result = await apiKeysStore.generateApiKey(generateApiData);
    await flushPromises();

    expect(result).toEqual(generateApiResponse.id);
  });

  it("should fetch API keys", async () => {
    const getApiResponse = [
      {
        id: "3e5a5194-9dec-4a32-98db-7434c6d49df1",
        tenant_id: "fake-tenant",
        name: "my api key",
        expires_in: 1707958989,
      },
    ];

    const fetchParams = {
      page: 1,
      perPage: 10,
    };

    mockApiKeysApi.onGet("http://localhost:3000/api/namespaces/api-key?page=1&per_page=10").reply(200, getApiResponse, {
      "x-total-count": 1,
    });

    await apiKeysStore.fetchApiKeys(fetchParams);

    expect(apiKeysStore.apiKeys).toEqual(getApiResponse);
    expect(apiKeysStore.apiKeysCount).toEqual(1);
  });

  it("should handle empty API keys response", async () => {
    const fetchParams = {
      page: 1,
      perPage: 10,
    };

    mockApiKeysApi.onGet("http://localhost:3000/api/namespaces/api-key?page=1&per_page=10").reply(200, [], { "x-total-count": 0 });

    await apiKeysStore.fetchApiKeys(fetchParams);

    expect(apiKeysStore.apiKeys).toEqual([]);
    expect(apiKeysStore.apiKeysCount).toEqual(0);
  });

  it("should edit API key", async () => {
    const editApiData = {
      key: "test-key",
      name: "updated name",
      role: "administrator",
    };

    mockApiKeysApi.onPatch("http://localhost:3000/api/namespaces/api-key/test-key").reply(200);

    await expect(apiKeysStore.editApiKey(editApiData)).resolves.not.toThrow();
  });

  it("should remove API key", async () => {
    const removeApiData = {
      key: "test-key",
    };

    mockApiKeysApi.onDelete("http://localhost:3000/api/namespaces/api-key/test-key").reply(200);

    await expect(apiKeysStore.removeApiKey(removeApiData)).resolves.not.toThrow();
  });
});
