<template>
  <fragment>
    <p class="ml-4 pt-4">
      In order to register a device on ShellHub, you need to install ShellHub agent onto it.
    </p>
    <p class="mt-4 ml-4 mr-4">
      The easiest way to install ShellHub agent is with our automatic one-line installation script,
      which works with all Linux distributions that have Docker installed and properly set up.
    </p>
    <div class="mt-4 ml-4 mr-4">
      <strong>Run the following command on your device:</strong>
      <v-text-field
        class="code"
        fill-width
        outlined
        readonly
        dense
        append-icon="mdi-content-copy"
        :value="command"
        data-test="command-text"
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
    </div>
  </fragment>
</template>

<script>

export default {
  name: 'WelcomeSecondScreenComponent',

  props: {
    command: {
      type: String,
      required: true,
    },
  },

  methods: {
    copyCommand() {
      this.$clipboard(this.command);
      this.$store.dispatch('snackbar/showSnackbarCopy', this.$copy.command);
    },
  },

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
