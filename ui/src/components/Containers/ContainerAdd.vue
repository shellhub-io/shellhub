<template>
  <v-btn
    @click="showDialog = true"
    color="primary"
    tabindex="0"
    variant="elevated"
    @keypress.enter="showDialog = true"
    data-test="container-add-btn"
  >
    Add Docker Host
  </v-btn>

  <WindowDialog
    v-model="showDialog"
    transition="dialog-bottom-transition"
    data-test="container-add-dialog"
    title="Registering a Docker host"
    description="Install the ShellHub Connector to add Docker containers"
    icon="mdi-docker"
    icon-color="primary"
    show-footer
    @close="showDialog = false"
  >
    <v-card-text class="mt-4 mb-0 pb-1 text-justify" data-test="dialog-text">
      <p class="text-body-2 mb-2">
        In order to add Docker containers to ShellHub, you need to install the
        ShellHub Connector on the Docker host.
      </p>

      <p class="text-body-2 mb-2">
        The easiest way to install the ShellHub Connector is with our automatic
        one-line installation script, which connects to the Docker API and exposes
        the running containers within ShellHub.
      </p>

      <p class="text-body-2 font-weight-bold mt-4">
        Run the following command on your Docker host:
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
          />
        </template>
      </CopyWarning>
    </v-card-text>

    <template #footer>
      <v-spacer />
      <v-btn variant="text" data-test="close-btn" @click="showDialog = false">
        Close
      </v-btn>
    </template>
  </WindowDialog>
</template>

<script setup lang="ts">
import { ref } from "vue";
import CopyWarning from "@/components/User/CopyWarning.vue";
import WindowDialog from "../WindowDialog.vue";
import useAuthStore from "@/store/modules/auth";

const authStore = useAuthStore();
const showDialog = ref(false);
const { tenantId } = authStore;

const command = () => {
  const port = window.location.port ? `:${window.location.port}` : "";
  const { hostname } = window.location;

  // eslint-disable-next-line vue/max-len
  return `curl -sSf ${window.location.protocol}//${hostname}${port}/install.sh | TENANT_ID=${tenantId} SERVER_ADDRESS=${window.location.protocol}//${hostname}${port} sh -s connector`;
};
</script>

<style lang="scss" scoped>
.code {
  font-family: monospace;
  font-size: 85%;
  font-weight: normal;
}
</style>
