<template>
  <div>
    <DeviceTable
      variant="device"
      header="primary"
      status="accepted"
      :storeMethods="storeMethods"
      data-test="device-table"
      :committable="true"
    />
  </div>
</template>

<script setup lang="ts">
import { ref } from "vue";
import DeviceTable from "../Tables/DeviceTable.vue";
import { FetchDevicesParams, IDeviceMethods } from "../../interfaces/IDevice";
import { useStore } from "@/store";

const store = useStore();

const filter = ref(btoa(JSON.stringify([
  {
    type: "property",
    params: {
      name: "info.platform",
      operator: "eq",
      value: "native",
    },
  },
  {
    type: "property",
    params: {
      name: "info.platform",
      operator: "eq",
      value: "docker",
    },
  },
  {
    type: "operator",
    params: {
      name: "or",
    },
  },
])));

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

const getFilter = () => filter.value;
const getDevicesList = () => store.getters["devices/list"];
const getSortStatusField = () => store.getters["devices/getSortStatusField"];
const getSortStatusString = () => store.getters["devices/getSortStatusString"];
const getNumberDevices = () => store.getters["devices/getNumberDevices"];

const storeMethods: IDeviceMethods = {
  fetchDevices,
  getFilter,
  getDevicesList,
  getSortStatusField,
  getSortStatusString,
  getNumberDevices,
};

</script>
