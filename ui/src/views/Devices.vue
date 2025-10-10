<template>
  <div
    class="d-flex flex-column justify-space-between align-center flex-sm-row"
    data-test="device-title"
  >
    <h1>Devices</h1>
    <v-col md="6">
      <v-text-field
        v-if="showDevices"
        label="Search by hostname"
        variant="outlined"
        color="primary"
        single-line
        hide-details
        v-model.trim="filter"
        @update:model-value="searchDevices"
        prepend-inner-icon="mdi-magnify"
        density="compact"
        data-test="search-text"
      />
    </v-col>

    <div class="d-flex" data-test="device-header-component-group">
      <TagSelector variant="device" v-if="isDeviceList" />
      <DeviceAdd />
    </div>
  </div>
  <div class="mt-2" v-if="showDevices" data-test="device-table-component">
    <Device />
  </div>

  <NoItemsMessage
    v-else
    class="mt-2"
    item="Devices"
    icon="mdi-developer-board"
    data-test="no-items-message-component"
  >
    <template #content>
      <p>In order to register a device on ShellHub, you need to install ShellHub agent onto it.</p>
      <p>The easiest way to install ShellHub agent is with our automatic one-line installation script,
        which works with all Linux distributions that have Docker installed and properly set up.
        <a
          rel="noopener noreferrer"
          target="_blank"
          href="https://docs.shellhub.io/user-guides/devices/adding"
        >See More</a>.
      </p>
    </template>
    <template #action>
      <DeviceAdd />
    </template>
  </NoItemsMessage>
</template>

<script setup lang="ts">
import { computed, ref } from "vue";
import { useRoute } from "vue-router";
import Device from "../components/Devices/Device.vue";
import DeviceAdd from "../components/Devices/DeviceAdd.vue";
import TagSelector from "../components/Tags/TagSelector.vue";
import NoItemsMessage from "../components/NoItemsMessage.vue";
import useSnackbar from "@/helpers/snackbar";
import useDevicesStore from "@/store/modules/devices";
import { DeviceStatus } from "@/interfaces/IDevice";

const devicesStore = useDevicesStore();
const route = useRoute();
const snackbar = useSnackbar();
const filter = ref("");
const showDevices = computed(() => devicesStore.showDevices);
const isDeviceList = computed(() => route.name === "DeviceList");

const statusMap: Record<string, DeviceStatus> = {
  "/devices": "accepted",
  "/devices/pending": "pending",
  "/devices/rejected": "rejected",
};

const searchDevices = async () => {
  const filterToEncodeBase64 = [{
    type: "property",
    params: { name: "name", operator: "contains", value: filter.value },
  }];

  const encodedFilter = filter.value ? btoa(JSON.stringify(filterToEncodeBase64)) : undefined;

  const status = statusMap[route.path] || "accepted";

  try {
    await devicesStore.fetchDeviceList({ filter: encodedFilter, status });
  } catch {
    snackbar.showError("Failed to load devices.");
  }
};
</script>
