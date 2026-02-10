<template>
  <v-btn
    color="primary"
    tabindex="0"
    variant="elevated"
    data-test="device-add-btn"
    :size
    text="Add Device"
    @click="showDialog = true"
    @keypress.enter="showDialog = true"
  />

  <WindowDialog
    v-model="showDialog"
    transition="dialog-bottom-transition"
    data-test="device-add-dialog"
    threshold="md"
    title="Adding a device"
    description="Choose an installation method and get your device connected"
    icon="mdi-developer-board"
    icon-color="primary"
    @close="showDialog = false"
  >
    <v-card-text
      class="mt-2 mb-4"
      data-test="dialog-text"
    >
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
              <v-icon
                :icon="method.icon"
                size="large"
                class="mr-3"
                :color="method.color"
              />
              <div class="flex-grow-1">
                <div class="text-h6">
                  {{ method.name }}
                </div>
                <div class="text-body-2 text-medium-emphasis">
                  {{ method.description }}
                </div>
              </div>
              <v-chip
                v-if="method.recommended"
                size="small"
                variant="tonal"
                color="success"
                class="ml-2"
              >
                <v-icon
                  size="small"
                  class="mr-1"
                >
                  mdi-star
                </v-icon>
                recommended
              </v-chip>
            </div>
          </v-expansion-panel-title>

          <v-expansion-panel-text>
            <div class="pa-4">
              <h6 class="text-subtitle-2 mb-3">
                Requirements:
              </h6>
              <div class="requirements mb-4">
                <div
                  v-for="req in method.requirements"
                  :key="req.text"
                  class="d-flex align-center mb-2"
                >
                  <v-icon
                    size="small"
                    color="success"
                    class="mr-2"
                  >
                    mdi-check
                  </v-icon>
                  <span class="text-body-2">{{ req.text }}</span>
                </div>
              </div>

              <!-- Script-based installation -->
              <v-alert
                v-if="!isManualInstall(method.value)"
                color="primary"
                variant="tonal"
                class="mb-3"
                title="Installation"
                icon="mdi-package-down"
                role="status"
                aria-live="polite"
              >
                Ready to install? Copy the command below and run it on your target device:
                <CopyCommandField
                  :command="getCommand(method.value)"
                  class="mt-3"
                />

                <!-- Advanced Options inside the alert -->
                <v-expansion-panels
                  v-model="methodAdvancedPanels[method.value]"
                  variant="accordion"
                  elevation="0"
                  class="mt-4"
                >
                  <v-expansion-panel
                    value="advanced"
                    bg-color="transparent"
                  >
                    <v-expansion-panel-title class="py-2">
                      <div class="d-flex align-center w-100">
                        <v-icon
                          icon="mdi-tune"
                          size="small"
                          class="mr-2"
                        />
                        <div class="flex-grow-1">
                          <div class="text-subtitle-2">
                            Advanced Options
                          </div>
                          <div class="text-caption">
                            Configure additional environment variables
                          </div>
                        </div>
                      </div>
                    </v-expansion-panel-title>

                    <v-expansion-panel-text>
                      <div class="pa-2">
                        <v-row>
                          <v-col
                            cols="12"
                            sm="6"
                          >
                            <v-text-field
                              v-model="advancedOptions.preferredHostname"
                              label="Preferred Hostname"
                              placeholder="e.g., my-device"
                              variant="outlined"
                              density="compact"
                              hint="Override device hostname"
                              persistent-hint
                            />
                          </v-col>
                          <v-col
                            cols="12"
                            sm="6"
                          >
                            <v-text-field
                              v-model="advancedOptions.preferredIdentity"
                              label="Preferred Identity"
                              placeholder="e.g., server-01"
                              variant="outlined"
                              density="compact"
                              hint="Override device identity"
                              persistent-hint
                            />
                          </v-col>
                        </v-row>
                      </div>
                    </v-expansion-panel-text>
                  </v-expansion-panel>
                </v-expansion-panels>
              </v-alert>

              <!-- Manual installation -->
              <v-alert
                v-else
                color="warning"
                variant="tonal"
                class="mb-3"
                title="Manual Installation Required"
                icon="mdi-book-open-variant"
                role="alert"
                aria-live="assertive"
              >
                This method requires manual configuration and build system integration. Please follow the detailed documentation:
                <div class="mt-3">
                  <v-btn
                    :href="getDocumentationUrl(method.value)"
                    target="_blank"
                    rel="noopener noreferrer"
                    color="warning"
                    variant="outlined"
                    prepend-icon="mdi-open-in-new"
                    size="small"
                    text="View Documentation"
                  />
                </div>
              </v-alert>
            </div>
          </v-expansion-panel-text>
        </v-expansion-panel>
      </v-expansion-panels>
    </v-card-text>

    <!-- Footer -->
    <template #footer>
      <v-spacer />
      <div class="text-caption text-medium-emphasis text-center">
        Check the
        <a
          :href="'https://docs.shellhub.io/user-guides/devices/adding'"
          target="_blank"
          rel="noopener noreferrer"
          data-test="documentation-link"
          class="text-primary text-decoration-none"
        >documentation</a>
        for more information and alternative install methods.
      </div>
      <v-spacer />
    </template>
  </WindowDialog>
</template>

<script setup lang="ts">
import { ref } from "vue";
import WindowDialog from "@/components/Dialogs/WindowDialog.vue";
import CopyCommandField from "@/components/CopyCommandField.vue";
import useAuthStore from "@/store/modules/auth";

enum InstallMethod {
  AUTO = "auto",
  DOCKER = "docker",
  PODMAN = "podman",
  SNAP = "snap",
  STANDALONE = "standalone",
  WSL = "wsl",
  YOCTO = "yocto",
  BUILDROOT = "buildroot",
  FREEBSD = "freebsd",
}

