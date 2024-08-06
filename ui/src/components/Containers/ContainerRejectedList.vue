<template>
  <div>
    <DeviceTable
      variant="container"
      header="secondary"
      status="rejected"
      :storeMethods="storeMethods"
      data-test="container-table"
    />
  </div>
</template>

<script setup lang="ts">
import DeviceTable from "../Tables/DeviceTable.vue";
import { FetchContainerParams, IContainerMethods } from "../../interfaces/IContainer";
import { useStore } from "@/store";

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

const getFilter = () => store.getters["container/getFilter"];
const getList = () => store.getters["container/list"];
const getSortStatusField = () => store.getters["container/getSortStatusField"];
const getSortStatusString = () => store.getters["container/getSortStatusString"];
const getNumber = () => store.getters["container/getNumberContainers"];

const storeMethods: IContainerMethods = {
  fetchDevices,
  getFilter,
  getList,
  getSortStatusField,
  getSortStatusString,
  getNumber,
};
</script>
