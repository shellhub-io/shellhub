<template>
  <v-btn
    color="primary"
    tabindex="0"
    variant="elevated"
    data-test="container-add-btn"
    text="Add Docker Host"
    @click="showDialog = true"
    @keypress.enter="showDialog = true"
  />

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
    <v-card-text
      class="mt-4 mb-0 pb-1 text-justify"
      data-test="dialog-text"
    >
      <p class="text-body-2 mb-2">
        In order to add Docker containers to ShellHub, you need to install the
        ShellHub Connector on the Docker host.
      </p>

      <p class="text-body-2 mb-2">
        The easiest way to install the ShellHub Connector is with our automatic
        one-line installation script, which connects to the Docker API and exposes
        the running containers within ShellHub.
      </p>

      <p class="text-body-2 font-weight-bold mt-3">
        Run the following command on your Docker host:
      </p>
      <CopyCommandField
        :command="command"
        class="mt-1 mb-3"
      />
    </v-card-text>

    <template #footer>
      <v-spacer />
      <v-btn
        variant="text"
        data-test="close-btn"
        text="Close"
        @click="showDialog = false"
      />
    </template>
  </WindowDialog>
</template>

<script setup lang="ts">
import { ref } from "vue";
import CopyCommandField from "@/components/CopyCommandField.vue";
import WindowDialog from "@/components/Dialogs/WindowDialog.vue";
import useAuthStore from "@/store/modules/auth";

const { tenantId } = useAuthStore();
const showDialog = ref(false);
const { origin } = window.location;
const command = `curl -sSf ${origin}/install.sh | TENANT_ID=${tenantId} SERVER_ADDRESS=${origin} sh -s connector`;
</script>
