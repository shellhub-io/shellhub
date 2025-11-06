<template>
  <div class="d-flex flex-column justify-space-between align-center flex-sm-row mb-2">
    <h1>Devices</h1>
    <v-spacer />
    <v-text-field
      v-model.trim="filter"
      class="w-50"
      label="Search by hostname"
      color="primary"
      single-line
      hide-details
      append-inner-icon="mdi-magnify"
      density="compact"
      @keyup="searchDevices"
    />
    <v-spacer />
  </div>
  <DeviceList />
</template>

<script setup lang="ts">
import { ref } from "vue";
import useDevicesStore from "@admin/store/modules/devices";
import useSnackbar from "@/helpers/snackbar";
import DeviceList from "../components/Device/DeviceList.vue";

const snackbar = useSnackbar();
const devicesStore = useDevicesStore();

const filter = ref("");

const searchDevices = async () => {
  let encodedFilter = "";

  if (filter.value) {
    const filterToEncodeBase64 = [
      { type: "property", params: { name: "name", operator: "contains", value: filter.value } },
    ];
    encodedFilter = btoa(JSON.stringify(filterToEncodeBase64));
  }

  devicesStore.setFilter(encodedFilter);

  try {
    await devicesStore.fetchDeviceList({ filter: encodedFilter, page: 1 });
  } catch {
    snackbar.showError("Failed to fetch the devices.");
  }
};

defineExpose({ filter });
</script>
