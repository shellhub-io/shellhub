<template>
  <div v-if="!hasStatus">
    <PageHeader
      icon="mdi-home"
      title="System Overview"
      overline="Admin Dashboard"
      description="Monitor and manage your ShellHub instance metrics and statistics."
      icon-color="primary"
    />

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
          icon="mdi-history"
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
import axios from "axios";
import { computed, onMounted, ref } from "vue";
import useStatsStore from "@admin/store/modules/stats";
import useSnackbar from "@/helpers/snackbar";
import StatCard from "@/components/StatCard.vue";
import PageHeader from "@/components/PageHeader.vue";

const snackbar = useSnackbar();
const statsStore = useStatsStore();
const hasStatus = ref(false);
const stats = computed(() => statsStore.stats);

onMounted(async () => {
  try {
    await statsStore.getStats();
  } catch (error: unknown) {
    hasStatus.value = true;
    if (axios.isAxiosError(error) && error.response?.status === 402) {
      snackbar.showError("Failed to load the dashboard stats. Check your license and try again.");
    } else snackbar.showError("Failed to load the dashboard stats. Please try again.");
  }
});
</script>
