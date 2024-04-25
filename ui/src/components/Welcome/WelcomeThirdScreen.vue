<template>
  <div class="pa-4">
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
          {{ getPendingDevice.name }}
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
          <div v-if="getPendingDevice.info">
            <DeviceIcon :icon="getPendingDevice.info.id" />
            {{ getPendingDevice.info.pretty_name }}
          </div>
        </v-card>
      </v-col>
    </v-row>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted } from "vue";
import { useStore } from "../../store";
import { INotificationsError } from "../../interfaces/INotifications";
import DeviceIcon from "../Devices/DeviceIcon.vue";

const store = useStore();
const getPendingDevice = computed(
  () => store.getters["devices/getFirstPending"],
);
onMounted(() => {
  try {
    store.dispatch("devices/setFirstPending");
  } catch {
    store.dispatch(
      "snackbar/showSnackbarErrorLoading",
      INotificationsError.devicePending,
    );
  }
});
</script>
