<template>
  <v-btn
    @click="dialog = !dialog"
    color="primary"
    tabindex="0"
    variant="elevated"
    aria-label="Dialog Add Docker"
    @keypress.enter="dialog = !dialog"
    data-test="device-add-btn"
    :size="props.size"
  >
    Add Docker Host
  </v-btn>

  <v-dialog v-model="dialog" width="800" transition="dialog-bottom-transition" data-test="dialog">
    <v-card class="bg-v-theme-surface">
      <v-card-title class="text-h5 pa-4 bg-primary" data-test="dialog-title">
        Registering a Docker host
      </v-card-title>

      <v-card-text class="mt-4 mb-0 pb-1" data-test="dialog-text">
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

        <v-divider />

        <p class="text-caption mt-2 mb-0">
          Check the
          <a
            :href="'https://docs.shellhub.io/overview/supported-platforms/docker'"
            target="_blank"
            rel="noopener noreferrer"
            data-test="documentation-link"
          >documentation</a
          >
          for more information about integration with Docker host.
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

<script setup lang="ts">
import { computed, ref } from "vue";
import { useStore } from "@/store";
import CopyWarning from "@/components/User/CopyWarning.vue";

const props = defineProps({
  size: {
    type: String,
    default: "default",
    required: false,
  },
});
const store = useStore();
const dialog = ref(false);

const tenant = computed(() => store.getters["auth/tenant"]);

const command = () => {
  const port = window.location.port ? `:${window.location.port}` : "";
  const { hostname } = window.location;

  // eslint-disable-next-line vue/max-len
  return `curl -sSf ${window.location.protocol}//${hostname}${port}/install.sh | TENANT_ID=${tenant.value} SERVER_ADDRESS=${window.location.protocol}//${hostname}${port} sh -s connector`;
};
</script>

<style lang="scss" scoped>
.code {
  font-family: monospace;
  font-size: 85%;
  font-weight: normal;
}
</style>
