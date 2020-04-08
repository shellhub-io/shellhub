<template>
<v-dialog v-model="show" max-width="800px">
  <v-card>

    <v-card-title class="headline grey lighten-2 text-center" primary-title>
      Registering a device
    </v-card-title>

    <v-card-text class="mt-4 mb-0 pb-1">
      <p>In order to register a device on ShellHub, you need to install ShellHub agent onto it.</p>
      <p>The easiest way to install ShellHub agent is with our automatic one-line installation script,
        which works with all Linux distributions that have Docker installed and properly set up.</p>

      <strong>Run the following command on your device:</strong>
      <v-text-field class="code" fill-width outlined readonly dense append-icon="mdi-content-copy" @click:append="copyCommand" :value="command()"/>

      <v-divider></v-divider>

      <p class="caption mb-0">
        Check the <a :href="'https://shellhub-io.github.io/guides/registering-device/'" target="_blank">documentation</a>
        for more information and alternative install methods.
      </p>
    </v-card-text>

    <v-card-actions>
      <v-spacer></v-spacer>
      <v-btn text @click="show = false">Close</v-btn>
    </v-card-actions>

    <v-snackbar v-model="copySnack" :timeout=3000>Command copied to clipboard</v-snackbar>
  </v-card>

</v-dialog>
</template>

<script>
export default {
  name: "DeviceAdd",

  data() {
    return {
      hostname: window.location.hostname,
      copySnack: false
    };
  },

  computed: {
    tenant() {
      return this.$store.getters["auth/tenant"];
    },

    show: {
      get() {
        return this.$store.getters["modals/add_device"];
      },

      set(value) {
        this.$store.dispatch("modals/showAddDevice", value);
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