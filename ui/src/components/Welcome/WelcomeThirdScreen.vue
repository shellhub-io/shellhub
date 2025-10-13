<template>
  <div class="pa-6" v-if="firstPendingDevice">
    <div class="text-center mb-6">
      <v-avatar
        size="64"
        color="success"
        class="mb-4"
      >
        <v-icon size="32" color="white" icon="mdi-check-circle" />
      </v-avatar>
      <h2 class="text-h4 mb-2">Device Detected!</h2>
      <p class="text-subtitle-1 text-medium-emphasis">Confirm this device to add it to your account</p>
    </div>

    <v-alert
      color="success"
      variant="tonal"
      class="mb-4"
      icon="mdi-lan-connect"
      title="Connection Successful"
      data-test="welcome-third-screen-name"
    >
      <p class="mb-0" data-test="welcome-third-screen-text">
        A device has been detected and is ready for enrollment.
        Please verify the device information below and confirm it belongs to you.
      </p>
    </v-alert>

    <v-expansion-panels>
      <v-expansion-panel
        class="border bg-background rounded"
      >
        <v-expansion-panel-title class="text-white bg-v-theme-surface border-b d-flex align-center">
          <DeviceIcon
            v-if="firstPendingDevice.info"
            :icon="firstPendingDevice.info.id"
            class="mr-3"
            size="32"
          />
          <div>
            <div class="text-h6" data-test="device-field">{{ firstPendingDevice.name }}</div>
            <div v-if="firstPendingDevice.info" class="text-body-2 text-white" data-test="device-pretty-name-field">
              {{ firstPendingDevice.info.pretty_name }}
            </div>
          </div>
        </v-expansion-panel-title>
        <v-expansion-panel-text class="pt-2">
          <div class="item-title mt-0">Device UID:</div>
          <p>{{ firstPendingDevice.uid }}</p>
          <div v-if="firstPendingDevice.identity" class="mb-3">
            <div class="item-title">MAC Address:</div>
            <code>{{ firstPendingDevice.identity.mac }}</code>
          </div>
          <div v-if="firstPendingDevice.info" class="mb-3">
            <div class="item-title">Agent Version:</div>
            <p>{{ firstPendingDevice.info.version }}</p>
          </div>
          <v-divider class="my-4" />
          <div class="text-caption text-medium-emphasis">
            <v-icon size="16" class="mr-1" icon="mdi-information" />
            This device will be added to your account after confirmation in the next step.
          </div>
        </v-expansion-panel-text>
      </v-expansion-panel>
    </v-expansion-panels>

  </div>
  <v-card
    v-else
    variant="tonal"
    color="warning"
    class="pa-4 text-center"
  >
    <v-icon size="48" color="warning" class="mb-3">mdi-clock-alert</v-icon>
    <h3 class="text-h6 mb-2" data-test="no-device-heading">No Device Detected Yet</h3>
    <p class="text-body-2 mb-0" data-test="no-device-text">
      Please run the installation command from the previous step and wait for your device to connect.
    </p>
  </v-card>
</template>

<script setup lang="ts">
import { onBeforeMount } from "vue";
import DeviceIcon from "../Devices/DeviceIcon.vue";
import useSnackbar from "@/helpers/snackbar";
import useDevicesStore from "@/store/modules/devices";
import { IDevice } from "@/interfaces/IDevice";

const devicesStore = useDevicesStore();
const snackbar = useSnackbar();
const firstPendingDevice = defineModel<IDevice>("firstPendingDevice");

onBeforeMount(async () => {
  try {
    firstPendingDevice.value = await devicesStore.getFirstPendingDevice();
  } catch {
    snackbar.showError("Failed to get pending device.");
  }
});
</script>

<style lang="scss" scoped>
.item-title {
  margin-top: 0.75rem;
  // Vuetify's text-overline styles
  font-size: 0.75rem;
  font-weight: 500;
  text-transform: uppercase;
  letter-spacing: 0.1666666667em;
  line-height: 2.667;
}
</style>
