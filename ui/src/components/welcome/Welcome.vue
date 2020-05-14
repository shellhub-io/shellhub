<template>
  <v-dialog
    v-model="show"
    :retain-focus="false"
    persistent
    max-width="800px"
  >
    <v-card>
      <v-card-title
        class="headline grey lighten-2 text-center"
        primary-title
      >
        Welcome to ShellHub
      </v-card-title>

      <v-card-actions v-if="screenFirst">
        <WelcomeFirstScreen />
      </v-card-actions>

      <v-card-actions v-if="!screenFirst">
        <WelcomeSecondScreen :command="command()" />
      </v-card-actions>

      <v-card-actions>
        <v-btn
          v-if="!screenFirst"
          text
          @click="backScreen"
        >
          Back
        </v-btn>

        <v-spacer />
        
        <v-btn
          text
          @click="nextScren"
        >
          Next
        </v-btn>
      </v-card-actions>

      <v-snackbar
        v-model="allowsUserContinue"
        :timeout="3000"
      >
        Please! You should enter a device to continue in the ShellHub
      </v-snackbar>
      <v-snackbar
        v-model="copySnack"
        :timeout="3000"
      >
        Command copied to clipboard
      </v-snackbar>
    </v-card>
  </v-dialog>
</template>

<script>
import WelcomeFirstScreen from '@/components/welcome/WelcomeFirstScreen.vue';
import WelcomeSecondScreen from '@/components/welcome/WelcomeSecondScreen.vue';


export default {
  name: 'Welcome',

  components: {
    WelcomeFirstScreen,
    WelcomeSecondScreen,
  },

  props: {
    screenWelcome: {
      type: Boolean,
      required: true
    },
  },

  data() {
    return {
      hostname: window.location.hostname,
      copySnack: false,      
      allowsUserContinue: false,
      screenFirst: true,
    };
  },

  computed: {
    tenant() {
      return this.$store.getters['auth/tenant'];
    },

    show: {
      get() {
        return this.screenWelcome;
      },

      set(value) {
        this.screenWelcome = value;
      }
    }
  },
  methods:{
    command() {
      return `curl "${location.protocol}//${this.hostname}/install.sh?tenant_id=${this.tenant}" | sh`;
    },

    copyCommand() {
      this.$clipboard(this.command());
      this.copySnack = true;
    },

    async validadeAddDevice(){
      if(this.$store.getters['stats/stats'].registered_devices === 0){
        await this.$store.dispatch('stats/get');
        
        if(this.$store.getters['stats/stats'].registered_devices !== 0){
          this.show = false;
        } else{
          this.allowsUserContinue = true;
        }
      }
    },

    nextScren() {
      this.screenFirst = false;
    },

    backScreen() {
      this.screenFirst = true;
    }

  }
};
</script>

<style lang="scss" scoped>
@import '~vuetify/src/styles/settings/_variables.scss';

.code {
  font-family: monospace;
  font-size: $code-kbd-font-size;
  font-weight: $code-kbd-font-weight;
}
</style>