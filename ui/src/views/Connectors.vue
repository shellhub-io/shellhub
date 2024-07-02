<template>
  <div
    class="d-flex flex-column justify-space-between align-center flex-sm-row"
    data-test="device-title"
  >
    <h1>Connectors</h1>

    <v-col md="6" sm="12" />

    <div class="d-flex mt-4 mb-2" data-test="device-header-component-group">
      <v-btn
        @click="router.push('/containers')"
        class="mr-2"
        color="primary"
        tabindex="0"
        variant="text"
        aria-label="Return to Containers List"
        data-test="connector-add-btn"
      >
        Containers
      </v-btn>
      <ConnectorAdd />
    </div>
  </div>
  <v-card class="mt-2" data-test="device-table-component">
    <ConnectorList />
  </v-card>

  <BoxMessage
    v-if="showMessageBox"
    class="mt-2"
    type-message="device"
    data-test="boxMessageDevice-component"
  />
</template>

<script setup lang="ts">
import { onMounted, computed, onUnmounted } from "vue";
import { useRouter } from "vue-router";
import axios, { AxiosError } from "axios";
import { useStore } from "../store";
import BoxMessage from "../components/Box/BoxMessage.vue";
import handleError from "@/utils/handleError";
import ConnectorList from "../components/Connector/ConnectorList.vue";
import ConnectorAdd from "../components/Connector/ConnectorAdd.vue";

const router = useRouter();
const store = useStore();

const hasContainer = computed(() => (
  store.getters["stats/stats"].registered_devices > 0
        || store.getters["stats/stats"].pending_devices > 0
        || store.getters["stats/stats"].rejected_devices > 0
));

const showMessageBox = computed(() => !hasContainer.value);

onMounted(async () => {
  try {
    await store.dispatch("stats/get");
  } catch (error: unknown) {
    if (axios.isAxiosError(error)) {
      const axiosError = error as AxiosError;
      if (axiosError.response?.status === 403) store.dispatch("snackbar/showSnackbarErrorAssociation");
    } else {
      store.dispatch("snackbar/showSnackbarErrorDefault");
    }
    handleError(error);
  }
});

onUnmounted(async () => {
  await store.dispatch("devices/setFilter", "");
});

</script>
