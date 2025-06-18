import { useStore } from "@/store";
import { FetchContainerParams, IContainerMethods, SortContainersParams } from "@/interfaces/IContainer";
import { FetchDevicesParams, IDeviceMethods, SortDevicesParams } from "@/interfaces/IDevice";

export function getContainerStoreMethods(): IContainerMethods {
  const store = useStore();

  const fetchDevices = async ({ perPage, page, filter, status, sortStatusField, sortStatusString }: FetchContainerParams) => {
    await store.dispatch("container/fetch", {
      perPage,
      page,
      filter,
      status,
      sortStatusField,
      sortStatusString,
    });
  };

  const setSort = ({ sortStatusField, sortStatusString }: SortContainersParams) => {
    store.dispatch("container/setSortStatus", {
      sortStatusField,
      sortStatusString,
    });
  };

  const getFilter = () => store.getters["container/getFilter"];
  const getList = () => store.getters["container/list"];
  const getSortStatusField = () => store.getters["container/getSortStatusField"];
  const getSortStatusString = () => store.getters["container/getSortStatusString"];
  const getNumber = () => store.getters["container/getNumberContainers"];

  return {
    fetchDevices,
    setSort,
    getFilter,
    getList,
    getSortStatusField,
    getSortStatusString,
    getNumber,
  };
}

export function getDeviceStoreMethods(): IDeviceMethods {
  const store = useStore();

  const fetchDevices = async ({ perPage, page, filter, status, sortStatusField, sortStatusString }: FetchDevicesParams) => {
    await store.dispatch("devices/fetch", {
      perPage,
      page,
      filter,
      status,
      sortStatusField,
      sortStatusString,
    });
  };

  const setSort = ({ sortStatusField, sortStatusString }: SortDevicesParams) => {
    store.dispatch("devices/setSortStatus", {
      sortStatusField,
      sortStatusString,
    });
  };

  const getFilter = () => store.getters["devices/getFilter"];
  const getList = () => store.getters["devices/list"];
  const getSortStatusField = () => store.getters["devices/getSortStatusField"];
  const getSortStatusString = () => store.getters["devices/getSortStatusString"];
  const getNumber = () => store.getters["devices/getNumberDevices"];

  return {
    fetchDevices,
    setSort,
    getFilter,
    getList,
    getSortStatusField,
    getSortStatusString,
    getNumber,
  };
}
