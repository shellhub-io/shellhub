<template>
  <v-row class="mt-4 ml-2" v-if="!hasStatus">
    <v-col cols="12" md="4" class="pt-0" v-for="(item, index) in items" :key="index">
      <div>
        <StatCard
          :title="item.title"
          :content="item.content"
          :icon="item.icon"
          :buttonLabel="item.buttonLabel"
          :path="item.path"
          :stat="item.stat"
        />
      </div>
    </v-col>
  </v-row>
  <v-card class="mt-2 pa-4" v-else>
    <p class="text-center">Something is wrong, try again !</p>
  </v-card>
</template>

<script setup lang="ts">
import axios, { AxiosError } from "axios";
import { computed, onMounted, ref } from "vue";
import useStatsStore from "@admin/store/modules/stats";
import { StatCardItem } from "@/interfaces/IStats";
import useSnackbar from "@/helpers/snackbar";
import StatCard from "@/components/StatCard.vue";

const snackbar = useSnackbar();
const statsStore = useStatsStore();
const items = ref<StatCardItem[]>([]);
const hasStatus = ref(false);
const itemsStats = computed(() => statsStore.getStats);

onMounted(async () => {
  try {
    await statsStore.get();
    items.value = [
      {
        title: "Registered Users",
        content: "Registered users",
        icon: "mdi-account-group",
        buttonLabel: "View all Users",
        path: "users",
        stat: itemsStats.value.registered_users ?? 0,
      },
      {
        title: "Registered Devices",
        content: "Registered devices",
        icon: "mdi-developer-board",
        buttonLabel: "View all Devices",
        path: "devices",
        stat: itemsStats.value.registered_devices ?? 0,
      },
      {
        title: "Online Devices",
        content: "Devices are online and ready for connecting",
        icon: "mdi-developer-board",
        buttonLabel: "View all Devices",
        path: "devices",
        stat: itemsStats.value.online_devices ?? 0,
      },
      {
        title: "Active Sessions",
        content: "Active SSH Sessions opened by users",
        icon: "mdi-developer-board",
        buttonLabel: "View all Sessions",
        path: "sessions",
        stat: itemsStats.value.active_sessions ?? 0,
      },
      {
        title: "Pending Devices",
        content: "Pending devices",
        icon: "mdi-developer-board",
        buttonLabel: "View all Devices",
        path: "devices",
        stat: itemsStats.value.pending_devices ?? 0,
      },
      {
        title: "Rejected Devices",
        content: "Rejected devices",
        icon: "mdi-developer-board",
        buttonLabel: "View all Devices",
        path: "devices",
        stat: itemsStats.value.rejected_devices ?? 0,
      },
    ];
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

defineExpose({ items, itemsStats, hasStatus });
</script>
