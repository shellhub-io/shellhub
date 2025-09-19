<template>
  <div class="d-flex flex-column justify-space-between align-center flex-sm-row">
    <h1>Devices</h1>
    <v-spacer />
    <div class="w-50">
      <v-text-field
        label="Search by hostname"
        variant="underlined"
        color="primary"
        single-line
        hide-details
        v-model.trim="filter"
        v-on:keyup="searchDevices"
        append-inner-icon="mdi-magnify"
        density="comfortable"
      />
    </div>
    <v-spacer />
  </div>
  <v-card class="mt-2">
    <DeviceList />
  </v-card>
</template>

<script setup lang="ts">
import { ref } from "vue";
import useDevicesStore from "@admin/store/modules/devices";
import useSnackbar from "@/helpers/snackbar";
import DeviceList from "../components/Device/DeviceList.vue";

const snackbar = useSnackbar();
const devicesStore = useDevicesStore();

const filter = ref("");

const searchDevices = () => {
  let encodedFilter = "";

  if (filter.value) {
    const filterToEncodeBase64 = [
      { type: "property", params: { name: "name", operator: "contains", value: filter.value } },
    ];
    encodedFilter = btoa(JSON.stringify(filterToEncodeBase64));
  }

  devicesStore.setFilter(encodedFilter);

  try {
    devicesStore.fetchDeviceList({ filter: encodedFilter, page: 1 });
  } catch {
    snackbar.showError("Failed to fetch the devices.");
  }
};

defineExpose({ filter });
</script>
