import { createPinia, setActivePinia } from "pinia";
import MockAdapter from "axios-mock-adapter";
import { describe, expect, it, beforeEach, afterEach } from "vitest";
import { adminApi } from "@/api/http";
import useNamespacesStore from "@admin/store/modules/namespaces";
import { IAdminNamespace } from "@admin/interfaces/INamespace";
import { buildUrl } from "@tests/utils/url";

const mockNamespaceBase: IAdminNamespace = {
  name: "admin-namespace",
  owner: "admin-user-id",
  type: "personal",
  devices_accepted_count: 5,
  devices_pending_count: 1,
  devices_rejected_count: 0,
  tenant_id: "tenant-id-123",
  members: [
    {
      id: "member-id-1",
      email: "admin@example.com",
      role: "owner",
      added_at: "2026-01-06T00:00:00.000Z",
      expires_at: "2027-01-06T00:00:00.000Z",
    },
  ],
  settings: {
    session_record: true,
    connection_announcement: "",
  },
  created_at: "2026-01-06T00:00:00.000Z",
  max_devices: 10,
};

describe("Admin Namespaces Store", () => {
  let namespacesStore: ReturnType<typeof useNamespacesStore>;
  let mockAdminApi: MockAdapter;

  beforeEach(() => {
    setActivePinia(createPinia());
    namespacesStore = useNamespacesStore();
    mockAdminApi = new MockAdapter(adminApi.getAxios());
  });

  afterEach(() => { mockAdminApi.reset(); });

  describe("Initial State", () => {
    it("should have empty namespaces array", () => {
      expect(namespacesStore.namespaces).toEqual([]);
    });

    it("should have zero namespace count", () => {
      expect(namespacesStore.namespaceCount).toBe(0);
    });

    it("should have empty namespace object", () => {
      expect(namespacesStore.namespace).toEqual({});
    });

    it("should have empty current filter", () => {
      expect(namespacesStore.currentFilter).toBe("");
    });
  });

  describe("setFilter", () => {
    it("should set filter value", () => {
      namespacesStore.setFilter("owner:admin");
      expect(namespacesStore.currentFilter).toBe("owner:admin");
    });

    it("should set empty string when filter is empty", () => {
      namespacesStore.setFilter("");
      expect(namespacesStore.currentFilter).toBe("");
    });
  });

  describe("fetchNamespaceList", () => {
    const baseUrl = "http://localhost:3000/admin/api/namespaces";

    it("should fetch namespaces list successfully with default pagination", async () => {
      const namespacesList = [mockNamespaceBase];

      mockAdminApi.onGet(buildUrl(baseUrl, { filter: "", page: "1", per_page: "10" })).reply(200, namespacesList, { "x-total-count": "1" });

      await expect(namespacesStore.fetchNamespaceList()).resolves.not.toThrow();

      expect(namespacesStore.namespaces).toEqual(namespacesList);
      expect(namespacesStore.namespaceCount).toBe(1);
    });

    it("should fetch namespaces list successfully with custom pagination", async () => {
      const namespacesList = [mockNamespaceBase];

      mockAdminApi.onGet(buildUrl(baseUrl, { filter: "", page: "2", per_page: "20" })).reply(200, namespacesList, { "x-total-count": "1" });

      await expect(namespacesStore.fetchNamespaceList({ page: 2, perPage: 20 })).resolves.not.toThrow();

      expect(namespacesStore.namespaces).toEqual(namespacesList);
      expect(namespacesStore.namespaceCount).toBe(1);
    });

    it("should fetch namespaces list with filter successfully", async () => {
      const namespacesList = [mockNamespaceBase];
      const filter = "test";

      mockAdminApi.onGet(buildUrl(baseUrl, { filter, page: "1", per_page: "10" })).reply(200, namespacesList, { "x-total-count": "1" });

      await expect(namespacesStore.fetchNamespaceList({ filter })).resolves.not.toThrow();

      expect(namespacesStore.namespaces).toEqual(namespacesList);
      expect(namespacesStore.namespaceCount).toBe(1);
    });

    it("should use current filter when not provided in parameters", async () => {
      namespacesStore.setFilter("old_filter");

      const namespacesList = [mockNamespaceBase];

      mockAdminApi
        .onGet(buildUrl(baseUrl, { filter: "old_filter", page: "1", per_page: "10" }))
        .reply(200, namespacesList, { "x-total-count": "1" });

      await expect(namespacesStore.fetchNamespaceList()).resolves.not.toThrow();

      expect(namespacesStore.namespaces).toEqual(namespacesList);
      expect(namespacesStore.namespaceCount).toBe(1);
    });

    it("should fetch empty namespaces list successfully", async () => {
      mockAdminApi.onGet(buildUrl(baseUrl, { filter: "", page: "1", per_page: "10" })).reply(200, [], { "x-total-count": "0" });

      await expect(namespacesStore.fetchNamespaceList()).resolves.not.toThrow();

      expect(namespacesStore.namespaces).toEqual([]);
      expect(namespacesStore.namespaceCount).toBe(0);
    });

    it("should throw on server error when fetching namespaces list", async () => {
      mockAdminApi.onGet(buildUrl(baseUrl, { filter: "", page: "1", per_page: "10" })).reply(500);

      await expect(namespacesStore.fetchNamespaceList()).rejects.toBeAxiosErrorWithStatus(500);
    });

    it("should throw on network error when fetching namespaces list", async () => {
      mockAdminApi.onGet(buildUrl(baseUrl, { filter: "", page: "1", per_page: "10" })).networkError();

      await expect(namespacesStore.fetchNamespaceList()).rejects.toThrow("Network Error");
    });
  });

  describe("fetchNamespaceById", () => {
    const namespaceId = "tenant-id-123";
    const baseGetNamespaceUrl = `http://localhost:3000/admin/api/namespaces/${namespaceId}`;

    it("should fetch namespace by id successfully and return data", async () => {
      mockAdminApi.onGet(baseGetNamespaceUrl).reply(200, mockNamespaceBase);

      await expect(namespacesStore.fetchNamespaceById(namespaceId)).resolves.not.toThrow();
      expect(namespacesStore.namespace).toEqual(mockNamespaceBase);
    });

    it("should throw on not found error when fetching namespace by id", async () => {
      mockAdminApi.onGet(baseGetNamespaceUrl).reply(404, { message: "Namespace not found" });

      await expect(namespacesStore.fetchNamespaceById(namespaceId)).rejects.toBeAxiosErrorWithStatus(404);
    });

    it("should throw on network error when fetching namespace by id", async () => {
      mockAdminApi.onGet(baseGetNamespaceUrl).networkError();

      await expect(namespacesStore.fetchNamespaceById(namespaceId)).rejects.toThrow("Network Error");
    });
  });

  describe("exportNamespacesToCsv", () => {
    const baseUrl = "http://localhost:3000/admin/api/export/namespaces";
    const csvData = "name,owner,devices\nadmin-namespace,admin,5";

    it("should export namespaces to CSV successfully and return data", async () => {
      const filter = "";

      mockAdminApi.onGet(buildUrl(baseUrl, { filter })).reply(200, csvData);

      const result = await namespacesStore.exportNamespacesToCsv(filter);

      expect(result).toBe(csvData);
    });

    it("should export namespaces with filter to CSV successfully", async () => {
      const filter = "owner:admin";

      mockAdminApi.onGet(buildUrl(baseUrl, { filter })).reply(200, csvData);

      const result = await namespacesStore.exportNamespacesToCsv(filter);

      expect(result).toBe(csvData);
    });

    it("should throw on not found error when exporting namespaces", async () => {
      const filter = "";

      mockAdminApi.onGet(buildUrl(baseUrl, { filter })).reply(404, { message: "No namespaces to export" });

      await expect(namespacesStore.exportNamespacesToCsv(filter)).rejects.toBeAxiosErrorWithStatus(404);
    });

    it("should throw on network error when exporting namespaces", async () => {
      const filter = "";

      mockAdminApi.onGet(buildUrl(baseUrl, { filter })).networkError();

      await expect(namespacesStore.exportNamespacesToCsv(filter)).rejects.toThrow("Network Error");
    });
  });

  describe("updateNamespace", () => {
    const baseUrl = `http://localhost:3000/admin/api/namespaces-update/${mockNamespaceBase.tenant_id}`;

    it("should update namespace successfully", async () => {
      mockAdminApi.onPut(baseUrl, mockNamespaceBase).reply(200);

      await expect(namespacesStore.updateNamespace(mockNamespaceBase)).resolves.not.toThrow();
    });

    it("should throw on not found error when updating namespace", async () => {
      mockAdminApi.onPut(baseUrl, mockNamespaceBase).reply(404, { message: "Namespace not found" });

      await expect(namespacesStore.updateNamespace(mockNamespaceBase)).rejects.toBeAxiosErrorWithStatus(404);
    });

    it("should throw on network error when updating namespace", async () => {
      mockAdminApi.onPut(baseUrl, mockNamespaceBase).networkError();

      await expect(namespacesStore.updateNamespace(mockNamespaceBase)).rejects.toThrow("Network Error");
    });
  });

  describe("deleteNamespace", () => {
    const tenantId = "tenant-id-123";
    const baseUrl = `http://localhost:3000/admin/api/namespaces/${tenantId}`;

    it("should delete namespace successfully", async () => {
      mockAdminApi.onDelete(baseUrl).reply(200);

      await expect(namespacesStore.deleteNamespace(tenantId)).resolves.not.toThrow();
    });

    it("should throw on not found error when deleting namespace", async () => {
      mockAdminApi.onDelete(baseUrl).reply(404, { message: "Namespace not found" });

      await expect(namespacesStore.deleteNamespace(tenantId)).rejects.toBeAxiosErrorWithStatus(404);
    });

    it("should throw on network error when deleting namespace", async () => {
      mockAdminApi.onDelete(baseUrl).networkError();

      await expect(namespacesStore.deleteNamespace(tenantId)).rejects.toThrow("Network Error");
    });
  });
});
