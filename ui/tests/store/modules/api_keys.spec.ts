import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import MockAdapter from "axios-mock-adapter";
import { flushPromises } from "@vue/test-utils";
import { store } from "@/store";
import { apiKeysApi } from "@/api/http";

describe("apiKeys Store Actions", () => {
  let mockApiKeys: MockAdapter;

  beforeEach(() => {
    vi.useFakeTimers();
    localStorage.setItem("tenant", "fake-tenant");
    mockApiKeys = new MockAdapter(apiKeysApi.getAxios());
  });

  afterEach(() => {
    vi.useRealTimers();
    vi.restoreAllMocks();
    mockApiKeys.reset();
  });

  describe("Default Values", () => {
    it("should return the default apiKeysentication variables", () => {
      expect(store.getters["apiKeys/getSortStatusField"]).toEqual(undefined);
      expect(store.getters["apiKeys/getSortStatusString"]).toEqual("asc");
      expect(store.getters["apiKeys/apiKey"]).toEqual("");
      expect(store.getters["apiKeys/apiKeyList"]).toEqual([]);
      expect(store.getters["apiKeys/getNumberApiKeys"]).toEqual(0);
    });
  });

  describe("API Key Actions", () => {
    it("should generate API key", async () => {
      const generateApiResponse = { id: "c629572a-b643-4301-90fe-4572b00d007e" };
      const generateApiData = {
        name: "dev",
        expires_at: 30,
        role: "owner",
        key: "c629572a-b643-4301-90fe-4572b00d007e",
      };

      const dispatchSpy = vi.spyOn(store, "dispatch");

      mockApiKeys.onPost("http://localhost:3000/api/namespaces/api-key").reply(200, generateApiResponse);

      await store.dispatch("apiKeys/generateApiKey", generateApiData);
      await flushPromises();

      expect(dispatchSpy).toHaveBeenCalledWith("apiKeys/generateApiKey", generateApiData);
      expect(store.getters["apiKeys/apiKey"]).toEqual(generateApiResponse.id);
    });

    it("should get API keys", async () => {
      const getApiResponse = [
        {
          id: "3e5a5194-9dec-4a32-98db-7434c6d49df1",
          tenant_id: "fake-tenant",
          user_id: "507f1f77bcf86cd799439011",
          name: "my api key",
          expires_in: 1707958989,
        },
      ];

      const getApiData = { tenant: "fake-tenant" };

      const dispatchSpy = vi.spyOn(store, "dispatch");

      mockApiKeys.onGet("http://localhost:3000/api/namespaces/api-key").reply(200, getApiResponse, { "x-total-count": 1 });

      await store.dispatch("apiKeys/getApiKey", getApiData);
      await flushPromises();

      expect(dispatchSpy).toHaveBeenCalledWith("apiKeys/getApiKey", getApiData);
      expect(store.getters["apiKeys/apiKeyList"]).toEqual(getApiResponse);
      expect(store.getters["apiKeys/getNumberApiKeys"]).toEqual(1);
    });
  });
});
