<template>
  <WindowDialog
    v-model="showDialog"
    title="Welcome to ShellHub!"
    :description="`Step ${el} of 4`"
    icon="mdi-door-open"
    icon-color="primary"
    @close="close"
  >
    <v-window v-model="el" class="overflow-y-auto" data-test="welcome-window">
      <v-window-item :value="1" data-test="welcome-first-screen">
        <WelcomeFirstScreen />
      </v-window-item>
      <v-window-item :value="2" data-test="welcome-second-screen">
        <WelcomeSecondScreen />
      </v-window-item>
      <v-window-item :value="3" data-test="welcome-third-screen">
        <WelcomeThirdScreen v-if="enable" v-model:first-pending-device="firstPendingDevice" />
      </v-window-item>
      <v-window-item :value="4" data-test="welcome-fourth-screen">
        <WelcomeFourthScreen />
      </v-window-item>
    </v-window>

    <template #footer>
      <div class="d-flex align-center w-100">
        <p v-if="el === 2" class="text-caption text-truncate" data-test="second-screen-helper-link">
          Check our
          <a
            href="https://docs.shellhub.io/user-guides/devices/adding"
            target="_blank"
            rel="noopener noreferrer"
            class="text-primary font-weight-medium"
          >
            documentation
            <v-icon size="12" icon="mdi-open-in-new" />
          </a>
          for alternative installation methods.
        </p>
        <v-spacer />
        <v-btn
          v-if="el === 1 || el === 2 || el === 3"
          @click="close"
          data-test="cancel-btn"
          class="mr-2"
        >
          Close
        </v-btn>
        <v-btn
          color="primary"
          @click="handleConfirm"
          :disabled="el === 2 && !enable"
          data-test="confirm-btn"
        >
          {{ el === 1 ? 'Next'
            : el === 2 ? 'Next'
              : el === 3 ? 'Accept'
                : el === 4 ? 'Finish'
                  : '' }}
        </v-btn>
      </div>
    </template>
  </WindowDialog>
</template>

<script setup lang="ts">
import { ref } from "vue";
import WelcomeFirstScreen from "./WelcomeFirstScreen.vue";
import WelcomeSecondScreen from "./WelcomeSecondScreen.vue";
import WelcomeThirdScreen from "./WelcomeThirdScreen.vue";
import WelcomeFourthScreen from "./WelcomeFourthScreen.vue";
import WindowDialog from "./../WindowDialog.vue";
import handleError from "@/utils/handleError";
import useSnackbar from "@/helpers/snackbar";
import { IDevice } from "@/interfaces/IDevice";
import useDevicesStore from "@/store/modules/devices";
import useNotificationsStore from "@/store/modules/notifications";
import useStatsStore from "@/store/modules/stats";

type Timer = ReturnType<typeof setInterval>;

const showDialog = defineModel<boolean>({ required: true });
const devicesStore = useDevicesStore();
const { fetchNotifications } = useNotificationsStore();
const statsStore = useStatsStore();
const snackbar = useSnackbar();
const el = ref<number>(1);
const firstPendingDevice = ref<IDevice>();
const polling = ref<Timer | undefined>(undefined);
const enable = ref(false);
const pollingDevices = () => {
  polling.value = setInterval(async () => {
    try {
      await statsStore.fetchStats();

      enable.value = statsStore.stats.pending_devices !== 0;
      if (enable.value) {
        el.value = 3;
        clearTimeout(polling.value);
      }
    } catch (error: unknown) {
      snackbar.showError("Failed to fetch devices.");
    }
  }, 3000);
};

const activePollingDevices = () => {
  el.value = 2;
  pollingDevices();
};

const acceptDevice = async () => {
  try {
    if (firstPendingDevice.value) {
      await devicesStore.acceptDevice(firstPendingDevice.value.uid);

      await fetchNotifications();
      await statsStore.fetchStats();

      el.value = 4;
    }
  } catch (error: unknown) {
    snackbar.showError("Failed to accept device.");
    handleError(error);
  }
};

const close = () => {
  showDialog.value = false;
  if (polling.value) {
    clearTimeout(polling.value);
  }
};

const handleConfirm = async () => {
  if (el.value === 1) {
    activePollingDevices();
  } else if (el.value === 2) {
    el.value = 3;
  } else if (el.value === 3) {
    await acceptDevice();
  } else if (el.value === 4) {
    close();
  }
};

const goToPreviousStep = () => {
  el.value--;
};

const goToNextStep = () => {
  el.value++;
};

defineExpose({ el, goToPreviousStep, goToNextStep, enable });
</script>
