<template>
  <PageHeader
    icon="mdi-connection"
    title="Docker Connectors"
    overline="Docker Integration"
    description="Connect to Docker Engine instances to view and manage containers.
      Add connectors to enable ShellHub to access your Docker environments."
    icon-color="primary"
    data-test="device-title"
  >
    <template #actions>
      <v-btn
        class="mr-2"
        color="primary"
        tabindex="0"
        variant="text"
        aria-label="Return to Containers List"
        data-test="connector-add-btn"
        @click="router.push('/containers')"
      >
        Containers
      </v-btn>
      <ConnectorAdd @update="getConnectors()" />
    </template>
  </PageHeader>
  <div
    class="mt-2"
    data-test="connector-table-component"
  >
    <ConnectorList />
  </div>
</template>

<script setup lang="ts">
import { useRouter } from "vue-router";
import axios from "axios";
import handleError from "@/utils/handleError";
import ConnectorList from "../components/Connector/ConnectorList.vue";
import ConnectorAdd from "../components/Connector/ConnectorAdd.vue";
import PageHeader from "../components/PageHeader.vue";
import useSnackbar from "@/helpers/snackbar";
import useConnectorStore from "@/store/modules/connectors";

const router = useRouter();
const connectorStore = useConnectorStore();
const snackbar = useSnackbar();

const getConnectors = async () => {
  try {
    await connectorStore.fetchConnectorList({
      page: 1,
      perPage: 10,
    });
  } catch (error: unknown) {
    if (axios.isAxiosError(error) && error.response?.status === 403) {
      snackbar.showError("You do not have permission to access the connectors.");
    } else snackbar.showError("Error loading the connectors.");

    handleError(error);
  }
};

</script>
