<template>
  <v-btn
    @click="dialog = !dialog"
    color="primary"
    tabindex="0"
    variant="elevated"
    aria-label="Dialog Add device"
    @keypress.enter="dialog = !dialog"
    data-test="device-add-btn"
    :size="size"
  >
    Add Device
  </v-btn>

  <v-dialog v-model="dialog" width="800" transition="dialog-bottom-transition">
    <v-card class="bg-v-theme-surface">
      <v-card-title class="text-h5 pa-4 bg-primary">
        Registering a device
      </v-card-title>

      <v-card-text class="mt-4 mb-0 pb-1">
        <p class="text-body-2 mb-2">
          In order to register a device on ShellHub, you need to install
          ShellHub agent onto it.
        </p>

        <p class="text-body-2 mb-2">
          The easiest way to install ShellHub agent is with our automatic
          one-line installation script, which works with all Linux distributions
          that have Docker installed and properly set up.
        </p>

        <p class="text-body-2 font-weight-bold mt-4">
          Run the following command on your device:
        </p>

        <v-text-field
          :model-value="command()"
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

        <p class="text-caption mt-2 mb-0">
          Check the
          <a
            :href="'https://docs.shellhub.io/user-guides/devices/adding'"
            target="_blank"
            rel="noopener noreferrer"
          >documentation</a
          >
          for more information and alternative install methods.
        </p>
      </v-card-text>

      <v-card-actions>
        <v-spacer />
        <v-btn variant="text" data-test="close-btn" @click="dialog = !dialog">
          Close
        </v-btn>
      </v-card-actions>
    </v-card>
  </v-dialog>
</template>

<script lang="ts">
import { computed, defineComponent, ref } from "vue";
import { useStore } from "../../store";
import { INotificationsCopy } from "@/interfaces/INotifications";

export default defineComponent({
  props: {
    size: {
      type: String,
      default: "default",
      required: false,
    },
  },
  setup() {
    const store = useStore();

    const dialog = ref(false);

    const tenant = computed(() => store.getters["auth/tenant"]);

    const command = () => {
      const port = window.location.port ? `:${window.location.port}` : "";
      const { hostname } = window.location;

      return `curl -sSf "${window.location.protocol}//${hostname}${port}/install.sh?tenant_id=${tenant.value}" | sh`;
    };

    const copyCommand = () => {
      navigator.clipboard.writeText(command());
      store.dispatch("snackbar/showSnackbarCopy", INotificationsCopy.command);
    };

    return {
      tenant,
      dialog,
      command,
      copyCommand,
    };
  },
});
</script>

<style lang="scss" scoped>
.code {
  font-family: monospace;
  font-size: 85%;
  font-weight: normal;
}
</style>
