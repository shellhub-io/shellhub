<template>
  <v-card>
    <v-card-text class="mt-4 mb-0 pb-1">
      <p>In order to register a device on ShellHub, you need to install ShellHub agent onto it.</p>
      <p>
        The easiest way to install ShellHub agent is with our automatic one-line installation script,
        which works with all Linux distributions that have Docker installed and properly set up.
      </p>

      <strong>Run the following command on your device:</strong>
      <v-text-field
        class="code"
        fill-width
        outlined
        readonly
        dense
        append-icon="mdi-content-copy"
        :value="command"
        @click:append="copyCommand"
      />

      <v-divider />

      <p class="caption mb-0">
        Check the <a
          :href="'https://shellhub-io.github.io/guides/registering-device/'"
          target="_blank"
        >documentation</a>
        for more information and alternative install methods.
      </p>
    </v-card-text>

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
</template>

<script>
export default {
  name: 'WelcomeSecondScreen',

  props: {
    command: {
      type: String,
      required: true
    },
  },

  data() {
    return {
      hostname: window.location.hostname,
      copySnack: false,      
      allowsUserContinue: false
    };
  },

  methods:{
    copyCommand() {
      this.$clipboard(this.command);
      this.copySnack = true;
    },
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