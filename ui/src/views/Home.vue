<template>
  <v-row v-if="!hasStatus">
    <v-col cols="12" md="4" class="pt-0" v-for="(item, index) in items" :key="index">
      <div data-test="home-card">
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
  <v-card data-test="home-failed" class="mt-2 pa-4 bg-v-theme-surface" v-else>
    <p class="text-center">Something is wrong, try again !</p>
  </v-card>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from "vue";
import axios, { AxiosError } from "axios";
import { StatCardItem } from "@/interfaces/IStats";
import StatCard from "@/components/StatCard.vue";
import handleError from "@/utils/handleError";
import useSnackbar from "@/helpers/snackbar";
import useNamespacesStore from "@/store/modules/namespaces";
import useStatsStore from "@/store/modules/stats";

const namespacesStore = useNamespacesStore();
const statsStore = useStatsStore();
const snackbar = useSnackbar();
const hasStatus = ref(false);
const stats = computed(() => statsStore.stats);
const hasNamespace = computed(() => namespacesStore.namespaceList.length !== 0);

const items = computed<StatCardItem[]>(() => [
  {
    title: "Registered Devices",
    content: "Registered devices into the tenancy account",
    icon: "mdi-developer-board",
    buttonLabel: "Add Device",
    path: "devices",
    stat: stats.value.registered_devices || 0,
  },
  {
    title: "Online Devices",
    content: "Devices are online and ready for connecting",
    icon: "mdi-developer-board",
    buttonLabel: "View all Devices",
    path: "devices",
    stat: stats.value.online_devices || 0,
  },
  {
    title: "Active Sessions",
    content: "Active SSH Sessions opened by users",
    icon: "mdi-developer-board",
    buttonLabel: "View all Sessions",
    path: "sessions",
    stat: stats.value.active_sessions || 0,
  },
]);

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
