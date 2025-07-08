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
import StatCard from "@/components/StatCard.vue";
import { useStore } from "../store";
import handleError from "@/utils/handleError";
import useSnackbar from "@/helpers/snackbar";

type ItemCard = {
  title: string;
  content: string;
  icon: string;
  buttonLabel: string;
  path: string;
  stat: number;
};

const store = useStore();
const snackbar = useSnackbar();
const hasStatus = ref(false);
const itemsStats = computed(() => store.getters["stats/stats"]);
const hasNamespace = computed(
  () => store.getters["namespaces/getNumberNamespaces"] !== 0,
);
const items = computed(() => [
  {
    title: "Registered Devices",
    content: "Registered devices into the tenancy account",
    icon: "mdi-devices",
    buttonLabel: "Add Device",
    path: "devices",
    stat: itemsStats.value.registered_devices || 0,
  },
  {
    title: "Online Devices",
    content: "Devices are online and ready for connecting",
    icon: "mdi-devices",
    buttonLabel: "View all Devices",
    path: "devices",
    stat: itemsStats.value.online_devices || 0,
  },
  {
    title: "Active Sessions",
    content: "Active SSH Sessions opened by users",
    icon: "mdi-devices",
    buttonLabel: "View all Sessions",
    path: "sessions",
    stat: itemsStats.value.active_sessions || 0,
  },
] as ItemCard[]);

onMounted(async () => {
  if (!hasNamespace.value) return;

  try {
    await store.dispatch("stats/get");
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
