import { createPinia, setActivePinia } from "pinia";
import { describe, expect, it, beforeEach, afterEach, vi } from "vitest";
import MockAdapter from "axios-mock-adapter";
import { namespacesApi } from "@/api/http";
import useSupportStore from "@/store/modules/support";

const mockChatWootReset = vi.fn();

vi.mock("@productdevbook/chatwoot/vue", () => ({
  useChatWoot: () => ({
    reset: mockChatWootReset,
  }),
}));

describe("Support Store", () => {
  let supportStore: ReturnType<typeof useSupportStore>;
  let mockNamespacesApi: MockAdapter;

  beforeEach(() => {
    setActivePinia(createPinia());
    supportStore = useSupportStore();
    mockNamespacesApi = new MockAdapter(namespacesApi.getAxios());
    mockChatWootReset.mockClear();
  });

  afterEach(() => { mockNamespacesApi.reset(); });

  describe("Initial State", () => {
    it("should have empty identifier", () => {
      expect(supportStore.identifier).toBe("");
    });

    it("should have isChatCreated as false", () => {
      expect(supportStore.isChatCreated).toBe(false);
    });
  });

  describe("getIdentifier", () => {
    it("should get identifier successfully and update state", async () => {
      const tenantId = "tenant-id-123";
      const mockResponse = { identifier: "support-identifier-abc" };

      mockNamespacesApi
        .onGet(`http://localhost:3000/api/namespaces/${tenantId}/support`)
        .reply(200, mockResponse);

      await expect(supportStore.getIdentifier(tenantId)).resolves.not.toThrow();

      expect(supportStore.identifier).toBe("support-identifier-abc");
      expect(mockChatWootReset).toHaveBeenCalledTimes(1);
    });

    it("should reset ChatWoot before fetching identifier", async () => {
      const tenantId = "tenant-id-123";
      const mockResponse = { identifier: "support-identifier-abc" };

      mockNamespacesApi
        .onGet(`http://localhost:3000/api/namespaces/${tenantId}/support`)
        .reply(200, mockResponse);

      await supportStore.getIdentifier(tenantId);

      expect(mockChatWootReset).toHaveBeenCalledTimes(1);
      expect(mockNamespacesApi.history.get.length).toBe(1);
    });

    it("should handle different tenant identifiers", async () => {
      const tenantId1 = "tenant-1";
      const tenantId2 = "tenant-2";

      mockNamespacesApi
        .onGet(`http://localhost:3000/api/namespaces/${tenantId1}/support`)
        .reply(200, { identifier: "identifier-1" });

      await supportStore.getIdentifier(tenantId1);
      expect(supportStore.identifier).toBe("identifier-1");

      mockNamespacesApi
        .onGet(`http://localhost:3000/api/namespaces/${tenantId2}/support`)
        .reply(200, { identifier: "identifier-2" });

      await supportStore.getIdentifier(tenantId2);
      expect(supportStore.identifier).toBe("identifier-2");
      expect(mockChatWootReset).toHaveBeenCalledTimes(2);
    });

    it("should handle not found error when getting identifier", async () => {
      const tenantId = "non-existent-tenant";

      mockNamespacesApi
        .onGet(`http://localhost:3000/api/namespaces/${tenantId}/support`)
        .reply(404, { message: "Namespace not found" });

      await expect(supportStore.getIdentifier(tenantId)).rejects.toBeAxiosErrorWithStatus(404);
      expect(mockChatWootReset).toHaveBeenCalledTimes(1);
    });

    it("should handle server error when getting identifier", async () => {
      const tenantId = "tenant-id-123";

      mockNamespacesApi
        .onGet(`http://localhost:3000/api/namespaces/${tenantId}/support`)
        .reply(500);

      await expect(supportStore.getIdentifier(tenantId)).rejects.toBeAxiosErrorWithStatus(500);
      expect(mockChatWootReset).toHaveBeenCalledTimes(1);
    });

    it("should handle network error when getting identifier", async () => {
      const tenantId = "tenant-id-123";

      mockNamespacesApi
        .onGet(`http://localhost:3000/api/namespaces/${tenantId}/support`)
        .networkError();

      await expect(supportStore.getIdentifier(tenantId)).rejects.toThrow("Network Error");
      expect(mockChatWootReset).toHaveBeenCalledTimes(1);
    });
  });

  describe("isChatCreated State", () => {
    it("should update isChatCreated to true", () => {
      supportStore.isChatCreated = true;

      expect(supportStore.isChatCreated).toBe(true);
    });

    it("should update isChatCreated to false", () => {
      supportStore.isChatCreated = true;
      supportStore.isChatCreated = false;

      expect(supportStore.isChatCreated).toBe(false);
    });
  });
});
