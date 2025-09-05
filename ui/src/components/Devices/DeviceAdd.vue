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

  <BaseDialog v-model="showDialog" transition="dialog-bottom-transition" data-test="device-add-dialog" threshold="md">
    <v-card class="bg-v-theme-surface">
      <v-btn
        icon
        variant="text"
        size="small"
        class="position-absolute"
        style="top: 8px; right: 8px; z-index: 1;"
        @click="showDialog = false"
        data-test="close-btn"
      >
        <v-icon>mdi-close</v-icon>
      </v-btn>

      <v-card-title class="text-h5 text-center mt-4" data-test="dialog-title">
        Adding a device
      </v-card-title>

      <v-card-text class="mt-2 mb-4" data-test="dialog-text">
        <p class="text-body-2 mb-4 text-center">
          Choose an installation method below. Expand each option to see requirements and get the specific command.
        </p>

        <v-expansion-panels 
          v-model="selectedPanel" 
          variant="accordion"
          elevation="0"
          class="bg-v-theme-surface"
        >
          <v-expansion-panel
            v-for="method in installMethods"
            :key="method.value"
            :value="method.value"
            class="bg-v-theme-surface"
          >
            <v-expansion-panel-title>
              <div class="d-flex align-center w-100">
                <v-icon :icon="method.icon" size="large" class="mr-3" :color="method.color"/>
                <div class="flex-grow-1">
                  <div class="text-h6">{{ method.name }}</div>
                  <div class="text-body-2 text-medium-emphasis">{{ method.description }}</div>
                </div>
                <v-chip 
                  v-if="method.recommended"
                  size="small" 
                  variant="tonal"
                  color="warning"
                  class="ml-2"
                >
                  <v-icon size="small" class="mr-1">mdi-star</v-icon>
                  recommended
                </v-chip>
              </div>
            </v-expansion-panel-title>

            <v-expansion-panel-text>
              <div class="pa-4">
                <h6 class="text-subtitle-2 mb-3">Requirements:</h6>
                <div class="requirements mb-4">
                  <div v-for="req in method.requirements" :key="req.text" class="d-flex align-center mb-2">
                    <v-icon 
                      size="small" 
                      :color="req.type === 'success' ? 'success' : req.type === 'warning' ? 'warning' : req.type === 'error' ? 'error' : 'info'" 
                      class="mr-2"
                    >
                      {{ req.type === 'success' ? 'mdi-check' : req.type === 'warning' ? 'mdi-alert' : req.type === 'error' ? 'mdi-alert-octagon' : 'mdi-information' }}
                    </v-icon>
                    <span class="text-body-2">{{ req.text }}</span>
                  </div>
                </div>

                <v-alert 
                  color="primary" 
                  variant="tonal" 
                  class="mb-3"
                  title="Installation"
                  icon="mdi-package-down"
                >
                  Ready to install? Copy the command below and run it on your target device:
                  <v-text-field
                    :model-value="getCommand(method.value)"
                    class="code mt-3"
                    variant="outlined"
                    readonly
                    density="compact"
                    hide-details
                  >
                    <template #append>
                      <v-btn
                        icon="mdi-content-copy"
                        color="primary"
                        variant="flat"
                        rounded
                        size="small"
                        @click="copyCommand(method.value)"
                      />
                    </template>
                  </v-text-field>
                </v-alert>
              </div>
            </v-expansion-panel-text>
          </v-expansion-panel>
        </v-expansion-panels>

        <v-divider class="my-4" />

        <p class="text-caption text-center mb-0">
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
const selectedPanel = ref(['auto']); // Auto panel expanded by default
const showCommand = ref<Record<string, boolean>>({});
const { tenantId } = authStore;

