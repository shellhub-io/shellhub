<template>
  <div
    class="d-flex flex-column justify-space-between align-center flex-sm-row"
    data-test="device-title"
  >
    <h1>Docker Connectors</h1>

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
      <ConnectorAdd @update="getConnectors()" />
    </div>
  </div>
  <v-card class="mt-2" data-test="device-table-component">
    <ConnectorList />
  </v-card>
</template>

<script setup lang="ts">
import { onMounted } from "vue";
import { useRouter } from "vue-router";
import axios, { AxiosError } from "axios";
import { useStore } from "../store";
import handleError from "@/utils/handleError";
import ConnectorList from "../components/Connector/ConnectorList.vue";
import ConnectorAdd from "../components/Connector/ConnectorAdd.vue";
import { INotificationsError } from "@/interfaces/INotifications";

const router = useRouter();
const store = useStore();

const getConnectors = async () => {
  try {
    await store.dispatch("connectors/fetch", {
      page: store.getters["connectors/getPage"],
      perPage: store.getters["connectors/getPerPage"],
    });
  } catch (error: unknown) {
    if (axios.isAxiosError(error)) {
      const axiosError = error as AxiosError;
      if (axiosError.response?.status === 403) {
        store.dispatch("snackbar/showSnackbarErrorAssociation");
        handleError(error);
      }
    } else {
      store.dispatch(
        "snackbar/showSnackbarErrorAction",
        INotificationsError.connectorLoad,
      );
      handleError(error);
    }
  }
};

onMounted(async () => {
  getConnectors();
});

</script>