interface InstallRequirement {
  text: string;
}

interface InstallMethodConfig {
  value: InstallMethod;
  name: string;
  icon: string;
  color: string;
  recommended: boolean;
  description: string;
  requirements: InstallRequirement[];
}

const { size } = defineProps<{ size?: string }>();
const authStore = useAuthStore();
const showDialog = ref(false);
const selectedPanel = ref(["auto"]); // Auto panel expanded by default
const methodAdvancedPanels = ref<Record<string, string[]>>({});
const { tenantId } = authStore;

const MANUAL_INSTALL_METHODS = [InstallMethod.YOCTO, InstallMethod.BUILDROOT, InstallMethod.FREEBSD];

const DOCUMENTATION_URLS: Partial<Record<InstallMethod, string>> = {
  [InstallMethod.YOCTO]: "https://docs.shellhub.io/overview/supported-platforms/yocto",
  [InstallMethod.BUILDROOT]: "https://docs.shellhub.io/overview/supported-platforms/buildroot",
  [InstallMethod.FREEBSD]: "https://docs.shellhub.io/overview/supported-platforms/freebsd",
};

const advancedOptions = ref({
  preferredHostname: "",
  preferredIdentity: "",
});

const installMethods: InstallMethodConfig[] = [
  {
    value: InstallMethod.AUTO,
    name: "Auto",
    icon: "mdi-auto-fix",
    color: "primary",
    recommended: true,
    description: "Automatically detects and uses the best available installation method",
    requirements: [
      { text: "Linux system with curl" },
      { text: "Internet connection" },
      { text: "Tries: Docker → Podman → Snap → Standalone" },
    ],
  },
  {
    value: InstallMethod.DOCKER,
    name: "Docker",
    icon: "mdi-docker",
    color: "blue",
    recommended: false,
    description: "Best performance and isolation using Docker containers",
    requirements: [
      { text: "Docker daemon running" },
      { text: "Access to /var/run/docker.sock" },
      { text: "Sufficient privileges (root/sudo)" },
    ],
  },
  {
    value: InstallMethod.PODMAN,
    name: "Podman",
    icon: "mdi-cube-outline",
    color: "purple",
    recommended: false,
    description: "Alternative to Docker with rootless capabilities",
    requirements: [
      { text: "Podman daemon running" },
      { text: "Access to /var/run/podman/podman.sock" },
      { text: "Sufficient privileges (root/sudo)" },
    ],
  },
  {
    value: InstallMethod.SNAP,
    name: "Snap",
    icon: "mdi-package-variant",
    color: "green",
    recommended: false,
    description: "Easy installation with automatic updates via Snap store",
    requirements: [
      { text: "Snap package manager installed" },
      { text: "Snapd service running" },
      { text: "Network access to Snap Store" },
    ],
  },
  {
    value: InstallMethod.STANDALONE,
    name: "Standalone",
    icon: "mdi-server",
    color: "orange",
    recommended: false,
    description: "Direct installation using runc and systemd services",
    requirements: [
      { text: "systemd-based Linux system" },
      { text: "Root/sudo privileges required" },
      { text: "Used when containers unavailable" },
    ],
  },
  {
    value: InstallMethod.WSL,
    name: "WSL",
    icon: "mdi-microsoft-windows",
    color: "cyan",
    recommended: false,
    description: "Optimized for Windows Subsystem for Linux environments",
    requirements: [
      { text: "WSL2 with systemd enabled" },
      { text: "Mirrored networking mode" },
      { text: "Root/sudo privileges required" },
    ],
  },
  {
    value: InstallMethod.YOCTO,
    name: "Yocto Project",
    icon: "mdi-chip",
    color: "teal",
    recommended: false,
    description: "For embedded systems built with Yocto Project framework",
    requirements: [
      { text: "Yocto Project build environment" },
      { text: "Custom layer and recipe creation" },
      { text: "Manual integration required" },
    ],
  },
  {
    value: InstallMethod.BUILDROOT,
    name: "Buildroot",
    icon: "mdi-memory",
    color: "indigo",
    recommended: false,
    description: "For embedded systems using Buildroot build system",
    requirements: [
      { text: "Buildroot build environment" },
      { text: "Custom package configuration" },
      { text: "Manual integration required" },
    ],
  },
  {
    value: InstallMethod.FREEBSD,
    name: "FreeBSD",
    icon: "mdi-freebsd",
    color: "red",
    recommended: false,
    description: "For FreeBSD systems using the official port",
    requirements: [
      { text: "FreeBSD operating system" },
      { text: "Ports tree installed" },
      { text: "Manual compilation and setup" },
    ],
  },
];

const isManualInstall = (method: string) => MANUAL_INSTALL_METHODS.includes(method as InstallMethod);

const getDocumentationUrl = (method: string) => DOCUMENTATION_URLS[method as InstallMethod] || DOCUMENTATION_URLS[InstallMethod.AUTO];

const getCommand = (method: InstallMethod) => {
  const { origin } = window.location;

  const envVars = [
    method !== InstallMethod.AUTO ? `INSTALL_METHOD=${method}` : "",
    `TENANT_ID=${tenantId}`,
    `SERVER_ADDRESS=${origin}`,
    advancedOptions.value.preferredHostname ? `PREFERRED_HOSTNAME="${advancedOptions.value.preferredHostname}"` : "",
    advancedOptions.value.preferredIdentity ? `PREFERRED_IDENTITY="${advancedOptions.value.preferredIdentity}"` : "",
  ].filter(Boolean);

  return [
    "curl -sSf",
    `${origin}/install.sh`,
    "|",
    ...envVars,
    "sh",
  ].join(" ");
};
</script>
