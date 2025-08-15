import { ref } from "vue";
import { defineStore } from "pinia";
import * as deviceApi from "../api/devices";
import * as billingApi from "../api/billing";
import { FetchDevicesParams, IDevice, IDeviceRename, IUpdateDeviceTags } from "@/interfaces/IDevice";

export const useDevicesStore = defineStore("devices", () => {
  const devices = ref<Array<IDevice>>([]);
  const device = ref<IDevice>({} as IDevice);
  const showDevices = ref<boolean>(false);
  const deviceCount = ref<number>(0);
  const duplicatedDeviceName = ref<string>("");

  const showDeviceChooser = ref<boolean>(false);
  const suggestedDevices = ref<Array<IDevice>>([]);
  const selectedDevices = ref<Array<IDevice>>([]);

  const fetchDeviceList = async (data?: FetchDevicesParams) => {
    try {
      const res = await deviceApi.fetchDevices(
        data?.page || 1,
        data?.perPage || 10,
        data?.status || "accepted",
        data?.filter,
        data?.sortField,
        data?.sortOrder,
      );
      devices.value = res.data as IDevice[];
      deviceCount.value = parseInt(res.headers["x-total-count"], 10);
      showDevices.value = true;
    } catch (error) {
      devices.value = [];
      deviceCount.value = 0;
      throw error;
    }
  };

  const removeDevice = async (uid: string) => {
    await deviceApi.removeDevice(uid);
  };

  const renameDevice = async (data: IDeviceRename) => {
    await deviceApi.renameDevice(data);
    device.value.name = data.name.name as string;
  };

  const fetchDevice = async (identifiers: { hostname?: string, uid?: string }) => {
    try {
      const { hostname, uid } = identifiers;
      const res = await deviceApi.resolveDevice(hostname, uid);
      device.value = res.data as IDevice;
    } catch (error) {
      device.value = {} as IDevice;
      throw error;
    }
  };

  const acceptDevice = async (uid: string) => {
    await deviceApi.acceptDevice(uid);
  };

  const rejectDevice = async (uid: string) => {
    await deviceApi.rejectDevice(uid);
  };

  const getFirstPendingDevice = async () => {
    const res = await deviceApi.fetchDevices(
      1,
      1,
      "pending",
    );
    return res.data[0] as IDevice;
  };

  const sendDeviceChoices = async (devices: Array<IDevice>) => {
    const uids = devices.map((device) => device.uid);
    await billingApi.postDevicesChooser({ choices: uids });
  };

  const fetchMostUsedDevices = async () => {
    try {
      const res = await billingApi.getDevicesMostUsed();
      suggestedDevices.value = res.data as IDevice[];
    } catch (error) {
      suggestedDevices.value = [];
      throw error;
    }
  };

  const updateDeviceTags = async (data: IUpdateDeviceTags) => {
    await deviceApi.updateDeviceTags(data);
  };

  return {
    devices,
    device,
    showDevices,
    deviceCount,
    showDeviceChooser,
    suggestedDevices,
    selectedDevices,
    duplicatedDeviceName,

    fetchDeviceList,
    removeDevice,
    renameDevice,
    fetchDevice,
    acceptDevice,
    rejectDevice,
    getFirstPendingDevice,
    sendDeviceChoices,
    fetchMostUsedDevices,
    updateDeviceTags,
  };
});

export default useDevicesStore;
