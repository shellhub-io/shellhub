<template>
  <v-btn
    @click="showDialog = true"
    color="primary"
    tabindex="0"
    variant="elevated"
    @keypress.enter="showDialog = true"
    data-test="device-add-btn"
    :size
  >
    Add Device
  </v-btn>

  <BaseDialog v-model="showDialog" transition="dialog-bottom-transition" data-test="device-add-dialog">
    <v-card class="bg-v-theme-surface">
      <v-card-title class="text-h5 pa-4 bg-primary" data-test="dialog-title">
        Registering a device
      </v-card-title>

      <v-card-text class="mt-4 mb-0 pb-1" data-test="dialog-text">
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
        <CopyWarning :copied-item="'Command'">
          <template #default="{ copyText }">
            <v-text-field
              :model-value="command()"
              @click:append="copyText(command())"
              class="code mt-1"
              variant="outlined"
              append-icon="mdi-content-copy"
              readonly
              active
              data-test="command-field"
              density="compact"
            /></template> </CopyWarning>

        <v-divider />

        <p class="text-caption mt-2 mb-0">
          Check the
          <a
            :href="'https://docs.shellhub.io/user-guides/devices/adding'"
            target="_blank"
            rel="noopener noreferrer"
            data-test="documentation-link"
          >documentation</a
          >
          for more information and alternative install methods.
        </p>
      </v-card-text>

      <v-card-actions>
        <v-spacer />
        <v-btn variant="text" data-test="close-btn" @click="showDialog = !showDialog">
          Close
        </v-btn>
      </v-card-actions>
    </v-card>
  </BaseDialog>
</template>

<script setup lang="ts">
import { ref } from "vue";
import CopyWarning from "@/components/User/CopyWarning.vue";
import BaseDialog from "../BaseDialog.vue";
import useAuthStore from "@/store/modules/auth";

const { size } = defineProps<{ size?: string }>();
const authStore = useAuthStore();
const showDialog = ref(false);
const { tenantId } = authStore;

const command = () => {
  const port = window.location.port ? `:${window.location.port}` : "";
  const { hostname } = window.location;

  // eslint-disable-next-line vue/max-len
  return `curl -sSf ${window.location.protocol}//${hostname}${port}/install.sh | TENANT_ID=${tenantId} SERVER_ADDRESS=${window.location.protocol}//${hostname}${port} sh`;
};
</script>

<style lang="scss" scoped>
.code {
  font-family: monospace;
  font-size: 85%;
  font-weight: normal;
}
</style>
