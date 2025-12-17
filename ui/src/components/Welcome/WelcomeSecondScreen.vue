<template>
  <div class="pa-6">
    <div class="text-center mb-6">
      <v-avatar
        size="64"
        color="primary"
        class="mb-4"
      >
        <v-icon
          size="32"
          color="white"
          icon="mdi-download"
        />
      </v-avatar>
      <h2 class="text-h4 mb-2">
        Install ShellHub Agent
      </h2>
      <p class="text-subtitle-1 text-medium-emphasis">
        Connect your device to ShellHub in just one step
      </p>
    </div>

    <v-card
      variant="outlined"
      class="pa-4 mb-4"
      color="primary"
    >
      <div class="mb-1 text-high-emphasis">
        <h3 class="text-h6">Requirements:</h3>
        <v-list
          class="bg-transparent"
          :lines="false"
        >
          <v-list-item
            v-for="requirement in requirements"
            :key="requirement"
            class="d-flex align-center pa-0"
            density="compact"
          >
            <v-icon
              size="small"
              color="success"
              class="mr-2"
              icon="mdi-check"
            />
            <span class="text-body-2">{{ requirement }}</span>
          </v-list-item>
        </v-list>
      </div>

      <v-alert
        color="primary"
        variant="tonal"
        title="Installation"
        icon="mdi-package-down"
      >
        <div data-test="welcome-second-run-title">
          Ready to install? Copy the command below and run it on your target device:
        </div>
        <CopyCommandField
          :command="command"
          class="mt-3"
        />
      </v-alert>
    </v-card>

    <v-card
      variant="tonal"
      color="warning"
      class="pa-3"
    >
      <div class="d-flex align-center h-100">
        <v-icon
          size="24"
          color="warning"
          class="mr-3"
          icon="mdi-clock-outline"
        />
        <div>
          <v-card-title class="text-h6 mb-1 pa-0">
            Waiting for device...
          </v-card-title>
          <v-card-text class="mb-0 pa-0">
            After running the command, your device will appear in the next step for approval.
          </v-card-text>
        </div>
      </div>
    </v-card>
  </div>
</template>

<script setup lang="ts">
import CopyCommandField from "@/components/CopyCommandField.vue";
import useAuthStore from "@/store/modules/auth";

const { tenantId } = useAuthStore();
const requirements = ["Linux system with curl", "Internet connection", "Tries: Docker → Podman → Snap → Standalone"];
const { origin } = window.location;
const command = `curl -sSf ${origin}/install.sh | TENANT_ID=${tenantId} SERVER_ADDRESS=${origin} sh`;
</script>
