<template>
  <PageHeader
    icon="mdi-web"
    title="Web Endpoints"
    overline="Web Access"
    :description="webEndpointDescription"
    icon-color="primary"
  >
    <template #actions>
      <v-btn
        color="primary"
        variant="elevated"
        data-test="tunnel-create-dialog-btn"
        :disabled="!canCreateWebEndpoint"
        @click="showWebEndpointCreate = true"
        @keypress.enter="showWebEndpointCreate = true"
      >
        Create Web Endpoint
      </v-btn>
    </template>
  </PageHeader>

  <v-text-field
    v-if="showList"
    v-model.trim="filter"
    label="Search by Address"
    variant="outlined"
    color="primary"
    single-line
    hide-details
    prepend-inner-icon="mdi-magnify"
    density="compact"
    data-test="search-text"
    @keyup="searchWebEndpoints"
  />

  <WebEndpointList
    v-if="showList"
    class="mt-2"
    data-test="web-endpoints-table-component"
  />

  <NoItemsMessage
    v-else
    class="mt-2"
    item="Web Endpoints"
    icon="mdi-web"
    data-test="no-items-message-component"
  >
    <template #content>
      <p>
        Web Endpoints enable secure, direct access to HTTP services on devices with the ShellHub Agent,
        eliminating the need for SSH local port forwarding.
      </p>
      <p>This simplifies connectivity, allowing users to access web-based interfaces seamlessly from their browser.</p>
    </template>
    <template #action>
      <v-btn
        color="primary"
        tabindex="0"
        variant="elevated"
        aria-label="Tunnel Create Dialog"
        data-test="tunnel-create-dialog-btn"
        :disabled="!canCreateWebEndpoint"
        @click="showWebEndpointCreate = true"
        @keypress.enter="showWebEndpointCreate = true"
      >
        Create Web Endpoint
      </v-btn>
    </template>
  </NoItemsMessage>

  <WebEndpointCreate
    v-model="showWebEndpointCreate"
    :use-devices-list="true"
    @update="searchWebEndpoints"
  />
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from "vue";
import WebEndpointList from "@/components/WebEndpoints/WebEndpointList.vue";
import NoItemsMessage from "@/components/NoItemsMessage.vue";
import PageHeader from "@/components/PageHeader.vue";
import useSnackbar from "@/helpers/snackbar";
import hasPermission from "@/utils/permission";
import WebEndpointCreate from "@/components/WebEndpoints/WebEndpointCreate.vue";
import useWebEndpointsStore from "@/store/modules/web_endpoints";
import handleError from "@/utils/handleError";

const webEndpointsStore = useWebEndpointsStore();
const snackbar = useSnackbar();
const filter = ref("");
const showList = computed(() => webEndpointsStore.showWebEndpoints);
const showWebEndpointCreate = ref(false);

// Keep the description on multiple shorter lines to satisfy the max-line-length rule
const webEndpointDescription = [
  "Secure direct access to HTTP services on your devices without SSH port forwarding.",
  "Access web-based interfaces seamlessly from your browser.",
].join(" ");

const canCreateWebEndpoint = hasPermission("webEndpoint:create");

const searchWebEndpoints = async () => {
  const addressFilter = [{
    type: "property",
    params: { name: "address", operator: "contains", value: filter.value },
  }];

  const encodedFilter = filter.value ? btoa(JSON.stringify(addressFilter)) : undefined;

  try {
    await webEndpointsStore.fetchWebEndpointsList({ filter: encodedFilter });
  } catch (error) {
    snackbar.showError("Failed to load web endpoints.");
    handleError(error);
  }
};

onMounted(async () => {
  try {
    await webEndpointsStore.fetchWebEndpointsList();
  } catch (error) {
    snackbar.showError("Failed to load web endpoints.");
    handleError(error);
  }
});
</script>
