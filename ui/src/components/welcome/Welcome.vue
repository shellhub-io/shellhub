<template>
  <v-card>
    <v-dialog
      v-model="showWelcome"
      :retain-focus="false"
      max-width="800px"
      persistent
    >
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
              height="250px"
            >
              <WelcomeFirstScreen />
            </v-card>
            <v-card-actions>
              <v-btn
                text
                @click="close"
              >
                Close
              </v-btn>
              <v-spacer />
              <v-btn
                color="primary"
                data-test="firstClick-btn"
                @click="activePollingDevices()"
              >
                Next
              </v-btn>
            </v-card-actions>
          </v-stepper-content>

          <v-stepper-content step="2">
            <v-card
              class="mb-12"
              height="250px"
            >
              <WelcomeSecondScreen
                :command="command()"
              />
            </v-card>
            <v-card-actions>
              <v-btn
                text
                data-test="close-btn"
                @click="close"
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
                data-test="secondClick-btn"
                @click="e1 = 3"
              >
                Next
              </v-btn>
            </v-card-actions>
          </v-stepper-content>

          <v-stepper-content step="3">
            <v-card
              class="mb-12"
              height="250px"
            >
              <WelcomeThirdScreen
                v-if="enable"
              />
            </v-card>
            <v-card-actions>
              <v-btn
                text
                @click="close"
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
                data-test="thirdClick-btn"
                @click="acceptDevice()"
              >
                Accept
              </v-btn>
            </v-card-actions>
          </v-stepper-content>

          <v-stepper-content step="4">
            <v-card
              class="mb-12"
              height="250px"
            >
              <WelcomeFourthScreen />
            </v-card>
            <v-card-actions>
              <v-spacer />
              <v-btn
                color="primary"
                @click="close"
              >
                Finish
              </v-btn>
            </v-card-actions>
          </v-stepper-content>
        </v-stepper-items>
      </v-stepper>
    </v-dialog>
  </v-card>
</template>

<script>

import WelcomeFirstScreen from '@/components/welcome/WelcomeFirstScreen';
import WelcomeSecondScreen from '@/components/welcome/WelcomeSecondScreen';
import WelcomeThirdScreen from '@/components/welcome/WelcomeThirdScreen';
import WelcomeFourthScreen from '@/components/welcome/WelcomeFourthScreen';

export default {
  name: 'WelcomeComponent',

  components: {
    WelcomeFirstScreen,
    WelcomeSecondScreen,
    WelcomeThirdScreen,
    WelcomeFourthScreen,
  },

  props: {
    show: {
      type: Boolean,
      required: true,
    },
  },

  data() {
    return {
      e1: 1,
      enable: false,
      polling: null,
      curl: {
        hostname: window.location.hostname,
        tenant: this.$store.getters['auth/tenant'],
      },
    };
  },

  computed: {
    showWelcome: {
      get() {
        return this.show;
      },

      set(value) {
        this.$emit('show', value);
      },
    },
  },

  methods: {
    command() {
      return `curl -sSf "${window.location.protocol}//${this.curl.hostname}/install.sh?tenant_id=${this.curl.tenant}" | sh`;
    },

    activePollingDevices() {
      this.e1 = 2;
      this.pollingDevices();
    },

    pollingDevices() {
      this.polling = setInterval(async () => {
        try {
          await this.$store.dispatch('stats/get');

          this.enable = this.$store.getters['stats/stats'].pending_devices !== 0;
          if (this.enable) {
            this.e1 = 3;
            clearTimeout(this.polling);
          }
        } catch {
          this.$store.dispatch('snackbar/showSnackbarErrorDefault');
        }
      }, 3000);
    },

    async acceptDevice() {
      const device = this.$store.getters['devices/getFirstPending'];
      try {
        await this.$store.dispatch('devices/accept', device.uid);

        this.$store.dispatch('notifications/fetch');
        this.$store.dispatch('stats/get');

        this.e1 = 4;
      } catch {
        this.$store.dispatch('snackbar/showSnackbarErrorAction', this.$errors.snackbar.deviceAccepting);
      }
    },

    async close() {
      this.$emit('update:show', false);
      clearTimeout(this.polling);
    },
  },
};
</script>
