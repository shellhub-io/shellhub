<template>
  <div class="pa-4" v-if="firstPendingDevice">
    <p class="ml-4 pt-4 text-subtitle-2" data-test="welcome-third-screen-name">
      A device connection has been detected.
    </p>
    <p class="ml-4 pt-4 text-subtitle-2" data-test="welcome-third-screen-text">
      Please confirm that this device is yours to enroll into your account.
      After confirmation, you will go to the last step of introducing
      <strong>ShellHub</strong>.
    </p>

    <v-row no-gutters class="mt-4">
      <v-col>
        <v-card class="pa-2 bg-v-theme-surface" tile :elevation="0" align="center" data-test="welcome-third-screen-hostname">
          <strong>Hostname</strong>
        </v-card>
        <v-divider />
        <v-card
          class="pa-2 bg-v-theme-surface"
          tile
          :elevation="0"
          align="center"
          data-test="device-field"
        >
          {{ firstPendingDevice.name }}
        </v-card>
      </v-col>
      <v-col>
        <v-card class="pa-2 bg-v-theme-surface" tile :elevation="0" align="center" data-test="welcome-third-screen-os">
          <strong>Operation System</strong>
        </v-card>
        <v-divider />
        <v-card
          class="pa-2 bg-v-theme-surface"
          tile
          :elevation="0"
          align="center"
          data-test="device-pretty-name-field"
        >
          <div v-if="firstPendingDevice.info">
            <DeviceIcon :icon="firstPendingDevice.info.id" />
            {{ firstPendingDevice.info.pretty_name }}
          </div>
        </v-card>
      </v-col>
    </v-row>
  </div>
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
