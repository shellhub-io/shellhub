<template>
  <div v-if="!hasStatus">
    <v-card
      class="bg-transparent mb-12"
      elevation="0"
    >
      <div class="d-flex align-start">
        <v-avatar
          color="primary"
          size="48"
          class="mr-4"
        >
          <v-icon
            size="32"
            icon="mdi-view-dashboard"
          />
        </v-avatar>
        <div>
          <h1 class="text-overline text-medium-emphasis mb-1">
            Admin Dashboard
          </h1>
          <h2 class="text-h5 font-weight-bold mb-2">
            System Overview
          </h2>
          <p class="text-body-2 text-medium-emphasis">
            Monitor and manage your ShellHub instance metrics and statistics
          </p>
        </div>
      </div>
    </v-card>

    <v-row class="d-flex align-center mb-2 pa-3">
      <v-icon
        class="mr-2"
        icon="mdi-chart-box-outline"
      />
      <h2 class="text-h6">
        Stats
      </h2>
    </v-row>
    <v-row>
      <v-col
        cols="12"
        md="4"
      >
        <StatCard
          title="Registered Users"
          :stat="stats.registered_users ?? 0"
          icon="mdi-account-group"
          button-label="View all Users"
          path="users"
        />
      </v-col>

      <v-col
        cols="12"
        md="4"
      >
        <StatCard
          title="Registered Devices"
          :stat="stats.registered_devices ?? 0"
          icon="mdi-developer-board"
          button-label="View all Devices"
          path="devices"
        />
      </v-col>

      <v-col
        cols="12"
        md="4"
      >
        <StatCard
          title="Online Devices"
          :stat="stats.online_devices ?? 0"
          icon="mdi-lan-connect"
          button-label="View Online Devices"
          path="devices"
        />
      </v-col>

      <v-col
        cols="12"
        md="4"
      >
        <StatCard
          title="Pending Devices"
          :stat="stats.pending_devices ?? 0"
          icon="mdi-clock-outline"
          button-label="View Pending Devices"
          path="devices"
        />
      </v-col>

      <v-col
        cols="12"
        md="4"
      >
        <StatCard
          title="Rejected Devices"
          :stat="stats.rejected_devices ?? 0"
          icon="mdi-close-circle"
          button-label="View Rejected Devices"
          path="devices"
        />
      </v-col>

      <v-col
        cols="12"
        md="4"
      >
        <StatCard
          title="Active Sessions"
          :stat="stats.active_sessions ?? 0"
          icon="mdi-console-network"
          button-label="View all Sessions"
          path="sessions"
        />
      </v-col>
    </v-row>
  </div>

  <v-card
    v-else
    data-test="dashboard-failed"
    class="mt-2 pa-4 bg-v-theme-surface"
  >
    <p class="text-center">
      Something is wrong, try again!
    </p>
  </v-card>
</template>

<script setup lang="ts">
import axios, { AxiosError } from "axios";
import { onMounted, ref } from "vue";
import useStatsStore from "@admin/store/modules/stats";
import { IAdminStats } from "@admin/interfaces/IStats";
import useSnackbar from "@/helpers/snackbar";
import StatCard from "@/components/StatCard.vue";

const snackbar = useSnackbar();
const statsStore = useStatsStore();
const hasStatus = ref(false);
const stats = ref({} as IAdminStats);

onMounted(async () => {
  try {
    stats.value = await statsStore.getStats();
  } catch (error: unknown) {
    hasStatus.value = true;
    if (axios.isAxiosError(error)) {
      const axiosError = error as AxiosError;
      if (axiosError.response?.status === 402) {
        snackbar.showError("Failed to load the dashboard stats. Check your license and try again.");
      } else {
        snackbar.showError("Failed to load the dashboard stats. Please try again.");
      }
    }
  }
});

defineExpose({ stats, hasStatus });
</script>
