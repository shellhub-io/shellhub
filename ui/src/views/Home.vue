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
                  <code class="text-primary" data-test="tenant-info-text">{{ namespace.tenant_id }}</code>
                  <CopyWarning :copied-item="'Tenant ID'">
                    <template #default="{ copyText }">
                      <v-btn
                        data-test="copy-tenant-btn"
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
      <v-col cols="12" md="3">
        <StatCard
          title="Accepted Devices"
          :stat="stats.registered_devices || 0"
          icon="mdi-check"
          button-label="View all devices"
          path="/devices"
        />
      </v-col>

      <v-col cols="12" md="3">
        <StatCard
          title="Online Devices"
          :stat="stats.online_devices || 0"
          icon="mdi-lan-connect"
          button-label="View Online Devices"
          path="/devices"
        />
      </v-col>

      <v-col cols="12" md="3">
        <StatCard
          title="Pending Devices"
          :stat="stats.pending_devices || 0"
          icon="mdi-clock-outline"
          button-label="Approve Devices"
          path="/devices/pending"
        />
      </v-col>

      <v-col cols="12" md="3">
        <v-card class="pa-6 bg-transparent text-center h-100 border border-dashed">
          <v-avatar color="surface-variant" size="64" class="mb-4" theme="dark">
            <v-icon size="40" color="primary" icon="mdi-developer-board" />
          </v-avatar>
          <v-card-title class="text-h6 font-weight-bold mb-2">Add a new device</v-card-title>
          <v-card-subtitle class="text-body-2 text-medium-emphasis mb-4 text-wrap">
            Register new devices to this namespace and start managing remote connections
          </v-card-subtitle>
          <DeviceAdd />
        </v-card>
      </v-col>
    </v-row>

  </div>
  <v-card data-test="home-failed" class="mt-2 pa-4 bg-v-theme-surface" v-else>
    <p class="text-center">Something is wrong, try again!</p>
  </v-card>
</template>

<script setup lang="ts">
import { computed, onMounted, ref, watch } from "vue";
import axios, { AxiosError } from "axios";
import handleError from "@/utils/handleError";
import useSnackbar from "@/helpers/snackbar";
import useNamespacesStore from "@/store/modules/namespaces";
import useStatsStore from "@/store/modules/stats";
import DeviceAdd from "@/components/Devices/DeviceAdd.vue";
import CopyWarning from "@/components/User/CopyWarning.vue";
import StatCard from "@/components/StatCard.vue";

const namespacesStore = useNamespacesStore();
const statsStore = useStatsStore();
const snackbar = useSnackbar();
const hasStatus = ref(false);
const stats = computed(() => statsStore.stats);
const namespace = computed(() => namespacesStore.currentNamespace);
const hasNamespace = computed(() => namespacesStore.namespaceList.length !== 0);

const fetchStats = async () => {
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
};

onMounted(async () => {
  await fetchStats();
});

watch(hasNamespace, (newValue) => {
  if (newValue) {
    fetchStats();
  }
});

defineExpose({ hasStatus });
</script>
