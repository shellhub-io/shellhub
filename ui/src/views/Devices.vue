<template>
  <div
    class="d-flex flex-column justify-space-between align-center flex-sm-row"
    data-test="device-title"
  >
    <h1>Devices</h1>
    <v-col md="6">
      <v-text-field
        v-if="showDevices"
        v-model.trim="filter"
        label="Search by hostname"
        variant="outlined"
        color="primary"
        single-line
        hide-details
        prepend-inner-icon="mdi-magnify"
        density="compact"
        data-test="search-text"
        @update:model-value="updateDeviceListFilter"
      />
    </v-col>

    <div
      class="d-flex"
      data-test="device-header-component-group"
    >
      <TagSelector
        v-if="isDeviceList"
        variant="device"
      />
      <DeviceAdd />
    </div>
  </div>
  <div
    v-if="showDevices"
    class="mt-2"
    data-test="device-table-component"
  >
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
      <p>
        The easiest way to install ShellHub agent is with our automatic one-line installation script,
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
import useDevicesStore from "@/store/modules/devices";

const devicesStore = useDevicesStore();
const route = useRoute();
const filter = ref("");
const showDevices = computed(() => devicesStore.showDevices);
const isDeviceList = computed(() => route.name === "DeviceList");

const updateDeviceListFilter = () => {
  const base64DeviceFilter = [{
    type: "property",
    params: { name: "name", operator: "contains", value: filter.value },
  }];

  const encodedFilter = filter.value ? btoa(JSON.stringify(base64DeviceFilter)) : undefined;

  devicesStore.deviceListFilter = encodedFilter;
};
</script>
