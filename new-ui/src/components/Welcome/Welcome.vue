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
        <span>Step</span>
        {{ el }}
        <span>of</span>
        4
        <v-divider class="mt-2" />
      </v-card-title>
      <v-window v-model="el">
        <v-window-item :value="1">
          <v-card class="bg-v-theme-surface" height="250px">
            <WelcomeFirstScreen />
          </v-card>
          <v-card-actions class="mt-4">
            <v-btn text @click="close">Close</v-btn>
            <v-spacer />
            <v-btn
              data-test="firstClick-btn"
              color="primary"
              @click="activePollingDevices()"
              >Next</v-btn
            >
          </v-card-actions>
        </v-window-item>

        <v-window-item :value="2">
          <v-card class="bg-v-theme-surface" height="250px">
            <WelcomeSecondScreen :command="command()" />
          </v-card>
          <v-card-actions>
            <v-btn text data-test="close-btn" @click="close"> Close </v-btn>
            <v-spacer />
            <v-btn text @click="el--">back</v-btn>
            <v-btn v-if="!enable" disabled> Waiting for Device </v-btn>
            <v-btn
              v-else
              color="primary"
              data-test="secondClick-btn"
              @click="el = 3"
            >
              Next
            </v-btn>
          </v-card-actions>
        </v-window-item>

        <v-window-item :value="3">
          <v-card class="bg-v-theme-surface" height="250px">
            <WelcomeThirdScreen v-if="enable" />
          </v-card>
          <v-card-actions>
            <v-btn variant="text" data-test="close-btn" @click="close">
              Close
            </v-btn>
            <v-spacer />
            <v-btn variant="text" @click="el--">back</v-btn>
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
          <v-card class="bg-v-theme-surface" height="250px">
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

<script lang="ts">
import { INotificationsError } from "../../interfaces/INotifications";
import { computed, defineComponent, ref } from "vue";
import { useStore } from "../../store";
import WelcomeFirstScreen from "./WelcomeFirstScreen.vue";
import WelcomeSecondScreen from "./WelcomeSecondScreen.vue";
import WelcomeThirdScreen from "./WelcomeThirdScreen.vue";
import WelcomeFourthScreen from "./WelcomeFourthScreen.vue";

export default defineComponent({
  props: {
    show: {
      type: Boolean,
      required: true,
    },
  },
  emits: ["update"],
  setup(props, ctx) {
    const store = useStore();
    const el = ref<number>(1);
    const polling = ref<any>(null);
    const enable = ref(false);
    const showWelcome = computed({
      get() {
        return props.show;
      },
      set(value: boolean) {
        ctx.emit("update", value);
      },
    });

    const curl = ref({
      hostname: window.location.hostname,
      tenant: store.getters["auth/tenant"],
    });

    const activePollingDevices = () => {
      el.value = 2;
      pollingDevices();
    };

    const pollingDevices = () => {
      polling.value = setInterval(async () => {
        try {
          await store.dispatch("stats/get");

          enable.value = store.getters["stats/stats"].pending_devices !== 0;
          if (enable.value) {
            el.value = 3;
            clearTimeout(polling.value);
          }
        } catch (error: any) {
          store.dispatch("snackbar/showSnackbarErrorDefault");
          throw new Error(error);
        }
      }, 3000);
    };

    const acceptDevice = async () => {
      const device = store.getters["devices/getFirstPending"];
      try {
        await store.dispatch("devices/accept", device.uid);

        store.dispatch("notifications/fetch");
        store.dispatch("stats/get");

        el.value = 4;
      } catch (error: any) {
        store.dispatch(
          "snackbar/showSnackbarErrorAction",
          INotificationsError.deviceAccepting
        );
        throw new Error(error);
      }
    };

    const command = () => {
      let port = window.location.port ? `:${window.location.port}` : "";
      let hostname = window.location.hostname;

      return `curl -sSf "${window.location.protocol}//${hostname}${port}/install.sh?tenant_id=${curl.value.tenant}" | sh`;
    };

    const close = () => {
      ctx.emit("update", false);
      showWelcome.value = false;
      clearTimeout(polling.value);
    };

    return {
      el,
      showWelcome,
      enable,
      polling,
      curl,
      close,
      activePollingDevices,
      acceptDevice,
      command,
    };
  },
  components: {
    WelcomeFirstScreen,
    WelcomeSecondScreen,
    WelcomeThirdScreen,
    WelcomeFourthScreen,
  },
});
</script>
