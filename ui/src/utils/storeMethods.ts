import { FetchContainerParams, IContainerMethods } from "@/interfaces/IContainer";
import { FetchDevicesParams, IDeviceMethods } from "@/interfaces/IDevice";
import useContainersStore from "@/store/modules/containers";
import useDevicesStore from "@/store/modules/devices";

export function getContainerStoreMethods(): IContainerMethods {
  const containersStore = useContainersStore();

  const fetchDevices = async ({ perPage, page, filter, status, sortField, sortOrder }: FetchContainerParams) => {
    await containersStore.fetchContainerList({
      perPage,
      page,
      filter,
      status,
      sortField,
      sortOrder,
    });
  };

  const getList = () => containersStore.containers;
  const getCount = () => containersStore.containerCount;
  const getFilter = () => containersStore.containerListFilter;

  return {
    fetchDevices,
    getList,
    getCount,
    getFilter,
  };
}

export function getDeviceStoreMethods(): IDeviceMethods {
  const devicesStore = useDevicesStore();

  const fetchDevices = async ({ perPage, page, filter, status, sortField, sortOrder }: FetchDevicesParams) => {
    await devicesStore.fetchDeviceList({
      perPage,
      page,
      filter,
      status,
      sortField,
      sortOrder,
    });
  };

  const getList = () => devicesStore.devices;
  const getCount = () => devicesStore.deviceCount;
  const getFilter = () => devicesStore.deviceListFilter;

  return {
    fetchDevices,
    getList,
    getCount,
    getFilter,
  };
}
