import { describe, expect, it, beforeEach, afterEach } from "vitest";
import MockAdapter from "axios-mock-adapter";
import { createPinia, setActivePinia } from "pinia";
import { devicesApi, billingApi } from "@/api/http";
import { IDevice } from "@/interfaces/IDevice";
import useDevicesStore from "@/store/modules/devices";
import { buildUrl } from "../../utils/url";

const mockDeviceBase: IDevice = {
  uid: "a582b47a42d",
  name: "test-device",
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

describe("Devices Store", () => {
  let mockDevicesApi: MockAdapter;
  let mockBillingApi: MockAdapter;
  let store: ReturnType<typeof useDevicesStore>;

  beforeEach(() => {
    setActivePinia(createPinia());
    mockDevicesApi = new MockAdapter(devicesApi.getAxios());
    mockBillingApi = new MockAdapter(billingApi.getAxios());
    store = useDevicesStore();
  });

  afterEach(() => {
    mockDevicesApi.reset();
    mockBillingApi.reset();
  });

  describe("Initial State", () => {
    it("should have correct default values", () => {
      expect(store.devices).toEqual([]);
      expect(store.device).toEqual({});
      expect(store.showDevices).toBe(false);
      expect(store.deviceCount).toBe(0);
      expect(store.totalDevicesCount).toBe(0);
      expect(store.onlineDevicesCount).toBe(0);
      expect(store.offlineDevicesCount).toBe(0);
      expect(store.pendingDevicesCount).toBe(0);
      expect(store.showDeviceChooser).toBe(false);
      expect(store.suggestedDevices).toEqual([]);
      expect(store.selectedDevices).toEqual([]);
      expect(store.duplicatedDeviceName).toBe("");
      expect(store.deviceListFilter).toBeUndefined();
      expect(store.onlineDevices).toEqual([]);
    });
  });

  describe("fetchDeviceList", () => {
    const baseUrl = "http://localhost:3000/api/devices";

    it("should fetch device list successfully with pagination", async () => {
      const mockDevices = [
        mockDeviceBase,
        { ...mockDeviceBase, uid: "b693c58b53e", name: "device-2", online: false },
      ];

      mockDevicesApi
        .onGet(buildUrl(baseUrl, { page: "1", per_page: "10", status: "accepted" }))
        .reply(200, mockDevices, {
          "x-total-count": "2",
        });

      await store.fetchDeviceList({ page: 1, perPage: 10, status: "accepted" });

      expect(store.devices).toEqual(mockDevices);
      expect(store.deviceCount).toBe(2);
      expect(store.showDevices).toBe(true);
    });

    it("should handle empty device list", async () => {
      mockDevicesApi
        .onGet(buildUrl(baseUrl, { page: "1", per_page: "10", status: "accepted" }))
        .reply(200, [], {
          "x-total-count": "0",
        });

      await store.fetchDeviceList({ page: 1, perPage: 10, status: "accepted" });

      expect(store.devices).toEqual([]);
      expect(store.deviceCount).toBe(0);
      expect(store.showDevices).toBe(false);
    });

    it("should fetch devices with filter parameter", async () => {
      const mockDevices = [mockDeviceBase];

      mockDevicesApi
        .onGet(buildUrl(baseUrl, { filter: "test", page: "1", per_page: "10", status: "accepted" }))
        .reply(200, mockDevices, {
          "x-total-count": "1",
        });

      await store.fetchDeviceList({ page: 1, perPage: 10, status: "accepted", filter: "test" });

      expect(store.devices).toEqual(mockDevices);
      expect(store.deviceCount).toBe(1);
      expect(store.deviceListFilter).toBe("test");
    });

    it("should fetch devices with sorting parameters", async () => {
      const mockDevices = [mockDeviceBase];

      mockDevicesApi
        .onGet(buildUrl(baseUrl, { page: "1", per_page: "10", status: "accepted", sort_by: "name", order_by: "asc" }))
        .reply(200, mockDevices, {
          "x-total-count": "1",
        });

      await store.fetchDeviceList({
        page: 1,
        perPage: 10,
        status: "accepted",
        sortField: "name",
        sortOrder: "asc",
      });

      expect(store.devices).toEqual(mockDevices);
    });

    it("should fetch devices with pending status", async () => {
      const mockDevices = [{ ...mockDeviceBase, status: "pending" }];

      mockDevicesApi
        .onGet(buildUrl(baseUrl, { page: "1", per_page: "10", status: "pending" }))
        .reply(200, mockDevices, {
          "x-total-count": "1",
        });

      await store.fetchDeviceList({ page: 1, perPage: 10, status: "pending" });

      expect(store.devices).toEqual(mockDevices);
    });

    it("should reset state when request fails with forbidden error", async () => {
      mockDevicesApi
        .onGet(buildUrl(baseUrl, { page: "1", per_page: "10", status: "accepted" }))
        .reply(403, { message: "Insufficient permissions" });

      await expect(
        store.fetchDeviceList({ page: 1, perPage: 10, status: "accepted" }),
      ).rejects.toBeAxiosErrorWithStatus(403);

      expect(store.devices).toEqual([]);
      expect(store.deviceCount).toBe(0);
    });

    it("should reset state when network error occurs", async () => {
      mockDevicesApi
        .onGet(buildUrl(baseUrl, { page: "1", per_page: "10", status: "accepted" }))
        .networkError();

      await expect(
        store.fetchDeviceList({ page: 1, perPage: 10, status: "accepted" }),
      ).rejects.toThrow();

      expect(store.devices).toEqual([]);
      expect(store.deviceCount).toBe(0);
    });
  });

  describe("fetchDevice", () => {
    const baseUrl = "http://localhost:3000/api/devices/resolve";

    it("should fetch device by UID successfully", async () => {
      mockDevicesApi
        .onGet(buildUrl(baseUrl, { uid: "a582b47a42d" }))
        .reply(200, mockDeviceBase);

      await store.fetchDevice({ uid: "a582b47a42d" });

      expect(store.device).toEqual(mockDeviceBase);
    });

    it("should fetch device by hostname successfully", async () => {
      mockDevicesApi
        .onGet(buildUrl(baseUrl, { hostname: "test-device" }))
        .reply(200, mockDeviceBase);

      await store.fetchDevice({ hostname: "test-device" });

      expect(store.device).toEqual(mockDeviceBase);
    });

    it("should handle not found error when fetching device", async () => {
      mockDevicesApi
        .onGet(buildUrl(baseUrl, { uid: "a582b47a42d" }))
        .reply(404, { message: "Device not found" });

      await expect(store.fetchDevice({ uid: "a582b47a42d" })).rejects.toBeAxiosErrorWithStatus(404);

      expect(store.device).toEqual({});
    });

    it("should reset device when network error occurs", async () => {
      mockDevicesApi
        .onGet(buildUrl(baseUrl, { uid: "a582b47a42d" }))
        .networkError();

      await expect(store.fetchDevice({ uid: "a582b47a42d" })).rejects.toThrow();

      expect(store.device).toEqual({});
    });
  });

  describe("fetchOnlineDevices", () => {
    const baseUrl = "http://localhost:3000/api/devices";

    it("should fetch online devices successfully", async () => {
      const mockDevices = [
        mockDeviceBase,
        { ...mockDeviceBase, uid: "device-2", online: true },
      ];

      mockDevicesApi
        .onGet(buildUrl(baseUrl, { page: "1", per_page: "10", status: "accepted" }))
        .reply(200, mockDevices, {
          "x-total-count": "2",
        });

      await store.fetchOnlineDevices();

      expect(store.onlineDevices).toEqual(mockDevices);
    });

    it("should filter out offline devices", async () => {
      const mockDevices = [
        mockDeviceBase,
        { ...mockDeviceBase, uid: "device-2", online: false },
      ];

      mockDevicesApi
        .onGet(buildUrl(baseUrl, { page: "1", per_page: "10", status: "accepted" }))
        .reply(200, mockDevices, {
          "x-total-count": "2",
        });

      await store.fetchOnlineDevices();

      expect(store.onlineDevices).toEqual([mockDeviceBase]);
    });

    it("should fetch online devices with filter", async () => {
      const mockDevices = [mockDeviceBase];

      mockDevicesApi
        .onGet(buildUrl(baseUrl, { filter: "test", page: "1", per_page: "10", status: "accepted" }))
        .reply(200, mockDevices, {
          "x-total-count": "1",
        });

      await store.fetchOnlineDevices("test");

      expect(store.onlineDevices).toEqual(mockDevices);
    });

    it("should reset onlineDevices when request fails with forbidden error", async () => {
      mockDevicesApi
        .onGet(buildUrl(baseUrl, { page: "1", per_page: "10", status: "accepted" }))
        .reply(403, { message: "Insufficient permissions" });

      await expect(store.fetchOnlineDevices()).rejects.toBeAxiosErrorWithStatus(403);

      expect(store.onlineDevices).toEqual([]);
    });

    it("should reset onlineDevices when network error occurs", async () => {
      mockDevicesApi
        .onGet(buildUrl(baseUrl, { page: "1", per_page: "10", status: "accepted" }))
        .networkError();

      await expect(store.fetchOnlineDevices()).rejects.toThrow();

      expect(store.onlineDevices).toEqual([]);
    });
  });

  describe("fetchDeviceCounts", () => {
    const baseUrl = "http://localhost:3000/api/devices";
    const onlineFilter = Buffer.from(JSON.stringify([
      { type: "property", params: { name: "online", operator: "eq", value: true } },
    ])).toString("base64");
    const offlineFilter = Buffer.from(JSON.stringify([
      { type: "property", params: { name: "online", operator: "eq", value: false } },
    ])).toString("base64");

    const acceptedUrl = buildUrl(baseUrl, { page: "1", per_page: "1", status: "accepted" });
    const pendingUrl = buildUrl(baseUrl, { page: "1", per_page: "1", status: "pending" });
    const onlineUrl = buildUrl(baseUrl, { filter: onlineFilter, page: "1", per_page: "1", status: "accepted" });
    const offlineUrl = buildUrl(baseUrl, { filter: offlineFilter, page: "1", per_page: "1", status: "accepted" });

    it("should fetch all device counts successfully", async () => {
      mockDevicesApi
        .onGet(acceptedUrl)
        .reply(200, [], { "x-total-count": "100" });

      mockDevicesApi
        .onGet(pendingUrl)
        .reply(200, [], { "x-total-count": "5" });

      mockDevicesApi
        .onGet(onlineUrl)
        .reply(200, [], { "x-total-count": "60" });

      mockDevicesApi
        .onGet(offlineUrl)
        .reply(200, [], { "x-total-count": "40" });

      await store.fetchDeviceCounts();

      expect(store.totalDevicesCount).toBe(100);
      expect(store.pendingDevicesCount).toBe(5);
      expect(store.onlineDevicesCount).toBe(60);
      expect(store.offlineDevicesCount).toBe(40);
    });

    it("should handle zero counts", async () => {
      mockDevicesApi
        .onGet(acceptedUrl)
        .reply(200, [], { "x-total-count": "0" });

      mockDevicesApi
        .onGet(pendingUrl)
        .reply(200, [], { "x-total-count": "0" });

      mockDevicesApi
        .onGet(onlineUrl)
        .reply(200, [], { "x-total-count": "0" });

      mockDevicesApi
        .onGet(offlineUrl)
        .reply(200, [], { "x-total-count": "0" });

      await store.fetchDeviceCounts();

      expect(store.totalDevicesCount).toBe(0);
      expect(store.pendingDevicesCount).toBe(0);
      expect(store.onlineDevicesCount).toBe(0);
      expect(store.offlineDevicesCount).toBe(0);
    });

    it("should handle error when fetching counts", async () => {
      mockDevicesApi
        .onGet(acceptedUrl)
        .reply(500, { message: "Internal Server Error" });

      await expect(store.fetchDeviceCounts()).rejects.toBeAxiosErrorWithStatus(500);
    });
  });

  describe("setDeviceListVisibility", () => {
    const url = "http://localhost:3000/api/devices?page=1&per_page=1";

    it("should set showDevices to true when devices exist", async () => {
      mockDevicesApi
        .onGet(url)
        .reply(200, [mockDeviceBase], {
          "x-total-count": "1",
        });

      await store.setDeviceListVisibility();

      expect(store.showDevices).toBe(true);
    });

    it("should keep showDevices false when no devices exist", async () => {
      mockDevicesApi
        .onGet(url)
        .reply(200, [], {
          "x-total-count": "0",
        });

      await store.setDeviceListVisibility();

      expect(store.showDevices).toBe(false);
    });
  });

  describe("acceptDevice", () => {
    const baseAcceptUrl = (uid: string) => `http://localhost:3000/api/devices/${uid}/accept`;

    it("should accept device successfully", async () => {
      mockDevicesApi
        .onPatch(baseAcceptUrl("a582b47a42d"))
        .reply(200);

      await expect(store.acceptDevice("a582b47a42d")).resolves.not.toThrow();
    });

    it("should handle not found error when accepting device", async () => {
      mockDevicesApi
        .onPatch(baseAcceptUrl("a582b47a42d"))
        .reply(404, { message: "Device not found" });

      await expect(
        store.acceptDevice("a582b47a42d"),
      ).rejects.toBeAxiosErrorWithStatus(404);
    });

    it("should handle network error when accepting device", async () => {
      mockDevicesApi
        .onPatch(baseAcceptUrl("a582b47a42d"))
        .networkError();

      await expect(store.acceptDevice("a582b47a42d")).rejects.toThrow();
    });
  });

  describe("rejectDevice", () => {
    const baseRejectUrl = (uid: string) => `http://localhost:3000/api/devices/${uid}/reject`;

    it("should reject device successfully", async () => {
      mockDevicesApi
        .onPatch(baseRejectUrl("a582b47a42d"))
        .reply(200);

      await expect(store.rejectDevice("a582b47a42d")).resolves.not.toThrow();
    });

    it("should handle not found error when rejecting device", async () => {
      mockDevicesApi
        .onPatch(baseRejectUrl("a582b47a42d"))
        .reply(404, { message: "Device not found" });

      await expect(
        store.rejectDevice("a582b47a42d"),
      ).rejects.toBeAxiosErrorWithStatus(404);
    });

    it("should handle network error when rejecting device", async () => {
      mockDevicesApi
        .onPatch(baseRejectUrl("a582b47a42d"))
        .networkError();

      await expect(store.rejectDevice("a582b47a42d")).rejects.toThrow();
    });
  });

  describe("removeDevice", () => {
    const baseRemoveUrl = (uid: string) => `http://localhost:3000/api/devices/${uid}`;

    it("should remove device successfully", async () => {
      mockDevicesApi
        .onDelete(baseRemoveUrl("a582b47a42d"))
        .reply(200);

      await expect(store.removeDevice("a582b47a42d")).resolves.not.toThrow();
    });

    it("should handle permission error when removing device", async () => {
      mockDevicesApi
        .onDelete(baseRemoveUrl("a582b47a42d"))
        .reply(403, { message: "Insufficient permissions" });

      await expect(
        store.removeDevice("a582b47a42d"),
      ).rejects.toBeAxiosErrorWithStatus(403);
    });

    it("should handle network error when removing device", async () => {
      mockDevicesApi
        .onDelete(baseRemoveUrl("a582b47a42d"))
        .networkError();

      await expect(store.removeDevice("a582b47a42d")).rejects.toThrow();
    });
  });

  describe("renameDevice", () => {
    const baseRenameUrl = (uid: string) => `http://localhost:3000/api/devices/${uid}`;

    beforeEach(() => {
      store.device = mockDeviceBase;
    });

    it("should rename device successfully", async () => {
      const renameData = {
        uid: "a582b47a42d",
        name: { name: "updated-device-name" },
      };

      mockDevicesApi
        .onPut(baseRenameUrl("a582b47a42d"))
        .reply(200);

      await store.renameDevice(renameData);

      expect(store.device.name).toBe("updated-device-name");
    });

    it("should handle validation error when renaming device", async () => {
      const renameData = {
        uid: "a582b47a42d",
        name: { name: "" },
      };

      mockDevicesApi
        .onPut(baseRenameUrl("a582b47a42d"))
        .reply(400, { message: "Invalid device name" });

      await expect(
        store.renameDevice(renameData),
      ).rejects.toBeAxiosErrorWithStatus(400);
    });

    it("should handle network error when renaming device", async () => {
      const renameData = {
        uid: "a582b47a42d",
        name: { name: "updated-name" },
      };

      mockDevicesApi
        .onPut(baseRenameUrl("a582b47a42d"))
        .networkError();

      await expect(store.renameDevice(renameData)).rejects.toThrow();
    });
  });

  describe("getFirstPendingDevice", () => {
    const pendingDeviceUrl = "http://localhost:3000/api/devices?page=1&per_page=1&status=pending";

    it("should fetch first pending device successfully", async () => {
      const pendingDevice = { ...mockDeviceBase, status: "pending" };

      mockDevicesApi
        .onGet(pendingDeviceUrl)
        .reply(200, [pendingDevice], {
          "x-total-count": "1",
        });

      const result = await store.getFirstPendingDevice();

      expect(result).toEqual(pendingDevice);
    });

    it("should return undefined when no pending devices exist", async () => {
      mockDevicesApi
        .onGet(pendingDeviceUrl)
        .reply(200, [], {
          "x-total-count": "0",
        });

      const result = await store.getFirstPendingDevice();

      expect(result).toBeUndefined();
    });
  });

  describe("fetchMostUsedDevices", () => {
    const mostUsedDevicesUrl = "http://localhost:3000/api/billing/devices-most-used";

    it("should fetch most used devices successfully", async () => {
      const mockDevices = [
        mockDeviceBase,
        { ...mockDeviceBase, uid: "device-2", name: "device-2" },
      ];

      mockBillingApi
        .onGet(mostUsedDevicesUrl)
        .reply(200, mockDevices);

      await store.fetchMostUsedDevices();

      expect(store.suggestedDevices).toEqual(mockDevices);
    });

    it("should handle empty most used devices list", async () => {
      mockBillingApi
        .onGet(mostUsedDevicesUrl)
        .reply(200, []);

      await store.fetchMostUsedDevices();

      expect(store.suggestedDevices).toEqual([]);
    });

    it("should reset suggestedDevices when request fails with forbidden error", async () => {
      mockBillingApi
        .onGet(mostUsedDevicesUrl)
        .reply(403, { message: "Insufficient permissions" });

      await expect(store.fetchMostUsedDevices()).rejects.toBeAxiosErrorWithStatus(403);

      expect(store.suggestedDevices).toEqual([]);
    });

    it("should reset suggestedDevices when network error occurs", async () => {
      mockBillingApi
        .onGet(mostUsedDevicesUrl)
        .networkError();

      await expect(store.fetchMostUsedDevices()).rejects.toThrow();

      expect(store.suggestedDevices).toEqual([]);
    });
  });

  describe("sendDeviceChoices", () => {
    const deviceChoiceUrl = "http://localhost:3000/api/billing/device-choice";

    it("should send device choices successfully", async () => {
      const selectedDevices = [
        mockDeviceBase,
        { ...mockDeviceBase, uid: "device-2" },
      ];

      mockBillingApi
        .onPost(deviceChoiceUrl)
        .reply(200);

      await expect(store.sendDeviceChoices(selectedDevices)).resolves.not.toThrow();
    });

    it("should send empty device choices", async () => {
      mockBillingApi
        .onPost(deviceChoiceUrl)
        .reply(200);

      await expect(store.sendDeviceChoices([])).resolves.not.toThrow();
    });

    it("should handle validation error when sending device choices", async () => {
      mockBillingApi
        .onPost(deviceChoiceUrl)
        .reply(400, { message: "Invalid device choices" });

      await expect(
        store.sendDeviceChoices([mockDeviceBase]),
      ).rejects.toBeAxiosErrorWithStatus(400);
    });

    it("should handle network error when sending device choices", async () => {
      mockBillingApi
        .onPost(deviceChoiceUrl)
        .networkError();

      await expect(store.sendDeviceChoices([mockDeviceBase])).rejects.toThrow();
    });
  });
});
