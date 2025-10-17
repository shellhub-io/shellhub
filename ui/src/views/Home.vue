<template>
  <div v-if="!hasStatus">
    <!-- Namespace Info Card -->
    <v-row>
      <v-col cols="12">
        <v-card class="bg-transparent mb-6" elevation="0" rounded="0">
          <v-row>
            <v-col cols="12" md="6">
              <div class="d-flex align-start">
                <v-avatar color="primary" size="48" class="mr-4">
                  <v-icon size="32">mdi-home</v-icon>
                </v-avatar>
                <div>
                  <div class="text-overline text-medium-emphasis mb-1">Home</div>
                  <div class="text-h5 font-weight-bold mb-2">{{ namespace.name }}</div>
                  <div class="text-body-2 text-medium-emphasis">
                    This is your active namespace. All devices, sessions and configurations are isolated within this namespace.
                  </div>
                </div>
              </div>
            </v-col>
            <v-col cols="12" md="6">
              <v-card class="pa-4" variant="tonal">
                <div class="text-overline text-medium-emphasis mb-2">TENANT ID</div>
                <div class="d-flex align-center justify-space-between">
                  <code class="text-primary">{{ namespace.tenant_id }}</code>
                  <CopyWarning :copied-item="'Tenant ID'">
                    <template #default="{ copyText }">
                      <v-btn
                        @click="copyText(namespace.tenant_id)"
                        color="primary"
                        variant="elevated"
                        size="small"
                        prepend-icon="mdi-content-copy"
                      >
                        Copy
                      </v-btn>
                    </template>
                  </CopyWarning>
                </div>
                <div class="text-caption text-medium-emphasis mt-2">
                  Use this ID to register new devices to this namespace
                </div>
              </v-card>
            </v-col>
          </v-row>
        </v-card>
      </v-col>
    </v-row>

    <!-- Devices Section -->
    <v-row>
      <v-col cols="12" class="d-flex align-center mb-2">
        <v-icon class="mr-2">mdi-devices</v-icon>
        <h2 class="text-h6">Devices</h2>
      </v-col>
    </v-row>

    <v-row>
      <!-- Accepted Devices Card -->
      <v-col cols="12" md="3">
        <v-card class="pa-6 bg-v-theme-surface text-center h-100" border>
          <v-avatar color="primary" size="64" class="mb-4">
            <v-icon size="40">mdi-check</v-icon>
          </v-avatar>

          <div class="text-overline text-medium-emphasis mb-2">ACCEPTED DEVICES</div>
          <div class="text-h2 font-weight-bold mb-4">{{ stats.registered_devices || 0 }}</div>

          <v-btn
            variant="text"
            color="primary"
            size="small"
            :to="'/devices'"
            block
          >
            View all devices
          </v-btn>
        </v-card>
      </v-col>

      <!-- Online Devices Card -->
      <v-col cols="12" md="3">
        <v-card class="pa-6 bg-v-theme-surface text-center h-100" border>
          <v-avatar color="primary" size="64" class="mb-4">
            <v-icon size="40">mdi-lan-connect</v-icon>
          </v-avatar>

          <div class="text-overline text-medium-emphasis mb-2">ONLINE DEVICES</div>
          <div class="text-h2 font-weight-bold mb-4">{{ stats.online_devices || 0 }}</div>

          <v-btn
            variant="text"
            color="primary"
            size="small"
            :to="'/devices'"
            block
          >
            View Online Devices
          </v-btn>
        </v-card>
      </v-col>

      <!-- Pending Devices Card -->
      <v-col cols="12" md="3">
        <v-card class="pa-6 bg-v-theme-surface text-center h-100" border>
          <v-avatar color="primary" size="64" class="mb-4">
            <v-icon size="40">mdi-clock-outline</v-icon>
          </v-avatar>

          <div class="text-overline text-medium-emphasis mb-2">PENDING DEVICES</div>
          <div class="text-h2 font-weight-bold mb-4">{{ stats.pending_devices || 0 }}</div>

          <v-btn
            variant="text"
            color="primary"
            size="small"
            :to="'/devices/pending'"
            block
          >
            Approve Devices
          </v-btn>
        </v-card>
      </v-col>

      <!-- Connect Device Card -->
      <v-col cols="12" md="3">
        <v-card class="pa-6 bg-transparent text-center h-100 border border-dashed">
          <v-avatar color="surface-variant" size="64" class="mb-4" theme="dark">
            <v-icon size="40" color="primary">mdi-developer-board</v-icon>
          </v-avatar>

          <div class="text-h6 font-weight-bold mb-2">Add a new device</div>
          <div class="text-body-2 text-medium-emphasis mb-4">
            Register new devices to this namespace and start managing remote connections
          </div>

          <DeviceAdd />
        </v-card>
      </v-col>
    </v-row>

  </div>
  <v-card data-test="home-failed" class="mt-2 pa-4 bg-v-theme-surface" v-else>
    <p class="text-center">Something is wrong, try again !</p>
  </v-card>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from "vue";
import axios, { AxiosError } from "axios";
import handleError from "@/utils/handleError";
import useSnackbar from "@/helpers/snackbar";
import useNamespacesStore from "@/store/modules/namespaces";
import useStatsStore from "@/store/modules/stats";
import DeviceAdd from "@/components/Devices/DeviceAdd.vue";
import CopyWarning from "@/components/User/CopyWarning.vue";

const namespacesStore = useNamespacesStore();
const statsStore = useStatsStore();
const snackbar = useSnackbar();
const hasStatus = ref(false);
const stats = computed(() => statsStore.stats);
const namespace = computed(() => namespacesStore.currentNamespace);
const hasNamespace = computed(() => namespacesStore.namespaceList.length !== 0);

onMounted(async () => {
  if (!hasNamespace.value) return;

  try {
    await statsStore.fetchStats();
  } catch (error: unknown) {
    if (axios.isAxiosError(error)) {
      const axiosError = error as AxiosError;
      switch (true) {
        case axiosError.response && axiosError.response?.status === 403: {
          hasStatus.value = true;
          break;
        }
        default: {
          hasStatus.value = true;
          snackbar.showError("Failed to load the home page.");
          break;
        }
      }
    }
    handleError(error);
  }
});

defineExpose({ hasStatus });
</script>
