import { createPinia, setActivePinia } from "pinia";
import MockAdapter from "axios-mock-adapter";
import { describe, expect, it, beforeEach, afterEach } from "vitest";
import { adminApi } from "@/api/http";
import useDevicesStore from "@admin/store/modules/devices";
import { IAdminDevice } from "@admin/interfaces/IDevice";
import { buildUrl } from "@tests/utils/url";

const mockDeviceBase: IAdminDevice = {
  uid: "device-uid-123",
  name: "admin-device",
  identity: {
    mac: "00:1A:2B:3C:4D:5E",
  },
  info: {
    id: "debian",
    pretty_name: "Debian GNU/Linux 11",
    version: "11",
    arch: "x86_64",
    platform: "docker",
  },
  public_key: "ssh-rsa AAAAB3NzaC1...",
  tenant_id: "tenant-id-789",
  last_seen: "2026-01-01T12:00:00.000Z",
  status_updated_at: "2026-01-01T12:00:00.000Z",
  online: true,
  namespace: "admin-namespace",
  status: "accepted",
  created_at: "2026-01-01T00:00:00.000Z",
  remote_addr: "192.168.1.100",
  position: { latitude: 0, longitude: 0 },
  tags: [{
    name: "admin",
    tenant_id: "tenant-id-789",
    created_at: "2026-01-01T00:00:00.000Z",
    updated_at: "2026-01-01T00:00:00.000Z",
  }],
};

