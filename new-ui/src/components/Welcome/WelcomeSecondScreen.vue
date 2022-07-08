<template>
  <p class="ml-4 pt-4 text-subtitle-2">
    In order to register a device on ShellHub, you need to install ShellHub
    agent onto it.
  </p>
  <p class="ml-4 pt-4 text-subtitle-2">
    The easiest way to install ShellHub agent is with our automatic one-line
    installation script, which works with all Linux distributions that have
    Docker installed and properly set up.
  </p>
  <div class="mt-4 ml-4 mr-4">
    <p class="ml-2 pt-4 text-subtitle-2 text-bold">
      Run the following command on your device:
    </p>

    <v-text-field
      :model-value="command"
      @click:append="copyCommand"
      class="code mt-1"
      variant="outlined"
      append-icon="mdi-content-copy"
      readonly
      active
      data-test="command-field"
      density="compact"
    />

    <v-divider />

    <p class="text-caption mb-0 mt-1">
      Check the
      <a
        :href="'https://shellhub-io.github.io/guides/registering-device/'"
        target="_blank"
        >documentation</a
      >
      for more information and alternative install methods.
    </p>
  </div>
</template>

<script lang="ts">
import { INotificationsCopy } from "../../interfaces/INotifications";
import { useStore } from "../../store";
import { defineComponent } from "vue";

export default defineComponent({
  props: {
    command: {
      type: String,
      required: true,
    },
  },

  setup(props) {
    const store = useStore();
    const copyCommand = () => {
      navigator.clipboard.writeText(props.command);
      store.dispatch("snackbar/showSnackbarCopy", INotificationsCopy.tenantId);
    };

    return {
      copyCommand,
    };
  },
});
</script>