const installMethods = [
  {
    value: "auto",
    name: "Auto",
    icon: "mdi-auto-fix",
    color: "primary",
    recommended: true,
    description: "Automatically detects and uses the best available installation method",
    requirements: [
      { text: "Linux system", type: "success" },
      { text: "Internet connection", type: "success" },
      { text: "Tries Docker → Podman → Snap → Standalone", type: "info" }
    ],
  },
  {
    value: "docker",
    name: "Docker",
    icon: "mdi-docker",
    color: "blue",
    recommended: false,
    description: "Best performance and isolation using Docker containers",
    requirements: [
      { text: "Docker installed and running", type: "warning" },
      { text: "Docker accessible in rootful mode", type: "warning" },
      { text: "Linux/WSL environment", type: "success" }
    ],
  },
  {
    value: "podman",
    name: "Podman",
    icon: "mdi-cube-outline",
    color: "purple",
    recommended: false,
    description: "Alternative to Docker with rootless capabilities",
    requirements: [
      { text: "Podman installed and running", type: "warning" },
      { text: "Podman accessible in rootful mode", type: "warning" },
      { text: "Linux environment", type: "success" }
    ],
  },
  {
    value: "snap",
    name: "Snap",
    icon: "mdi-package-variant",
    color: "green",
    recommended: false,
    description: "Easy installation with automatic updates via Snap store",
    requirements: [
      { text: "Snap package manager installed", type: "warning" },
      { text: "Ubuntu/supported Linux distro", type: "success" },
      { text: "Automatic updates included", type: "success" }
    ],
  },
  {
    value: "standalone",
    name: "Standalone",
    icon: "mdi-server",
    color: "orange",
    recommended: false,
    description: "Direct installation using runc and systemd services",
    requirements: [
      { text: "Root/sudo privileges required", type: "error" },
      { text: "systemd-based Linux system", type: "warning" },
      { text: "Fallback when containers unavailable", type: "info" }
    ],
  },
  {
    value: "wsl",
    name: "WSL",
    icon: "mdi-microsoft-windows",
    color: "cyan",
    recommended: false,
    description: "Optimized for Windows Subsystem for Linux environments",
    requirements: [
      { text: "WSL2 environment", type: "warning" },
      { text: "systemd enabled in WSL", type: "warning" },
      { text: "Mirrored networking mode", type: "warning" }
    ],
  },
  {
    value: "yocto",
    name: "Yocto",
    icon: "mdi-chip",
    color: "teal",
    recommended: false,
    description: "For embedded systems built with Yocto Project framework",
    requirements: [
      { text: "Yocto-based embedded Linux", type: "warning" },
      { text: "Custom build integration required", type: "error" },
      { text: "Manual recipe configuration", type: "warning" }
    ],
  },
  {
    value: "buildroot",
    name: "Buildroot",
    icon: "mdi-memory",
    color: "indigo",
    recommended: false,
    description: "For embedded systems using Buildroot build system",
    requirements: [
      { text: "Buildroot-based system", type: "warning" },
      { text: "Custom package configuration", type: "error" },
      { text: "Manual build system integration", type: "warning" }
    ],
  },
  {
    value: "freebsd",
    name: "FreeBSD",
    icon: "mdi-freebsd",
    color: "red",
    recommended: false,
    description: "For FreeBSD systems using the official port",
    requirements: [
      { text: "FreeBSD operating system", type: "warning" },
      { text: "Ports collection available", type: "warning" },
      { text: "Manual port installation", type: "error" }
    ],
  },
];

const getCommand = (method: string) => {
  const port = window.location.port ? `:${window.location.port}` : "";
  const baseUrl = `${window.location.protocol}//${window.location.hostname}${port}`;

  return [
    "curl -sSf",
    `${baseUrl}/install.sh`,
    "|",
    !/^(auto)?$/.test(method) ? `INSTALL_METHOD=${method}` : "",
    `TENANT_ID=${tenantId}`,
    `SERVER_ADDRESS=${baseUrl}`,
    "sh",
  ].filter(Boolean).join(" ");
};

const copyAndShowCommand = async (methodValue: string) => {
  try {
    const command = getCommand(methodValue);
    await navigator.clipboard.writeText(command);
    showCommand.value[methodValue] = true;
    // Could add a toast notification here
  } catch (err) {
    console.error('Failed to copy: ', err);
  }
};

const copyCommand = async (methodValue: string) => {
  try {
    const command = getCommand(methodValue);
    await navigator.clipboard.writeText(command);
    // Could add a toast notification here
  } catch (err) {
    console.error('Failed to copy: ', err);
  }
};
</script>

<style lang="scss" scoped>
.code {
  font-family: monospace;
  font-size: 85%;
  font-weight: normal;
}

</style>