describe("Admin Devices Store", () => {
  let devicesStore: ReturnType<typeof useDevicesStore>;
  let mockAdminApi: MockAdapter;

  beforeEach(() => {
    setActivePinia(createPinia());
    devicesStore = useDevicesStore();
    mockAdminApi = new MockAdapter(adminApi.getAxios());
  });

  afterEach(() => { mockAdminApi.reset(); });

  describe("Initial State", () => {
    it("should have empty devices array", () => {
      expect(devicesStore.devices).toEqual([]);
    });

    it("should have zero device count", () => {
      expect(devicesStore.deviceCount).toBe(0);
    });

    it("should have empty current filter", () => {
      expect(devicesStore.currentFilter).toBe("");
    });

    it("should have undefined current sort field", () => {
      expect(devicesStore.currentSortField).toBeUndefined();
    });

    it("should have undefined current sort order", () => {
      expect(devicesStore.currentSortOrder).toBeUndefined();
    });
  });

  describe("setFilter", () => {
    it("should set filter value", () => {
      devicesStore.setFilter("status:accepted");
      expect(devicesStore.currentFilter).toBe("status:accepted");
    });

    it("should set empty string when filter is empty", () => {
      devicesStore.setFilter("");
      expect(devicesStore.currentFilter).toBe("");
    });
  });

  describe("setSort", () => {
    it("should set sort field and order", () => {
      devicesStore.setSort("name", "asc");
      expect(devicesStore.currentSortField).toBe("name");
      expect(devicesStore.currentSortOrder).toBe("asc");
    });

    it("should set undefined sort field and order when not provided", () => {
      devicesStore.setSort();
      expect(devicesStore.currentSortField).toBeUndefined();
      expect(devicesStore.currentSortOrder).toBeUndefined();
    });
  });

  describe("fetchDeviceList", () => {
    const baseUrl = "http://localhost:3000/admin/api/devices";

    it("should fetch devices list successfully with default pagination", async () => {
      const devicesList = [mockDeviceBase];

      mockAdminApi
        .onGet(buildUrl(baseUrl, { filter: "", page: "1", per_page: "10" }))
        .reply(200, devicesList, { "x-total-count": "1" });

      await expect(devicesStore.fetchDeviceList()).resolves.not.toThrow();

      expect(devicesStore.devices).toEqual(devicesList);
      expect(devicesStore.deviceCount).toBe(1);
    });

    it("should fetch devices list successfully with custom pagination", async () => {
      const devicesList = [mockDeviceBase];

      mockAdminApi
        .onGet(buildUrl(baseUrl, { filter: "", page: "2", per_page: "20" }))
        .reply(200, devicesList, { "x-total-count": "1" });

      await expect(devicesStore.fetchDeviceList({ page: 2, perPage: 20 })).resolves.not.toThrow();

      expect(devicesStore.devices).toEqual(devicesList);
      expect(devicesStore.deviceCount).toBe(1);
    });

    it("should fetch devices list with filter successfully", async () => {
      const devicesList = [mockDeviceBase];
      const filter = "test";

      mockAdminApi
        .onGet(buildUrl(baseUrl, { filter, page: "1", per_page: "10" }))
        .reply(200, devicesList, { "x-total-count": "1" });

      await expect(devicesStore.fetchDeviceList({ filter })).resolves.not.toThrow();

      expect(devicesStore.devices).toEqual(devicesList);
      expect(devicesStore.deviceCount).toBe(1);
    });

    it("should fetch devices list with sort successfully", async () => {
      const devicesList = [mockDeviceBase];

      mockAdminApi
        .onGet(buildUrl(baseUrl, { filter: "", page: "1", per_page: "10", sort_by: "name", order_by: "asc" }))
        .reply(200, devicesList, { "x-total-count": "1" });

      await expect(devicesStore.fetchDeviceList({ sortField: "name", sortOrder: "asc" })).resolves.not.toThrow();

      expect(devicesStore.devices).toEqual(devicesList);
      expect(devicesStore.deviceCount).toBe(1);
    });

    it("should use current filter and sort when not provided in parameters", async () => {
      devicesStore.setFilter("old_filter");
      devicesStore.setSort("created_at", "desc");

      const devicesList = [mockDeviceBase];

      mockAdminApi
        .onGet(buildUrl(baseUrl, { filter: "old_filter", page: "1", per_page: "10", sort_by: "created_at", order_by: "desc" }))
        .reply(200, devicesList, { "x-total-count": "1" });

      await expect(devicesStore.fetchDeviceList()).resolves.not.toThrow();

      expect(devicesStore.devices).toEqual(devicesList);
      expect(devicesStore.deviceCount).toBe(1);
    });

    it("should fetch empty devices list successfully", async () => {
      mockAdminApi
        .onGet(buildUrl(baseUrl, { filter: "", page: "1", per_page: "10" }))
        .reply(200, [], { "x-total-count": "0" });

      await expect(devicesStore.fetchDeviceList()).resolves.not.toThrow();

      expect(devicesStore.devices).toEqual([]);
      expect(devicesStore.deviceCount).toBe(0);
    });

    it("should throw on server error when fetching devices list", async () => {
      mockAdminApi
        .onGet(buildUrl(baseUrl, { filter: "", page: "1", per_page: "10" }))
        .reply(500);

      await expect(devicesStore.fetchDeviceList()).rejects.toBeAxiosErrorWithStatus(500);
    });

    it("should throw on network error when fetching devices list", async () => {
      mockAdminApi
        .onGet(buildUrl(baseUrl, { filter: "", page: "1", per_page: "10" }))
        .networkError();

      await expect(devicesStore.fetchDeviceList()).rejects.toThrow("Network Error");
    });
  });

  describe("fetchDeviceById", () => {
    const baseGetDeviceUrl = (deviceUid: string) => `http://localhost:3000/admin/api/devices/${deviceUid}`;

    it("should fetch device by id successfully and return data", async () => {
      const deviceUid = "device-uid-123";

      mockAdminApi.onGet(baseGetDeviceUrl(deviceUid)).reply(200, mockDeviceBase);

      const result = await devicesStore.fetchDeviceById(deviceUid);

      expect(result).toEqual(mockDeviceBase);
    });

    it("should throw on not found error when fetching device by id", async () => {
      const deviceUid = "non-existent-device";

      mockAdminApi.onGet(baseGetDeviceUrl(deviceUid)).reply(404, { message: "Device not found" });

      await expect(devicesStore.fetchDeviceById(deviceUid)).rejects.toBeAxiosErrorWithStatus(404);
    });

    it("should throw on network error when fetching device by id", async () => {
      const deviceUid = "device-uid-123";

      mockAdminApi.onGet(baseGetDeviceUrl(deviceUid)).networkError();

      await expect(devicesStore.fetchDeviceById(deviceUid)).rejects.toThrow("Network Error");
    });
  });
});
