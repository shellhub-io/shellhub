<template>
  <v-row class="mt-4 ml-2" v-if="!hasStatus">
    <v-col cols="12" md="4" class="pt-0" v-for="item in items" :key="item.id">
      <div>
        <Card
          :id="item.id"
          :title="item.title"
          :fieldObject="item.fieldObject"
          :content="item.content"
          :icon="item.icon"
          :buttonName="item.buttonName"
          :pathName="item.pathName"
          :nameUseTest="item.nameUseTest"
          :stats="item.stats"
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
import useSnackbar from "@/helpers/snackbar";
import Card from "../components/Card.vue";

type ItemCard = {
  id: number;
  title: string;
  fieldObject: string;
  content: string;
  icon: string;
  buttonName: string;
  pathName: string;
  nameUseTest: string;
  stats: number;
};

const snackbar = useSnackbar();
const statsStore = useStatsStore();
const items = ref<ItemCard[]>([]);
const hasStatus = ref(false);
const itemsStats = computed(() => statsStore.getStats);

onMounted(async () => {
  try {
    await statsStore.get();
    items.value = [
      {
        id: 0,
        title: "Registered Users",
        fieldObject: "registered_users",
        content: "Registered users",
        icon: "mdi-account-group",
        stats: itemsStats.value.registered_users ?? 0,
        buttonName: "View all Users",
        pathName: "users",
        nameUseTest: "viewUsers-btn",
      },
      {
        id: 1,
        title: "Registered Devices",
        fieldObject: "registered_devices",
        content: "Registered devices",
        icon: "mdi-devices",
        stats: itemsStats.value.registered_devices ?? 0,
        buttonName: "View all Devices",
        pathName: "devices",
        nameUseTest: "viewRegisteredDevices-btn",
      },
      {
        id: 2,
        title: "Online Devices",
        fieldObject: "online_devices",
        content: "Devices are online and ready for connecting",
        icon: "mdi-devices",
        stats: itemsStats.value.online_devices ?? 0,
        buttonName: "View all Devices",
        pathName: "devices",
        nameUseTest: "viewOnlineDevices-btn",
      },
      {
        id: 3,
        title: "Active Sessions",
        fieldObject: "active_sessions",
        content: "Active SSH Sessions opened by users",
        icon: "mdi-devices",
        stats: itemsStats.value.active_sessions ?? 0,
        buttonName: "View all Sessions",
        pathName: "sessions",
        nameUseTest: "viewActiveSession-btn",
      },
      {
        id: 4,
        title: "Pending Devices",
        fieldObject: "pending_devices",
        content: "Pending devices",
        icon: "mdi-devices",
        stats: itemsStats.value.pending_devices ?? 0,
        buttonName: "View all Devices",
        pathName: "devices",
        nameUseTest: "viewPendingDevices-btn",
      },
      {
        id: 5,
        title: "Rejected Devices",
        fieldObject: "rejected_devices",
        content: "Rejected devices",
        icon: "mdi-devices",
        stats: itemsStats.value.rejected_devices ?? 0,
        buttonName: "View all Devices",
        pathName: "devices",
        nameUseTest: "viewRejectedDevices-btn",
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
