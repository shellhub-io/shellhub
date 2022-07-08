<template>
  <div class="pa-4">
    <p class="ml-4 pt-4 text-subtitle-2">
      A device connection has been detected.
    </p>
    <p class="ml-4 pt-4 text-subtitle-2">
      Please confirm that this device is yours to enroll into your account.
      After confirmation, you will go to the last step of introducing
      <strong>ShellHub</strong>.
    </p>

    <v-row no-gutters class="mt-4">
      <v-col>
        <v-card class="pa-2" tile :elevation="0">
          <strong>Hostname</strong>
        </v-card>
        <v-card
          class="pa-2"
          tile
          :elevation="0"
          data-test="deviceName-field"
        >
          {{ getPendingDevice.name }}
        </v-card>
      </v-col>
      <v-col>
        <v-card class="pa-2" tile :elevation="0">
          <strong>Operation System</strong>
        </v-card>
        <v-card
          class="pa-2"
          tile
          :elevation="0"
          data-test="devicePrettyName-field"
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

<script lang="ts">
import { useStore } from "../../store";
import { defineComponent, computed, onMounted } from "vue";
import { INotificationsError } from "../../interfaces/INotifications";
import DeviceIcon from "../Devices/DeviceIcon.vue";

export default defineComponent({
  setup() {
    const store = useStore();
    const getPendingDevice = computed(
      () => store.getters["devices/getFirstPending"]
    );
    onMounted(() => {
      try {
        store.dispatch("devices/setFirstPending");
      } catch {
        store.dispatch(
          "snackbar/showSnackbarErrorLoading",
          INotificationsError.devicePending
        );
      }
    });
    return {
      getPendingDevice,
    };
  },
  components: { DeviceIcon },
});
</script>
