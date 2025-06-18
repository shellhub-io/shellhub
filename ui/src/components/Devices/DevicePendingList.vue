<template>
  <div>
    <DeviceTable
      variant="device"
      header="secondary"
      status="pending"
      :storeMethods="storeMethods"
      data-test="device-table"
    />
  </div>
</template>

<script setup lang="ts">
import DeviceTable from "../Tables/DeviceTable.vue";
import { FetchDevicesParams, IDeviceMethods, SortDevicesParams } from "@/interfaces/IDevice";
import { useStore } from "@/store";

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

const storeMethods: IDeviceMethods = {
  fetchDevices,
  setSort,
  getFilter,
  getList,
  getSortStatusField,
  getSortStatusString,
  getNumber,
};

</script>
