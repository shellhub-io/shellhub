<template>
  <PageHeader
    icon="mdi-developer-board"
    title="Devices"
    overline="Fleet Oversight"
    description="Audit every registered device across all namespaces and quickly find specific hosts."
    icon-color="primary"
  >
    <v-text-field
      v-model.trim="filter"
      class="w-100 w-md-50"
      label="Search by hostname"
      color="primary"
      single-line
      hide-details
      append-inner-icon="mdi-magnify"
      density="compact"
      @keyup="searchDevices"
    />
  </PageHeader>
  <DeviceList />
</template>

<script setup lang="ts">
import { ref } from "vue";
import PageHeader from "@/components/PageHeader.vue";
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
