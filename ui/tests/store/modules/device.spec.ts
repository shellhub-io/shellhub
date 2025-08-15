import { describe, expect, it, beforeEach, afterEach } from "vitest";
import MockAdapter from "axios-mock-adapter";
import { createPinia, setActivePinia } from "pinia";
import { devicesApi, billingApi } from "@/api/http";
import { IDevice } from "@/interfaces/IDevice";
import useDevicesStore from "@/store/modules/devices";

describe("Devices Pinia Store", () => {
  setActivePinia(createPinia());
  const mockDevices = new MockAdapter(devicesApi.getAxios());
  const mockBilling = new MockAdapter(billingApi.getAxios());
  let deviceStore: ReturnType<typeof useDevicesStore>;

  beforeEach(() => {
    deviceStore = useDevicesStore();
  });

  afterEach(() => {
    mockDevices.reset();
    mockBilling.reset();
  });

  describe("initial state", () => {
    it("should have initial state values", () => {
      expect(deviceStore.devices).toEqual([]);
      expect(deviceStore.device).toEqual({});
      expect(deviceStore.showDevices).toBe(false);
      expect(deviceStore.deviceCount).toBe(0);
      expect(deviceStore.showDeviceChooser).toBe(false);
      expect(deviceStore.suggestedDevices).toEqual([]);
      expect(deviceStore.selectedDevices).toEqual([]);
      expect(deviceStore.duplicatedDeviceName).toBe("");
    });
  });

  describe("actions", () => {
    it("should fetch device list successfully", async () => {
      const devicesData = [
        { uid: "a582b47a42d", name: "Device 1" },
        { uid: "a582b47a42e", name: "Device 2" },
      ];

      mockDevices.onGet("http://localhost:3000/api/devices?page=1&per_page=10&status=accepted").reply(200, devicesData, {
        "x-total-count": "2",
      });

      await deviceStore.fetchDeviceList({ page: 1, perPage: 10, status: "accepted" });

      expect(deviceStore.devices).toEqual(devicesData);
      expect(deviceStore.deviceCount).toBe(2);
      expect(deviceStore.showDevices).toBe(true);
    });

    it("should handle empty device list", async () => {
      mockDevices.onGet("http://localhost:3000/api/devices?page=1&per_page=10&status=accepted").reply(200, [], {
        "x-total-count": "0",
      });

      await deviceStore.fetchDeviceList({ page: 1, perPage: 10, status: "accepted" });

      expect(deviceStore.devices).toEqual([]);
      expect(deviceStore.deviceCount).toBe(0);
    });

    it("should remove device", async () => {
      mockDevices.onDelete("http://localhost:3000/api/devices/a582b47a42d").reply(200);

      await expect(deviceStore.removeDevice("a582b47a42d")).resolves.not.toThrow();
    });

    it("should rename device", async () => {
      const renameData = { uid: "a582b47a42d", name: { name: "Updated Device 1" } };

      // Set initial device state
      deviceStore.device = { uid: "a582b47a42d", name: "Device 1" } as IDevice;

      mockDevices.onPut("http://localhost:3000/api/devices/a582b47a42d").reply(200);

      await deviceStore.renameDevice(renameData);

      expect(deviceStore.device.name).toBe("Updated Device 1");
    });

    it("should fetch device by UID", async () => {
      const deviceData = { uid: "a582b47a42d", name: "Device 1" };

      mockDevices.onGet("http://localhost:3000/api/devices/resolve?uid=a582b47a42d").reply(200, deviceData);

      await deviceStore.fetchDevice({ uid: "a582b47a42d" });

      expect(deviceStore.device).toEqual(deviceData);
    });

    it("should fetch device by hostname", async () => {
      const deviceData = { uid: "a582b47a42d", name: "Device1" };

      mockDevices.onGet("http://localhost:3000/api/devices/resolve?hostname=Device1").reply(200, deviceData);

      await deviceStore.fetchDevice({ hostname: "Device1" });

      expect(deviceStore.device).toEqual(deviceData);
    });

    it("should accept device", async () => {
      mockDevices.onPatch("http://localhost:3000/api/devices/a582b47a42d/accept").reply(200);

      await expect(deviceStore.acceptDevice("a582b47a42d")).resolves.not.toThrow();
    });

    it("should reject device", async () => {
      mockDevices.onPatch("http://localhost:3000/api/devices/a582b47a42d/reject").reply(200);

      await expect(deviceStore.rejectDevice("a582b47a42d")).resolves.not.toThrow();
    });

    it("should get first pending device", async () => {
      const deviceData = { uid: "a582b47a42d", name: "Device 1" };

      mockDevices.onGet("http://localhost:3000/api/devices?page=1&per_page=1&status=pending").reply(200, [deviceData]);

      const result = await deviceStore.getFirstPendingDevice();

      expect(result).toEqual(deviceData);
    });

    it("should fetch most used devices", async () => {
      const devicesData = [
        { uid: "a582b47a42d", name: "Device 1" },
        { uid: "a582b47a42e", name: "Device 2" },
      ];

      mockBilling.onGet("http://localhost:3000/api/billing/devices-most-used").reply(200, devicesData);

      await deviceStore.fetchMostUsedDevices();

      expect(deviceStore.suggestedDevices).toEqual(devicesData);
    });

    it("should update device tags", async () => {
      const updateData = { uid: "a582b47a42d", tags: { tags: ["tag1", "tag2"] } };

      mockDevices.onPut("http://localhost:3000/api/devices/a582b47a42d/tags").reply(200);

      await expect(deviceStore.updateDeviceTags(updateData)).resolves.not.toThrow();
    });

    it("should handle fetch device list error", async () => {
      mockDevices.onGet("http://localhost:3000/api/devices?page=1&per_page=10&status=accepted").reply(500);

      await expect(deviceStore.fetchDeviceList({ page: 1, perPage: 10, status: "accepted" })).rejects.toThrow();

      expect(deviceStore.devices).toEqual([]);
      expect(deviceStore.deviceCount).toBe(0);
    });

    it("should handle fetch device error", async () => {
      mockDevices.onGet("http://localhost:3000/api/devices/resolve?uid=a582b47a42d").reply(404);

      await expect(deviceStore.fetchDevice({ uid: "a582b47a42d" })).rejects.toThrow();

      expect(deviceStore.device).toEqual({});
    });

    it("should handle fetch most used devices error", async () => {
      mockBilling.onGet("http://localhost:3000/api/billing/devices-most-used").reply(500);

      await expect(deviceStore.fetchMostUsedDevices()).rejects.toThrow();

      expect(deviceStore.suggestedDevices).toEqual([]);
    });
  });
});
