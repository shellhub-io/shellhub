<template>
  <v-dialog
    v-model="show"
    :retain-focus="false"
    max-width="800px"
    persistent
  >
    <v-card>
      <v-stepper v-model="e1">
        <v-stepper-header>
          <v-stepper-step
            :complete="e1 > 1"
            step="1"
          >
            Welcome
          </v-stepper-step>

          <v-divider />

          <v-stepper-step
            :complete="e1 > 2"
            step="2"
          >
            Connecting device
          </v-stepper-step>

          <v-divider />

          <v-stepper-step
            :complete="e1 > 3"
            step="3"
          >
            Authorizing device
          </v-stepper-step>

          <v-divider />

          <v-stepper-step step="4">
            Finish
          </v-stepper-step>
        </v-stepper-header>

        <v-stepper-items>
          <v-stepper-content step="1">
            <v-card
              class="mb-12"
              color="grey lighten-4"
              height="250px"
            >
              <WelcomeFirstScreen />
            </v-card>
            <v-card-actions>
              <v-btn
                text
                @click="finished"
              >
                Close
              </v-btn>
              <v-spacer />
              <v-btn
                color="primary"
                @click="e1 = 2"
              >
                Next
              </v-btn>
            </v-card-actions>
          </v-stepper-content>

          <v-stepper-content step="2">
            <v-card
              class="mb-12"
              color="grey lighten-4"
              height="250px"
            >
              <WelcomeSecondScreen
                :command="command()"
                @expClip="receiveClip"
              />
            </v-card>
            <v-card-actions>
              <v-btn
                text
                @click="finished"
              >
                Close
              </v-btn>
              <v-spacer />
              <v-btn
                @click="e1 = 1"
              >
                Back
              </v-btn>
              <v-btn
                v-if="!enable"
                disabled
              >
                Waiting for Device
              </v-btn>
              <v-btn
                v-else
                color="primary"
                @click="e1 = 3"
              >
                Next
              </v-btn>
            </v-card-actions>
            <v-snackbar
              v-model="copy"
              :timeout="3000"
            >
              Command copied to clipboard
            </v-snackbar>
          </v-stepper-content>

          <v-stepper-content step="3">
            <v-card
              class="mb-12"
              color="grey lighten-4"
              height="250px"
            >
              <WelcomeThirdScreen
                v-if="enable"
              />
            </v-card>
            <v-card-actions>
              <v-btn
                text
                @click="finished"
              >
                Close
              </v-btn>
              <v-spacer />
              <v-btn
                @click="e1 = 2"
              >
                Back
              </v-btn>
              <v-btn
                color="primary"
                @click="acceptDevice()"
              >
                Accept
              </v-btn>
            </v-card-actions>
            <v-snackbar
              v-model="copy"
              :timeout="3000"
            >
              Command copied to clipboard
            </v-snackbar>
          </v-stepper-content>

          <v-stepper-content step="4">
            <v-card
              class="mb-12"
              color="grey lighten-4"
              height="250px"
            >
              <WelcomeFourthScreen :command="command()" />
            </v-card>
            <v-card-actions>
              <v-spacer />
              <v-btn
                color="primary"
                @click="finished"
              >
                Finish
              </v-btn>
            </v-card-actions>
          </v-stepper-content>
        </v-stepper-items>
      </v-stepper>
    </v-card>
  </v-dialog>
</template>

<script>

import WelcomeFirstScreen from '@/components/welcome/WelcomeFirstScreen';
import WelcomeSecondScreen from '@/components/welcome/WelcomeSecondScreen';
import WelcomeThirdScreen from '@/components/welcome/WelcomeThirdScreen';
import WelcomeFourthScreen from '@/components/welcome/WelcomeFourthScreen';

export default {
  name: 'Welcome',

  components: {
    WelcomeFirstScreen,
    WelcomeSecondScreen,
    WelcomeThirdScreen,
    WelcomeFourthScreen,
  },

  props: {
    dialog: {
      type: Boolean,
      required: true,
    },

    curl: {
      type: Object,
      required: true,
    },
  },

  data() {
    return {
      e1: 1,
      copy: false,
      enable: false,
      polling: null,
      trigger: null,
    };
  },

  computed: {
    show: {
      get() {
        return this.dialog;
      },

      set(value) {
        this.$emit('show', value);
      },
    },
  },

  created() {
    this.pollingDevices();
  },

  methods: {
    receiveClip(params) {
      this.copy = params;
    },

    beforeDestroy() {
      clearInterval(this.polling);
    },

    command() {
      return `curl "${window.location.protocol}//${this.curl.hostname}/install.sh?tenant_id=${this.curl.tenant}" | sh`;
    },

    finished() {
      clearTimeout(this.trigger);
      this.show = false;
      this.$emit('finishedEvent', false);
      this.beforeDestroy();
    },

    pollingDevices() {
      this.polling = setInterval(async () => {
        await this.$store.dispatch('stats/get', {});
        this.enable = this.checkDevice();
        if (this.enable) {
          this.e1 = 3;
        }
      }, 3000);
    },

    checkDevice() {
      return this.$store.getters['stats/stats'].pending_devices !== 0;
    },

    acceptDevice() {
      const device = this.$store.getters['devices/getFirstPending'];
      this.$store.dispatch('devices/accept', device.uid);

      this.$store.dispatch('notifications/fetch');
      this.$store.dispatch('stats/get');

      this.e1 = 4;
    },
  },
};

</script>
