import { describe, expect, it, afterEach } from "vitest";
import MockAdapter from "axios-mock-adapter";
import { createPinia, setActivePinia } from "pinia";
import { devicesApi, billingApi } from "@/api/http";
import { IDevice } from "@/interfaces/IDevice";
import useDevicesStore from "@/store/modules/devices";

const deviceData = {
  uid: "a582b47a42d",
  name: "Device1",
  status: "accepted",
};

describe("Devices Pinia Store", () => {
  setActivePinia(createPinia());
  const mockDevicesApi = new MockAdapter(devicesApi.getAxios());
  const mockBillingApi = new MockAdapter(billingApi.getAxios());
  const devicesStore = useDevicesStore();

  afterEach(() => {
    mockDevicesApi.reset();
    mockBillingApi.reset();
  });

  it("should have initial state values", () => {
    expect(devicesStore.devices).toEqual([]);
    expect(devicesStore.device).toEqual({});
    expect(devicesStore.showDevices).toBe(false);
    expect(devicesStore.deviceCount).toBe(0);
    expect(devicesStore.showDeviceChooser).toBe(false);
    expect(devicesStore.suggestedDevices).toEqual([]);
    expect(devicesStore.selectedDevices).toEqual([]);
    expect(devicesStore.duplicatedDeviceName).toBe("");
  });

  it("should fetch device list successfully", async () => {
    const devicesData = [
      deviceData,
      { uid: "a582b47a42e", name: "Device2", status: "accepted" },
    ];

    mockDevicesApi.onGet("http://localhost:3000/api/devices?page=1&per_page=10&status=accepted").reply(200, devicesData, {
      "x-total-count": "2",
    });

    await devicesStore.fetchDeviceList({ page: 1, perPage: 10, status: "accepted" });

    expect(devicesStore.devices).toEqual(devicesData);
    expect(devicesStore.deviceCount).toBe(2);
    expect(devicesStore.showDevices).toBe(true);
  });

  it("should handle empty device list", async () => {
    mockDevicesApi.onGet("http://localhost:3000/api/devices?page=1&per_page=10&status=accepted").reply(200, [], {
      "x-total-count": "0",
    });

    await devicesStore.fetchDeviceList({ page: 1, perPage: 10, status: "accepted" });

    expect(devicesStore.devices).toEqual([]);
    expect(devicesStore.deviceCount).toBe(0);
  });

  it("should remove device", async () => {
    mockDevicesApi.onDelete("http://localhost:3000/api/devices/a582b47a42d").reply(200);
    await expect(devicesStore.removeDevice("a582b47a42d")).resolves.not.toThrow();
  });

  it("should rename device", async () => {
    const renameData = { uid: "a582b47a42d", name: { name: "UpdatedDevice1" } };
    devicesStore.device = deviceData as IDevice;
    mockDevicesApi.onPut("http://localhost:3000/api/devices/a582b47a42d").reply(200);
    await devicesStore.renameDevice(renameData);
    expect(devicesStore.device.name).toBe("UpdatedDevice1");
  });

  it("should fetch device by UID", async () => {
    mockDevicesApi.onGet("http://localhost:3000/api/devices/resolve?uid=a582b47a42d").reply(200, deviceData);
    await devicesStore.fetchDevice({ uid: "a582b47a42d" });
    expect(devicesStore.device).toEqual(deviceData);
  });

  it("should fetch device by hostname", async () => {
    mockDevicesApi.onGet("http://localhost:3000/api/devices/resolve?hostname=Device1").reply(200, deviceData);
    await devicesStore.fetchDevice({ hostname: "Device1" });
    expect(devicesStore.device).toEqual(deviceData);
  });

  it("should accept device", async () => {
    mockDevicesApi.onPatch("http://localhost:3000/api/devices/a582b47a42d/accept").reply(200);
    await expect(devicesStore.acceptDevice("a582b47a42d")).resolves.not.toThrow();
  });

  it("should reject device", async () => {
    mockDevicesApi.onPatch("http://localhost:3000/api/devices/a582b47a42d/reject").reply(200);
    await expect(devicesStore.rejectDevice("a582b47a42d")).resolves.not.toThrow();
  });

  it("should get first pending device", async () => {
    const pendingDeviceData = { ...deviceData, status: "pending" };
    mockDevicesApi.onGet("http://localhost:3000/api/devices?page=1&per_page=1&status=pending").reply(200, [pendingDeviceData]);
    const result = await devicesStore.getFirstPendingDevice();
    expect(result).toEqual(pendingDeviceData);
  });

  it("should fetch most used devices", async () => {
    const devicesData = [
      deviceData,
      { uid: "a582b47a42e", name: "Device2", status: "accepted" },
    ];

    mockBillingApi.onGet("http://localhost:3000/api/billing/devices-most-used").reply(200, devicesData);
    await devicesStore.fetchMostUsedDevices();
    expect(devicesStore.suggestedDevices).toEqual(devicesData);
  });

  it("should update device tags", async () => {
    const updateData = { uid: "a582b47a42d", tags: { tags: ["tag1", "tag2"] } };
    mockDevicesApi.onPut("http://localhost:3000/api/devices/a582b47a42d/tags").reply(200);
    await expect(devicesStore.updateDeviceTags(updateData)).resolves.not.toThrow();
  });

  it("should handle fetch device list error", async () => {
    mockDevicesApi.onGet("http://localhost:3000/api/devices?page=1&per_page=10&status=accepted").reply(500);
    await expect(devicesStore.fetchDeviceList({ page: 1, perPage: 10, status: "accepted" })).rejects.toThrow();
    expect(devicesStore.devices).toEqual([]);
    expect(devicesStore.deviceCount).toBe(0);
  });

  it("should handle fetch device error", async () => {
    mockDevicesApi.onGet("http://localhost:3000/api/devices/resolve?uid=a582b47a42d").reply(404);
    await expect(devicesStore.fetchDevice({ uid: "a582b47a42d" })).rejects.toThrow();
    expect(devicesStore.device).toEqual({});
  });

  it("should handle fetch most used devices error", async () => {
    mockBillingApi.onGet("http://localhost:3000/api/billing/devices-most-used").reply(500);
    await expect(devicesStore.fetchMostUsedDevices()).rejects.toThrow();
    expect(devicesStore.suggestedDevices).toEqual([]);
  });
});
