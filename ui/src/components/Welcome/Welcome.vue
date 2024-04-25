<template>
  <v-dialog
    v-model="showWelcome"
    max-width="800px"
    min-width="60vw"
    :retain-focus="false"
    persistent
  >
    <v-card class="pa-6 bg-grey-darken-4 bg-v-theme-surface">
      <v-card-title class="text-center mb-4">
        <span data-test="step-counter">Step {{ el }} of 4</span>
        <v-divider class="mt-2" />
      </v-card-title>
      <v-window v-model="el">
        <v-window-item :value="1">
          <v-card class="bg-v-theme-surface" height="200px" :elevation="0" data-test="welcome-first-screen">
            <WelcomeFirstScreen />
          </v-card>
          <v-card-actions class="mt-4">
            <v-btn @click="close" data-test="close-btn">Close</v-btn>
            <v-spacer />
            <v-btn
              data-test="first-click-btn"
              color="primary"
              @click="activePollingDevices()"
            >Next</v-btn>
          </v-card-actions>
        </v-window-item>

        <v-window-item :value="2">
          <v-card class="bg-v-theme-surface" height="250px" :elevation="0" data-test="welcome-second-screen">
            <WelcomeSecondScreen :command="command()" />
          </v-card>
          <v-card-actions>
            <v-btn data-test="close2-btn" @click="close">Close</v-btn>
            <v-spacer />
            <v-btn @click="goToPreviousStep" data-test="back-btn">Back</v-btn>
            <v-btn v-if="!enable" data-test="waiting-message" disabled>Waiting for Device</v-btn>
            <v-btn
              v-else
              color="primary"
              data-test="next-btn"
              @click="goToNextStep"
            >
              Next
            </v-btn>
          </v-card-actions>
        </v-window-item>

        <v-window-item :value="3">
          <v-card class="bg-v-theme-surface" height="250px" :elevation="0" data-test="welcome-third-screen">
            <WelcomeThirdScreen v-if="enable" />
          </v-card>
          <v-card-actions>
            <v-btn variant="text" data-test="close3-btn" @click="close">
              Close
            </v-btn>
            <v-spacer />
            <v-btn variant="text" @click="goToPreviousStep" data-test="back2-btn">Back</v-btn>
            <v-btn
              color="primary"
              data-test="accept-btn"
              @click="acceptDevice()"
            >
              Accept
            </v-btn>
          </v-card-actions>
        </v-window-item>

        <v-window-item :value="4">
          <v-card class="bg-v-theme-surface" height="250px" :elevation="0" data-test="welcome-fourth-screen">
            <WelcomeFourthScreen />
          </v-card>
          <v-card-actions>
            <v-spacer />
            <v-btn color="primary" data-test="finish-btn" @click="close">
              Finish
            </v-btn>
          </v-card-actions>
        </v-window-item>
      </v-window>
    </v-card>
  </v-dialog>
</template>

<script setup lang="ts">
import { computed, ref } from "vue";
import { INotificationsError } from "../../interfaces/INotifications";
import { useStore } from "../../store";
import WelcomeFirstScreen from "./WelcomeFirstScreen.vue";
import WelcomeSecondScreen from "./WelcomeSecondScreen.vue";
import WelcomeThirdScreen from "./WelcomeThirdScreen.vue";
import WelcomeFourthScreen from "./WelcomeFourthScreen.vue";
import handleError from "../../utils/handleError";

type Timer = ReturnType<typeof setInterval>;

const props = defineProps({
  show: {
    type: Boolean,
    required: true,
  },
});
const emit = defineEmits(["update"]);
const store = useStore();
const el = ref<number>(1);
const polling = ref<Timer | undefined>(undefined);
const enable = ref(false);
const showWelcome = computed<boolean>({
  get() {
    return props.show;
  },
  set(value) {
    emit("update", value);
  },
});

const curl = ref({
  hostname: window.location.hostname,
  tenant: store.getters["auth/tenant"],
});

const pollingDevices = () => {
  polling.value = setInterval(async () => {
    try {
      await store.dispatch("stats/get");

      enable.value = store.getters["stats/stats"].pending_devices !== 0;
      if (enable.value) {
        el.value = 3;
        clearTimeout(polling.value);
      }
    } catch (error: unknown) {
      store.dispatch("snackbar/showSnackbarErrorDefault");
      handleError(error);
    }
  }, 3000);
};

const activePollingDevices = () => {
  el.value = 2;
  pollingDevices();
};

const acceptDevice = async () => {
  const device = store.getters["devices/getFirstPending"];
  try {
    if (device) {
      await store.dispatch("devices/accept", device.uid);

      store.dispatch("notifications/fetch");
      store.dispatch("stats/get");

      el.value = 4;
    }
  } catch (error: unknown) {
    store.dispatch(
      "snackbar/showSnackbarErrorAction",
      INotificationsError.deviceAccepting,
    );
    handleError(error);
  }
};

const command = () => {
  const port = window.location.port ? `:${window.location.port}` : "";
  const { hostname } = window.location;

  return `curl -sSf "${window.location.protocol}//${hostname}${port}/install.sh?tenant_id=${curl.value.tenant}" | sh`;
};

const close = () => {
  emit("update", false);
  showWelcome.value = false;
  if (polling.value) {
    clearTimeout(polling.value);
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
