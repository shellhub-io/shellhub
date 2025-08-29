import { ref } from "vue";
import { defineStore } from "pinia";
import * as devicesApi from "../api/devices";
import * as billingApi from "../api/billing";
import { FetchDevicesParams, IDevice, IDeviceRename, IUpdateDeviceTags } from "@/interfaces/IDevice";

const useDevicesStore = defineStore("devices", () => {
  const devices = ref<Array<IDevice>>([]);
  const device = ref<IDevice>({} as IDevice);
  const showDevices = ref<boolean>(false);
  const deviceCount = ref<number>(0);
  const duplicatedDeviceName = ref<string>("");

  const onlineDevices = ref<Array<IDevice>>([]);

  const showDeviceChooser = ref<boolean>(false);
  const suggestedDevices = ref<Array<IDevice>>([]);
  const selectedDevices = ref<Array<IDevice>>([]);

  const fetchDeviceList = async (data?: FetchDevicesParams) => {
    try {
      const res = await devicesApi.fetchDevices(
        data?.page || 1,
        data?.perPage || 10,
        data?.status || "accepted",
        data?.filter,
        data?.sortField,
        data?.sortOrder,
      );
      devices.value = res.data as IDevice[];
      deviceCount.value = parseInt(res.headers["x-total-count"], 10);
      if (deviceCount.value) showDevices.value = true;
    } catch (error) {
      devices.value = [];
      deviceCount.value = 0;
      throw error;
    }
  };

  const fetchOnlineDevices = async (filter?: string) => {
    try {
      const res = await devicesApi.fetchDevices(
        1,
        10,
        "accepted",
        filter,
      );
      onlineDevices.value = res.data.filter((device) => device.online) as IDevice[];
    } catch (error) {
      onlineDevices.value = [];
      throw error;
    }
  };

  const removeDevice = async (uid: string) => {
    await devicesApi.removeDevice(uid);
  };

  const renameDevice = async (data: IDeviceRename) => {
    await devicesApi.renameDevice(data);
    device.value.name = data.name.name as string;
  };

  const fetchDevice = async (identifiers: { hostname?: string, uid?: string }) => {
    try {
      const { hostname, uid } = identifiers;
      const res = await devicesApi.resolveDevice(hostname, uid);
      device.value = res.data as IDevice;
    } catch (error) {
      device.value = {} as IDevice;
      throw error;
    }
  };

  const acceptDevice = async (uid: string) => {
    await devicesApi.acceptDevice(uid);
  };

  const rejectDevice = async (uid: string) => {
    await devicesApi.rejectDevice(uid);
  };

  const getFirstPendingDevice = async () => {
    const res = await devicesApi.fetchDevices(
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
    await devicesApi.updateDeviceTags(data);
  };

  return {
    devices,
    device,
    showDevices,
    deviceCount,
    onlineDevices,
    showDeviceChooser,
    suggestedDevices,
    selectedDevices,
    duplicatedDeviceName,

    fetchDeviceList,
    fetchOnlineDevices,
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
