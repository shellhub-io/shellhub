import { describe, expect, it, beforeEach, afterEach } from "vitest";
import MockAdapter from "axios-mock-adapter";
import { createPinia, setActivePinia } from "pinia";
import { containersApi } from "@/api/http";
import useContainersStore from "@/store/modules/containers";
import { IContainer } from "@/interfaces/IContainer";
import { buildUrl } from "../../utils/url";

const mockContainerBase: IContainer = {
  uid: "a582b47a42d",
  name: "test-container",
  identity: {
    mac: "00:11:22:33:44:55",
  },
  info: {
    id: "ubuntu",
    pretty_name: "Ubuntu 22.04 LTS",
    version: "22.04",
    arch: "x86_64",
    platform: "docker",
  },
  public_key: "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQ...",
  tenant_id: "tenant-123",
  last_seen: "2023-12-18T10:00:00Z",
  online: true,
  namespace: "production",
  status: "accepted",
  created_at: "2023-01-01T00:00:00Z",
  status_updated_at: "2023-12-18T09:00:00Z",
  remote_addr: "192.168.1.100",
  position: {
    latitude: 0,
    longitude: 0,
  },
  tags: [],
};

describe("Containers Store", () => {
  let mockContainersApi: MockAdapter;
  let store: ReturnType<typeof useContainersStore>;

  beforeEach(() => {
    setActivePinia(createPinia());
    mockContainersApi = new MockAdapter(containersApi.getAxios());
    store = useContainersStore();
  });

  afterEach(() => { mockContainersApi.reset(); });

  describe("Initial State", () => {
    it("should have correct default values", () => {
      expect(store.containers).toEqual([]);
      expect(store.container).toEqual({});
      expect(store.containerCount).toBe(0);
      expect(store.showContainers).toBe(false);
      expect(store.containerListFilter).toBeUndefined();
    });
  });

  describe("fetchContainerList", () => {
    const baseUrl = "http://localhost:3000/api/containers";

    it("should fetch container list successfully with pagination", async () => {
      const mockContainers = [
        mockContainerBase,
        { ...mockContainerBase, uid: "b693c58b53e", name: "container-2", online: false },
      ];

      mockContainersApi
        .onGet(buildUrl(baseUrl, { page: 1, per_page: 10, status: "accepted" }))
        .reply(200, mockContainers, {
          "x-total-count": "2",
        });

      await store.fetchContainerList({ page: 1, perPage: 10, status: "accepted" });

      expect(store.containers).toEqual(mockContainers);
      expect(store.containerCount).toBe(2);
    });

    it("should handle empty container list", async () => {
      mockContainersApi
        .onGet(buildUrl(baseUrl, { page: 1, per_page: 10, status: "accepted" }))
        .reply(200, [], {
          "x-total-count": "0",
        });

      await store.fetchContainerList({ page: 1, perPage: 10, status: "accepted" });

      expect(store.containers).toEqual([]);
      expect(store.containerCount).toBe(0);
    });

    it("should fetch containers with filter parameter", async () => {
      const mockContainers = [mockContainerBase];

      mockContainersApi
        .onGet(buildUrl(baseUrl, { filter: "test", page: 1, per_page: 10, status: "accepted" }))
        .reply(200, mockContainers, {
          "x-total-count": "1",
        });

      await store.fetchContainerList({ page: 1, perPage: 10, status: "accepted", filter: "test" });

      expect(store.containers).toEqual(mockContainers);
      expect(store.containerCount).toBe(1);
      expect(store.containerListFilter).toBe("test");
    });

    it("should fetch containers with sorting parameters", async () => {
      const mockContainers = [mockContainerBase];

      mockContainersApi
        .onGet(buildUrl(baseUrl, { page: 1, per_page: 10, status: "accepted", sort_by: "name", order_by: "asc" }))
        .reply(200, mockContainers, {
          "x-total-count": "1",
        });

      await store.fetchContainerList({
        page: 1,
        perPage: 10,
        status: "accepted",
        sortField: "name",
        sortOrder: "asc",
      });

      expect(store.containers).toEqual(mockContainers);
    });

    it("should fetch containers with pending status", async () => {
      const mockContainers = [{ ...mockContainerBase, status: "pending" }];

      mockContainersApi
        .onGet(buildUrl(baseUrl, { page: 1, per_page: 10, status: "pending" }))
        .reply(200, mockContainers, {
          "x-total-count": "1",
        });

      await store.fetchContainerList({ page: 1, perPage: 10, status: "pending" });

      expect(store.containers).toEqual(mockContainers);
    });

    it("should reset state when request fails with forbidden error", async () => {
      mockContainersApi
        .onGet(buildUrl(baseUrl, { page: 1, per_page: 10, status: "accepted" }))
        .reply(403, { message: "Insufficient permissions" });

      await expect(
        store.fetchContainerList({ page: 1, perPage: 10, status: "accepted" }),
      ).rejects.toBeAxiosErrorWithStatus(403);

      expect(store.containers).toEqual([]);
      expect(store.containerCount).toBe(0);
    });

    it("should reset state when network error occurs", async () => {
      mockContainersApi
        .onGet(buildUrl(baseUrl, { page: 1, perPage: 10, status: "accepted" }))
        .networkError();

      await expect(
        store.fetchContainerList({ page: 1, perPage: 10, status: "accepted" }),
      ).rejects.toThrow();

      expect(store.containers).toEqual([]);
      expect(store.containerCount).toBe(0);
    });
  });

  describe("getContainer", () => {
    const generateGetContainerUrl = (id: string) => `http://localhost:3000/api/containers/${id}`;

    it("should fetch container by ID successfully", async () => {
      mockContainersApi
        .onGet(generateGetContainerUrl("a582b47a42d"))
        .reply(200, mockContainerBase);

      await store.getContainer("a582b47a42d");

      expect(store.container).toEqual(mockContainerBase);
    });

    it("should handle not found error when fetching container", async () => {
      mockContainersApi
        .onGet(generateGetContainerUrl("a582b47a42d"))
        .reply(404, { message: "Container not found" });

      await expect(store.getContainer("a582b47a42d")).rejects.toBeAxiosErrorWithStatus(404);

      expect(store.container).toEqual({});
    });

    it("should reset container when network error occurs", async () => {
      mockContainersApi
        .onGet(generateGetContainerUrl("a582b47a42d"))
        .networkError();

      await expect(store.getContainer("a582b47a42d")).rejects.toThrow();

      expect(store.container).toEqual({});
    });
  });

  describe("setContainerListVisibility", () => {
    const visibilityCheckUrl = "http://localhost:3000/api/containers?page=1&per_page=1";

    it("should set showContainers to true when containers exist", async () => {
      mockContainersApi
        .onGet(visibilityCheckUrl)
        .reply(200, [mockContainerBase], {
          "x-total-count": "1",
        });

      await store.setContainerListVisibility();

      expect(store.showContainers).toBe(true);
    });

    it("should keep showContainers false when no containers exist", async () => {
      mockContainersApi
        .onGet(visibilityCheckUrl)
        .reply(200, [], {
          "x-total-count": "0",
        });

      await store.setContainerListVisibility();

      expect(store.showContainers).toBe(false);
    });
  });

  describe("acceptContainer", () => {
    const generateAcceptUrl = (id: string) => `http://localhost:3000/api/containers/${id}/accept`;

    it("should accept container successfully", async () => {
      mockContainersApi
        .onPatch(generateAcceptUrl("a582b47a42d"))
        .reply(200);

      await expect(store.acceptContainer("a582b47a42d")).resolves.not.toThrow();
    });

    it("should handle not found error when accepting container", async () => {
      mockContainersApi
        .onPatch(generateAcceptUrl("a582b47a42d"))
        .reply(404, { message: "Container not found" });

      await expect(
        store.acceptContainer("a582b47a42d"),
      ).rejects.toBeAxiosErrorWithStatus(404);
    });

    it("should handle network error when accepting container", async () => {
      mockContainersApi
        .onPatch(generateAcceptUrl("a582b47a42d"))
        .networkError();

      await expect(store.acceptContainer("a582b47a42d")).rejects.toThrow();
    });
  });

  describe("rejectContainer", () => {
    const generateRejectUrl = (id: string) => `http://localhost:3000/api/containers/${id}/reject`;

    it("should reject container successfully", async () => {
      mockContainersApi
        .onPatch(generateRejectUrl("a582b47a42d"))
        .reply(200);

      await expect(store.rejectContainer("a582b47a42d")).resolves.not.toThrow();
    });

    it("should handle not found error when rejecting container", async () => {
      mockContainersApi
        .onPatch(generateRejectUrl("a582b47a42d"))
        .reply(404, { message: "Container not found" });

      await expect(
        store.rejectContainer("a582b47a42d"),
      ).rejects.toBeAxiosErrorWithStatus(404);
    });

    it("should handle network error when rejecting container", async () => {
      mockContainersApi
        .onPatch(generateRejectUrl("a582b47a42d"))
        .networkError();

      await expect(store.rejectContainer("a582b47a42d")).rejects.toThrow();
    });
  });

  describe("removeContainer", () => {
    const generateRemoveUrl = (id: string) => `http://localhost:3000/api/containers/${id}`;

    it("should remove container successfully", async () => {
      mockContainersApi
        .onDelete(generateRemoveUrl("a582b47a42d"))
        .reply(200);

      await expect(store.removeContainer("a582b47a42d")).resolves.not.toThrow();
    });

    it("should handle permission error when removing container", async () => {
      mockContainersApi
        .onDelete(generateRemoveUrl("a582b47a42d"))
        .reply(403, { message: "Insufficient permissions" });

      await expect(
        store.removeContainer("a582b47a42d"),
      ).rejects.toBeAxiosErrorWithStatus(403);
    });

    it("should handle network error when removing container", async () => {
      mockContainersApi
        .onDelete(generateRemoveUrl("a582b47a42d"))
        .networkError();

      await expect(store.removeContainer("a582b47a42d")).rejects.toThrow();
    });
  });

  describe("renameContainer", () => {
    const generateRenameUrl = (id: string) => `http://localhost:3000/api/containers/${id}`;

    beforeEach(() => { store.container = mockContainerBase; });

    it("should rename container successfully", async () => {
      const renameData = {
        uid: "a582b47a42d",
        name: { name: "updated-container-name" },
      };

      mockContainersApi
        .onPut(generateRenameUrl("a582b47a42d"))
        .reply(200);

      await store.renameContainer(renameData);

      expect(store.container.name).toBe("updated-container-name");
    });

    it("should handle validation error when renaming container", async () => {
      const renameData = {
        uid: "a582b47a42d",
        name: { name: "" },
      };

      mockContainersApi
        .onPut(generateRenameUrl("a582b47a42d"))
        .reply(400, { message: "Invalid container name" });

      await expect(
        store.renameContainer(renameData),
      ).rejects.toBeAxiosErrorWithStatus(400);
    });

    it("should handle network error when renaming container", async () => {
      const renameData = {
        uid: "a582b47a42d",
        name: { name: "updated-name" },
      };

      mockContainersApi
        .onPut(generateRenameUrl("a582b47a42d"))
        .networkError();

      await expect(store.renameContainer(renameData)).rejects.toThrow();
    });
  });
});
