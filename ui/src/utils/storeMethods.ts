import { ref } from "vue";
import { useStore } from "@/store";
import { FetchContainerParams, IContainerMethods, SortContainersParams } from "@/interfaces/IContainer";
import { FetchDevicesParams, IDeviceMethods, SortDevicesParams } from "@/interfaces/IDevice";
import useContainersStore from "@/store/modules/containers";

export function getContainerStoreMethods(): IContainerMethods {
  const containersStore = useContainersStore();

  const currentFilter = ref("");
  const sortStatusField = ref<string>("");
  const sortStatusString = ref<"asc" | "desc" | undefined>("asc");

  const fetchDevices = async ({ perPage, page, filter, status, sortField, sortOrder }: FetchContainerParams) => {
    currentFilter.value = filter || "";
    sortStatusField.value = sortField || "";
    sortStatusString.value = sortOrder;

    await containersStore.fetchContainerList({
      perPage,
      page,
      filter,
      status,
      sortField,
      sortOrder,
    });
  };

  const setSort = ({ sortField, sortOrder }: SortContainersParams) => {
    sortStatusField.value = sortField || "";
    sortStatusString.value = sortOrder;
  };

  const getFilter = () => currentFilter.value;
  const getList = () => containersStore.containers;
  const getSortStatusField = () => sortStatusField.value;
  const getSortStatusString = () => sortStatusString.value;
  const getNumber = () => containersStore.containerCount;

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

  const fetchDevices = async ({ perPage, page, filter, status, sortField, sortOrder }: FetchDevicesParams) => {
    await store.dispatch("devices/fetch", {
      perPage,
      page,
      filter,
      status,
      sortField,
      sortOrder,
    });
  };

  const setSort = ({ sortField, sortOrder }: SortDevicesParams) => {
    store.dispatch("devices/setSortStatus", {
      sortField,
      sortOrder,
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
