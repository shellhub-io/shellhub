<template>
  <p class="ml-4 pt-4 text-subtitle-2" data-test="welcome-second-title">
    In order to register a device on ShellHub, you need to install ShellHub
    agent onto it.
  </p>
  <p class="ml-4 pt-4 text-subtitle-2" data-test="welcome-second-text">
    The easiest way to install ShellHub agent is with our automatic one-line
    installation script, which works with all Linux distributions that have
    Docker installed and properly set up.
  </p>
  <div class="mt-4 ml-4 mr-4">
    <p class="ml-2 pt-4 text-subtitle-2 text-bold" data-test="welcome-second-run-title">
      Run the following command on your device:
    </p>
    <CopyWarning :copied-item="'Tenant ID'">
      <template #default="{ copyText }">
        <v-text-field
          :model-value="command"
          @click:append="copyText(props.command)"
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

    <p class="text-caption mb-0 mt-1" data-test="welcome-second-link-docs">
      Check the
      <a
        :href="'https://docs.shellhub.io/user-guides/devices/adding'"
        target="_blank"
        rel="noopener noreferrer"
      >documentation</a
      >
      for more information and alternative install methods.
    </p>
  </div>
</template>

<script setup lang="ts">
import CopyWarning from "@/components/User/CopyWarning.vue";

const props = defineProps({
  command: {
    type: String,
    required: true,
  },
});

</script>
