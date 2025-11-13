import { defineStore } from "pinia";
import { ref } from "vue";
import { IAdminDevice } from "@admin/interfaces/IDevice";
import * as devicesApi from "../api/devices";

const useDevicesStore = defineStore("adminDevices", () => {
  const devices = ref<IAdminDevice[]>([]);
  const deviceCount = ref(0);

  const currentFilter = ref<string>("");
  const currentSortField = ref<string | undefined>(undefined);
  const currentSortOrder = ref<"asc" | "desc" | undefined>(undefined);

  const setFilter = (filter: string) => { currentFilter.value = filter || ""; };
  const setSort = (field?: string, order?: "asc" | "desc") => {
    currentSortField.value = field;
    currentSortOrder.value = order;
  };

  const fetchDeviceList = async (data?: {
    page?: number;
    perPage?: number;
    filter?: string;
    sortField?: string;
    sortOrder?: "asc" | "desc";
  }) => {
    const page = data?.page || 1;
    const perPage = data?.perPage || 10;
    const filter = data?.filter ?? currentFilter.value ?? "";
    const sortField = data?.sortField ?? currentSortField.value;
    const sortOrder = data?.sortOrder ?? currentSortOrder.value;

    const res = await devicesApi.getDevices(page, perPage, filter, sortField, sortOrder);
    devices.value = res.data as unknown as IAdminDevice[];
    deviceCount.value = parseInt(res.headers["x-total-count"] as string, 10);
  };

  const fetchDeviceById = async (uid: string) => {
    const res = await devicesApi.getDevice(uid);
    return res.data as unknown as IAdminDevice;
  };

  return {
    devices,
    deviceCount,
    currentFilter,
    currentSortField,
    currentSortOrder,
    setFilter,
    setSort,

    fetchDeviceList,
    fetchDeviceById,
  };
});

export default useDevicesStore;
