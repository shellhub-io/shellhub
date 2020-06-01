<template>
  <v-dialog
    v-model="dialog"
    :retain-focus="false"
    persistent
    max-width="800px"
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
            Conecting device
          </v-stepper-step>

          <v-divider />

          <v-stepper-step step="3">
            Congratulations
          </v-stepper-step>
        </v-stepper-header>

        <v-stepper-items>
          <v-stepper-content step="1">
            <v-card 
              class="mb-12"
              color="grey lighten-1"
              height="200px"
            >
              <WelcomeFirstScreen />
            </v-card>

            <v-btn
              color="primary"
              @click="e1 = 2"
            >
              Continue
            </v-btn>

            <v-btn text>
              Cancel
            </v-btn>
          </v-stepper-content>

          <v-stepper-content step="2">
            <v-card 
              class="mb-12"
              color="grey lighten-1"
              height="200px"
            >
              <WelcomeSecondScreen :command="command()" />
            </v-card>

            <v-btn
              color="primary"
              @click="e1 = 3"
            >
              Continue
            </v-btn>

            <v-btn text>
              Cancel
            </v-btn>
          </v-stepper-content>

          <v-stepper-content step="3">
            <v-card 
              class="mb-12"
              color="grey lighten-1"
              height="200px"
            >
              <WelcomeThirdScreen :command="command()" />
            </v-card>

            <v-btn
              color="primary"
              @click="e1 = 1"
            >
              Continue
            </v-btn>

            <v-btn text>
              Cancel
            </v-btn>
          </v-stepper-content>
        </v-stepper-items>
      </v-stepper>
    </v-card>
  </v-dialog>
</template>

<script>

import WelcomeFirstScreen from '@/components/welcome/WelcomeFirstScreen.vue';
import WelcomeSecondScreen from '@/components/welcome/WelcomeSecondScreen.vue';
import WelcomeThirdScreen from '@/components/welcome/WelcomeThirdScreen.vue';

export default {
  name: 'Welcome',

  components: {
    WelcomeFirstScreen,
    WelcomeSecondScreen,
    WelcomeThirdScreen
  },

  props: {
    dialog: {
      type: Boolean,
      required: true
    },
  },

  data () {
    return {
      e1: 1,
    };
  },

  created(){
    this.pollingDevices();
  },

  methods: {
    pollingDevices(){
      this.polling = setInterval(async() => {
        await this.$store.dispatch('stats/get', {});

        if(this.$store.getters['stats/stats'].registered_devices !== 0){
          this.beforeDestroy();
        }
      }, 3000);
    },

    beforeDestroy () {
      clearInterval(this.polling);
    },

    command() {
      return `curl "${location.protocol}//${this.hostname}/install.sh?tenant_id=${this.tenant}" | sh`;
    },
  }
};
</script>