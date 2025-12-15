<template>
  <PageHeader
    icon="mdi-developer-board"
    title="Devices"
    overline="Device Management"
    description="Manage and monitor all devices connected to this namespace. Install the ShellHub agent on your devices to register them."
    icon-color="primary"
    data-test="device-title"
  >
    <template #actions>
      <DeviceAdd />
    </template>
  </PageHeader>
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
import { computed } from "vue";
import Device from "../components/Devices/Device.vue";
import DeviceAdd from "../components/Devices/DeviceAdd.vue";
import NoItemsMessage from "../components/NoItemsMessage.vue";
import PageHeader from "../components/PageHeader.vue";
import useDevicesStore from "@/store/modules/devices";

const devicesStore = useDevicesStore();
const showDevices = computed(() => devicesStore.showDevices);
</script>
