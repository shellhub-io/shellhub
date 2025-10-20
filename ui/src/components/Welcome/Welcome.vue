<template>
  <WindowDialog
    v-model="showDialog"
    title="Welcome to ShellHub!"
    :description="`Step ${currentStep} of 4`"
    icon="mdi-door-open"
    icon-color="primary"
    @close="closeDialog"
  >
    <v-window v-model="currentStep" class="overflow-y-auto" data-test="welcome-window">
      <v-window-item :value="1" data-test="welcome-first-screen">
        <WelcomeFirstScreen />
      </v-window-item>
      <v-window-item :value="2" data-test="welcome-second-screen">
        <WelcomeSecondScreen />
      </v-window-item>
      <v-window-item :value="3" data-test="welcome-third-screen">
        <WelcomeThirdScreen v-if="hasDeviceDetected" v-model:first-pending-device="firstPendingDevice" />
      </v-window-item>
      <v-window-item :value="4" data-test="welcome-fourth-screen">
        <WelcomeFourthScreen />
      </v-window-item>
    </v-window>

    <template #footer>
      <div class="d-flex align-center w-100">
        <p v-if="currentStep === 2" class="text-caption text-truncate" data-test="second-screen-helper-link">
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
          v-if="currentStep !== 4"
          @click="closeDialog"
          data-test="cancel-btn"
          class="mr-2"
          text="Close"
        />
        <v-btn
          color="primary"
          @click="handleConfirm"
          :disabled="currentStep === 2 && !hasDeviceDetected"
          data-test="confirm-btn"
          :text="currentStepConfig.buttonText"
        />
      </div>
    </template>
  </WindowDialog>
</template>

<script setup lang="ts">
import { computed, ref } from "vue";
import WelcomeFirstScreen from "./WelcomeFirstScreen.vue";
import WelcomeSecondScreen from "./WelcomeSecondScreen.vue";
import WelcomeThirdScreen from "./WelcomeThirdScreen.vue";
import WelcomeFourthScreen from "./WelcomeFourthScreen.vue";
import WindowDialog from "@/components/Dialogs/WindowDialog.vue";
import handleError from "@/utils/handleError";
import useSnackbar from "@/helpers/snackbar";
import { IDevice } from "@/interfaces/IDevice";
import useDevicesStore from "@/store/modules/devices";
import useNotificationsStore from "@/store/modules/notifications";
import useStatsStore from "@/store/modules/stats";

type PollingTimer = ReturnType<typeof setInterval>;

interface StepConfig {
  buttonText: string;
  action: () => void | Promise<void>;
}

const showDialog = defineModel<boolean>({ required: true });
const devicesStore = useDevicesStore();
const { fetchNotifications } = useNotificationsStore();
const statsStore = useStatsStore();
const snackbar = useSnackbar();
const currentStep = ref<number>(1);
const firstPendingDevice = ref<IDevice>();
const pollingTimer = ref<PollingTimer | undefined>(undefined);
const hasDeviceDetected = ref(false);

const stopDevicePolling = () => {
  if (pollingTimer.value) {
    clearInterval(pollingTimer.value);
    pollingTimer.value = undefined;
  }
};

const startDevicePolling = () => {
  currentStep.value = 2;
  pollingTimer.value = setInterval(async () => {
    try {
      await statsStore.fetchStats();

      const hasPendingDevices = statsStore.stats.pending_devices > 0;
      if (hasPendingDevices) {
        hasDeviceDetected.value = true;
        currentStep.value = 3;
        stopDevicePolling();
      }
    } catch (error: unknown) {
      snackbar.showError("Failed to fetch devices.");
    }
  }, 3000);
};

const acceptDevice = async () => {
  if (!firstPendingDevice.value) return;

  try {
    await devicesStore.acceptDevice(firstPendingDevice.value.uid);
    await fetchNotifications();
    await statsStore.fetchStats();
    currentStep.value = 4;
  } catch (error: unknown) {
    snackbar.showError("Failed to accept device.");
    handleError(error);
  }
};

const closeDialog = () => {
  showDialog.value = false;
  stopDevicePolling();
};

const stepConfigs: Record<number, StepConfig> = {
  1: {
    buttonText: "Next",
    action: startDevicePolling,
  },
  2: {
    buttonText: "Next",
    action: () => { currentStep.value = 3; },
  },
  3: {
    buttonText: "Accept",
    action: acceptDevice,
  },
  4: {
    buttonText: "Finish",
    action: closeDialog,
  },
};

const currentStepConfig = computed(() => stepConfigs[currentStep.value]);
const handleConfirm = async () => { await currentStepConfig.value.action(); };

defineExpose({ currentStep, hasDeviceDetected });
</script>
