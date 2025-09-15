import { defineStore } from "pinia";
import { ref } from "vue";
import { IAdminDevice } from "@admin/interfaces/IDevice";
import * as devicesApi from "../api/devices";

const useDevicesStore = defineStore("devices", () => {
  const devices = ref<Array<IAdminDevice>>([]);
  const deviceCount = ref(0);

  const fetchDeviceList = async (data?: {
    page?: number;
    perPage?: number;
    filter?: string;
    sortField?: string;
    sortOrder?: "asc" | "desc";
  }) => {
    const res = await devicesApi.getDevices(
      data?.page || 1,
      data?.perPage || 10,
      data?.filter,
      data?.sortField,
      data?.sortOrder,
    );
    devices.value = res.data as IAdminDevice[];
    deviceCount.value = parseInt(res.headers["x-total-count"], 10);
  };

  const fetchDeviceById = async (uid: string) => {
    const res = await devicesApi.getDevice(uid);
    return res.data as IAdminDevice;
  };

  return {
    devices,
    deviceCount,
    fetchDeviceList,
    fetchDeviceById,
  };
});

export default useDevicesStore;
