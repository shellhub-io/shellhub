<template>
  <v-row class="mt-2 ml-2" v-if="!hasStatus">
    <v-col cols="12" md="4" class="pt-0" v-for="item in items" :key="item.id">
      <div data-test="dashboard-card">
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
  <v-card data-test="dashboard-failed" class="mt-2 pa-4 bg-v-theme-surface" v-else>
    <p class="text-center">Something is wrong, try again !</p>
  </v-card>
</template>

<script lang="ts">
import { computed, defineComponent, onMounted, ref } from "vue";
import axios, { AxiosError } from "axios";
import { INotificationsError } from "../interfaces/INotifications";
import Card from "../components/Card/Card.vue";
import { useStore } from "../store";
import handleError from "@/utils/handleError";

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

export default defineComponent({
  name: "DashboardView",
  components: { Card },
  setup() {
    const store = useStore();
    const hasStatus = ref(false);
    const itemsStats = computed(() => store.getters["stats/stats"]);
    const hasNamespace = computed(
      () => store.getters["namespaces/getNumberNamespaces"] !== 0,
    );
    const items = computed(() => [
      {
        id: 1,
        title: "Registered Devices",
        fieldObject: "registered_devices",
        content: "Registered devices into the tenancy account",
        icon: "mdi-devices",
        stats: itemsStats.value.registered_devices || 0,
        buttonName: "Add Device",
        pathName: "devices",
        nameUseTest: "registeredDevices-btn",
      },
      {
        id: 2,
        title: "Online Devices",
        fieldObject: "online_devices",
        content: "Devices are online and ready for connecting",
        icon: "mdi-devices",
        stats: itemsStats.value.online_devices || 0,
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
        stats: itemsStats.value.active_sessions || 0,
        buttonName: "View all Sessions",
        pathName: "sessions",
        nameUseTest: "viewActiveSession-btn",
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
              store.dispatch(
                "snackbar/showSnackbarErrorAction",
                INotificationsError.dashboard,
              );
              break;
            }
          }
        }
        handleError(error);
      }
    });

    return {
      hasStatus,
      itemsStats,
      items,
    };
  },
});
</script>
