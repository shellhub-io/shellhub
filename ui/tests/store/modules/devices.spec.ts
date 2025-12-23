import { describe, expect, it, beforeEach, afterEach } from "vitest";
import MockAdapter from "axios-mock-adapter";
import { createPinia, setActivePinia } from "pinia";
import { devicesApi, billingApi } from "@/api/http";
import { IDevice } from "@/interfaces/IDevice";
import useDevicesStore from "@/store/modules/devices";

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
    it("should fetch device list successfully with pagination", async () => {
      const mockDevices = [
        mockDeviceBase,
        { ...mockDeviceBase, uid: "b693c58b53e", name: "device-2", online: false },
      ];

      mockDevicesApi
        .onGet("http://localhost:3000/api/devices?page=1&per_page=10&status=accepted")
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
        .onGet("http://localhost:3000/api/devices?page=1&per_page=10&status=accepted")
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
        .onGet("http://localhost:3000/api/devices?filter=test&page=1&per_page=10&status=accepted")
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
        .onGet("http://localhost:3000/api/devices?page=1&per_page=10&status=accepted&sort_by=name&order_by=asc")
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
        .onGet("http://localhost:3000/api/devices?page=1&per_page=10&status=pending")
        .reply(200, mockDevices, {
          "x-total-count": "1",
        });

      await store.fetchDeviceList({ page: 1, perPage: 10, status: "pending" });

      expect(store.devices).toEqual(mockDevices);
    });

    it("should reset state when request fails with forbidden error", async () => {
      mockDevicesApi
        .onGet("http://localhost:3000/api/devices?page=1&per_page=10&status=accepted")
        .reply(403, { message: "Insufficient permissions" });

      await expect(
        store.fetchDeviceList({ page: 1, perPage: 10, status: "accepted" }),
      ).rejects.toBeAxiosErrorWithStatus(403);

      expect(store.devices).toEqual([]);
      expect(store.deviceCount).toBe(0);
    });

    it("should reset state when request fails with server error", async () => {
      mockDevicesApi
        .onGet("http://localhost:3000/api/devices?page=1&per_page=10&status=accepted")
        .reply(500, { message: "Internal Server Error" });

      await expect(
        store.fetchDeviceList({ page: 1, perPage: 10, status: "accepted" }),
      ).rejects.toBeAxiosErrorWithStatus(500);

      expect(store.devices).toEqual([]);
      expect(store.deviceCount).toBe(0);
    });

    it("should reset state when network error occurs", async () => {
      mockDevicesApi
        .onGet("http://localhost:3000/api/devices?page=1&per_page=10&status=accepted")
        .networkError();

      await expect(
        store.fetchDeviceList({ page: 1, perPage: 10, status: "accepted" }),
      ).rejects.toThrow();

      expect(store.devices).toEqual([]);
      expect(store.deviceCount).toBe(0);
    });
  });

  describe("fetchDevice", () => {
    it("should fetch device by UID successfully", async () => {
      mockDevicesApi
        .onGet("http://localhost:3000/api/devices/resolve?uid=a582b47a42d")
        .reply(200, mockDeviceBase);

      await store.fetchDevice({ uid: "a582b47a42d" });

      expect(store.device).toEqual(mockDeviceBase);
    });

    it("should fetch device by hostname successfully", async () => {
      mockDevicesApi
        .onGet("http://localhost:3000/api/devices/resolve?hostname=test-device")
        .reply(200, mockDeviceBase);

      await store.fetchDevice({ hostname: "test-device" });

      expect(store.device).toEqual(mockDeviceBase);
    });

    it("should handle not found error when fetching device", async () => {
      mockDevicesApi
        .onGet("http://localhost:3000/api/devices/resolve?uid=a582b47a42d")
        .reply(404, { message: "Device not found" });

      await expect(store.fetchDevice({ uid: "a582b47a42d" })).rejects.toBeAxiosErrorWithStatus(404);

      expect(store.device).toEqual({});
    });

    it("should reset device when request fails with server error", async () => {
      mockDevicesApi
        .onGet("http://localhost:3000/api/devices/resolve?uid=a582b47a42d")
        .reply(500, { message: "Internal Server Error" });

      await expect(store.fetchDevice({ uid: "a582b47a42d" })).rejects.toBeAxiosErrorWithStatus(500);

      expect(store.device).toEqual({});
    });

    it("should reset device when network error occurs", async () => {
      mockDevicesApi
        .onGet("http://localhost:3000/api/devices/resolve?uid=a582b47a42d")
        .networkError();

      await expect(store.fetchDevice({ uid: "a582b47a42d" })).rejects.toThrow();

      expect(store.device).toEqual({});
    });
  });

  describe("fetchOnlineDevices", () => {
    it("should fetch online devices successfully", async () => {
      const mockDevices = [
        mockDeviceBase,
        { ...mockDeviceBase, uid: "device-2", online: true },
      ];

      mockDevicesApi
        .onGet("http://localhost:3000/api/devices?page=1&per_page=10&status=accepted")
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
        .onGet("http://localhost:3000/api/devices?page=1&per_page=10&status=accepted")
        .reply(200, mockDevices, {
          "x-total-count": "2",
        });

      await store.fetchOnlineDevices();

      expect(store.onlineDevices).toEqual([mockDeviceBase]);
    });

    it("should fetch online devices with filter", async () => {
      const mockDevices = [mockDeviceBase];

      mockDevicesApi
        .onGet("http://localhost:3000/api/devices?filter=test&page=1&per_page=10&status=accepted")
        .reply(200, mockDevices, {
          "x-total-count": "1",
        });

      await store.fetchOnlineDevices("test");

      expect(store.onlineDevices).toEqual(mockDevices);
    });

    it("should reset onlineDevices when request fails with forbidden error", async () => {
      mockDevicesApi
        .onGet("http://localhost:3000/api/devices?page=1&per_page=10&status=accepted")
        .reply(403, { message: "Insufficient permissions" });

      await expect(store.fetchOnlineDevices()).rejects.toBeAxiosErrorWithStatus(403);

      expect(store.onlineDevices).toEqual([]);
    });

    it("should reset onlineDevices when request fails with server error", async () => {
      mockDevicesApi
        .onGet("http://localhost:3000/api/devices?page=1&per_page=10&status=accepted")
        .reply(500, { message: "Internal Server Error" });

      await expect(store.fetchOnlineDevices()).rejects.toBeAxiosErrorWithStatus(500);

      expect(store.onlineDevices).toEqual([]);
    });

    it("should reset onlineDevices when network error occurs", async () => {
      mockDevicesApi
        .onGet("http://localhost:3000/api/devices?page=1&per_page=10&status=accepted")
        .networkError();

      await expect(store.fetchOnlineDevices()).rejects.toThrow();

      expect(store.onlineDevices).toEqual([]);
    });
  });

  describe("setDeviceListVisibility", () => {
    it("should set showDevices to true when devices exist", async () => {
      mockDevicesApi
        .onGet("http://localhost:3000/api/devices?page=1&per_page=1")
        .reply(200, [mockDeviceBase], {
          "x-total-count": "1",
        });

      await store.setDeviceListVisibility();

      expect(store.showDevices).toBe(true);
    });

    it("should keep showDevices false when no devices exist", async () => {
      mockDevicesApi
        .onGet("http://localhost:3000/api/devices?page=1&per_page=1")
        .reply(200, [], {
          "x-total-count": "0",
        });

      await store.setDeviceListVisibility();

      expect(store.showDevices).toBe(false);
    });
  });

  describe("acceptDevice", () => {
    it("should accept device successfully", async () => {
      mockDevicesApi
        .onPatch("http://localhost:3000/api/devices/a582b47a42d/accept")
        .reply(200);

      await expect(store.acceptDevice("a582b47a42d")).resolves.not.toThrow();
    });

    it("should handle not found error when accepting device", async () => {
      mockDevicesApi
        .onPatch("http://localhost:3000/api/devices/a582b47a42d/accept")
        .reply(404, { message: "Device not found" });

      await expect(
        store.acceptDevice("a582b47a42d"),
      ).rejects.toBeAxiosErrorWithStatus(404);
    });

    it("should handle server error when accepting device", async () => {
      mockDevicesApi
        .onPatch("http://localhost:3000/api/devices/a582b47a42d/accept")
        .reply(500, { message: "Internal Server Error" });

      await expect(
        store.acceptDevice("a582b47a42d"),
      ).rejects.toBeAxiosErrorWithStatus(500);
    });

    it("should handle network error when accepting device", async () => {
      mockDevicesApi
        .onPatch("http://localhost:3000/api/devices/a582b47a42d/accept")
        .networkError();

      await expect(store.acceptDevice("a582b47a42d")).rejects.toThrow();
    });
  });

  describe("rejectDevice", () => {
    it("should reject device successfully", async () => {
      mockDevicesApi
        .onPatch("http://localhost:3000/api/devices/a582b47a42d/reject")
        .reply(200);

      await expect(store.rejectDevice("a582b47a42d")).resolves.not.toThrow();
    });

    it("should handle not found error when rejecting device", async () => {
      mockDevicesApi
        .onPatch("http://localhost:3000/api/devices/a582b47a42d/reject")
        .reply(404, { message: "Device not found" });

      await expect(
        store.rejectDevice("a582b47a42d"),
      ).rejects.toBeAxiosErrorWithStatus(404);
    });

    it("should handle server error when rejecting device", async () => {
      mockDevicesApi
        .onPatch("http://localhost:3000/api/devices/a582b47a42d/reject")
        .reply(500, { message: "Internal Server Error" });

      await expect(
        store.rejectDevice("a582b47a42d"),
      ).rejects.toBeAxiosErrorWithStatus(500);
    });

    it("should handle network error when rejecting device", async () => {
      mockDevicesApi
        .onPatch("http://localhost:3000/api/devices/a582b47a42d/reject")
        .networkError();

      await expect(store.rejectDevice("a582b47a42d")).rejects.toThrow();
    });
  });

  describe("removeDevice", () => {
    it("should remove device successfully", async () => {
      mockDevicesApi
        .onDelete("http://localhost:3000/api/devices/a582b47a42d")
        .reply(200);

      await expect(store.removeDevice("a582b47a42d")).resolves.not.toThrow();
    });

    it("should handle permission error when removing device", async () => {
      mockDevicesApi
        .onDelete("http://localhost:3000/api/devices/a582b47a42d")
        .reply(403, { message: "Insufficient permissions" });

      await expect(
        store.removeDevice("a582b47a42d"),
      ).rejects.toBeAxiosErrorWithStatus(403);
    });

    it("should handle server error when removing device", async () => {
      mockDevicesApi
        .onDelete("http://localhost:3000/api/devices/a582b47a42d")
        .reply(500, { message: "Internal Server Error" });

      await expect(
        store.removeDevice("a582b47a42d"),
      ).rejects.toBeAxiosErrorWithStatus(500);
    });

    it("should handle network error when removing device", async () => {
      mockDevicesApi
        .onDelete("http://localhost:3000/api/devices/a582b47a42d")
        .networkError();

      await expect(store.removeDevice("a582b47a42d")).rejects.toThrow();
    });
  });

  describe("renameDevice", () => {
    beforeEach(() => {
      store.device = mockDeviceBase;
    });

    it("should rename device successfully", async () => {
      const renameData = {
        uid: "a582b47a42d",
        name: { name: "updated-device-name" },
      };

      mockDevicesApi
        .onPut("http://localhost:3000/api/devices/a582b47a42d")
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
        .onPut("http://localhost:3000/api/devices/a582b47a42d")
        .reply(400, { message: "Invalid device name" });

      await expect(
        store.renameDevice(renameData),
      ).rejects.toBeAxiosErrorWithStatus(400);
    });

    it("should handle server error when renaming device", async () => {
      const renameData = {
        uid: "a582b47a42d",
        name: { name: "updated-name" },
      };

      mockDevicesApi
        .onPut("http://localhost:3000/api/devices/a582b47a42d")
        .reply(500, { message: "Internal Server Error" });

      await expect(
        store.renameDevice(renameData),
      ).rejects.toBeAxiosErrorWithStatus(500);
    });

    it("should handle network error when renaming device", async () => {
      const renameData = {
        uid: "a582b47a42d",
        name: { name: "updated-name" },
      };

      mockDevicesApi
        .onPut("http://localhost:3000/api/devices/a582b47a42d")
        .networkError();

      await expect(store.renameDevice(renameData)).rejects.toThrow();
    });
  });

  describe("getFirstPendingDevice", () => {
    it("should fetch first pending device successfully", async () => {
      const pendingDevice = { ...mockDeviceBase, status: "pending" };

      mockDevicesApi
        .onGet("http://localhost:3000/api/devices?page=1&per_page=1&status=pending")
        .reply(200, [pendingDevice], {
          "x-total-count": "1",
        });

      const result = await store.getFirstPendingDevice();

      expect(result).toEqual(pendingDevice);
    });

    it("should return undefined when no pending devices exist", async () => {
      mockDevicesApi
        .onGet("http://localhost:3000/api/devices?page=1&per_page=1&status=pending")
        .reply(200, [], {
          "x-total-count": "0",
        });

      const result = await store.getFirstPendingDevice();

      expect(result).toBeUndefined();
    });
  });

  describe("fetchMostUsedDevices", () => {
    it("should fetch most used devices successfully", async () => {
      const mockDevices = [
        mockDeviceBase,
        { ...mockDeviceBase, uid: "device-2", name: "device-2" },
      ];

      mockBillingApi
        .onGet("http://localhost:3000/api/billing/devices-most-used")
        .reply(200, mockDevices);

      await store.fetchMostUsedDevices();

      expect(store.suggestedDevices).toEqual(mockDevices);
    });

    it("should handle empty most used devices list", async () => {
      mockBillingApi
        .onGet("http://localhost:3000/api/billing/devices-most-used")
        .reply(200, []);

      await store.fetchMostUsedDevices();

      expect(store.suggestedDevices).toEqual([]);
    });

    it("should reset suggestedDevices when request fails with forbidden error", async () => {
      mockBillingApi
        .onGet("http://localhost:3000/api/billing/devices-most-used")
        .reply(403, { message: "Insufficient permissions" });

      await expect(store.fetchMostUsedDevices()).rejects.toBeAxiosErrorWithStatus(403);

      expect(store.suggestedDevices).toEqual([]);
    });

    it("should reset suggestedDevices when request fails with server error", async () => {
      mockBillingApi
        .onGet("http://localhost:3000/api/billing/devices-most-used")
        .reply(500, { message: "Internal Server Error" });

      await expect(store.fetchMostUsedDevices()).rejects.toBeAxiosErrorWithStatus(500);

      expect(store.suggestedDevices).toEqual([]);
    });

    it("should reset suggestedDevices when network error occurs", async () => {
      mockBillingApi
        .onGet("http://localhost:3000/api/billing/devices-most-used")
        .networkError();

      await expect(store.fetchMostUsedDevices()).rejects.toThrow();

      expect(store.suggestedDevices).toEqual([]);
    });
  });

  describe("sendDeviceChoices", () => {
    it("should send device choices successfully", async () => {
      const selectedDevices = [
        mockDeviceBase,
        { ...mockDeviceBase, uid: "device-2" },
      ];

      mockBillingApi
        .onPost("http://localhost:3000/api/billing/device-choice")
        .reply(200);

      await expect(store.sendDeviceChoices(selectedDevices)).resolves.not.toThrow();
    });

    it("should send empty device choices", async () => {
      mockBillingApi
        .onPost("http://localhost:3000/api/billing/device-choice")
        .reply(200);

      await expect(store.sendDeviceChoices([])).resolves.not.toThrow();
    });

    it("should handle validation error when sending device choices", async () => {
      mockBillingApi
        .onPost("http://localhost:3000/api/billing/device-choice")
        .reply(400, { message: "Invalid device choices" });

      await expect(
        store.sendDeviceChoices([mockDeviceBase]),
      ).rejects.toBeAxiosErrorWithStatus(400);
    });

    it("should handle server error when sending device choices", async () => {
      mockBillingApi
        .onPost("http://localhost:3000/api/billing/device-choice")
        .reply(500, { message: "Internal Server Error" });

      await expect(
        store.sendDeviceChoices([mockDeviceBase]),
      ).rejects.toBeAxiosErrorWithStatus(500);
    });

    it("should handle network error when sending device choices", async () => {
      mockBillingApi
        .onPost("http://localhost:3000/api/billing/device-choice")
        .networkError();

      await expect(store.sendDeviceChoices([mockDeviceBase])).rejects.toThrow();
    });
  });
});
