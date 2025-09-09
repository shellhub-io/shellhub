<template>
  <v-row class="align-center justify-space-between flex-column flex-sm-row mb-2 ga-4">
    <h1 class="text-center text-sm-left">Web Endpoints</h1>

    <v-text-field
      v-if="showList"
      class="w-75 w-sm-auto"
      label="Search by Address"
      variant="outlined"
      color="primary"
      single-line
      hide-details
      v-model.trim="filter"
      @keyup="searchWebEndpoints"
      prepend-inner-icon="mdi-magnify"
      density="compact"
      data-test="search-text"
    />
    <v-btn
      @click="showWebEndpointCreate = true"
      color="primary"
      variant="elevated"
      @keypress.enter="showWebEndpointCreate = true"
      data-test="tunnel-create-dialog-btn"
      :disabled="!canCreateWebEndpoint"
    >
      Create Web Endpoint
    </v-btn>
  </v-row>

  <WebEndpointList v-if="showList" class="mt-2" data-test="web-endpoints-table-component" />

  <NoItemsMessage
    v-else
    class="mt-2"
    item="Web Endpoints"
    icon="mdi-web"
    data-test="no-items-message-component"
  >
    <template #content>
      <p>Web Endpoints enable secure, direct access to HTTP services on devices with the ShellHub Agent,
        eliminating the need for SSH local port forwarding.</p>
      <p>This simplifies connectivity, allowing users to access web-based interfaces seamlessly from their browser.</p>
    </template>
    <template #action>
      <v-btn
        @click="showWebEndpointCreate = true"
        color="primary"
        tabindex="0"
        variant="elevated"
        aria-label="Tunnel Create Dialog"
        @keypress.enter="showWebEndpointCreate = true"
        data-test="tunnel-create-dialog-btn"
        :disabled="!canCreateWebEndpoint"
      >
        Create Web Endpoint
      </v-btn>
    </template>
  </NoItemsMessage>

  <WebEndpointCreate v-model="showWebEndpointCreate" @update="searchWebEndpoints" :useDevicesList="true" />
</template>

<script setup lang="ts">
import { computed, ref } from "vue";
import WebEndpointList from "@/components/WebEndpoints/WebEndpointList.vue";
import NoItemsMessage from "@/components/NoItemsMessage.vue";
import useSnackbar from "@/helpers/snackbar";
import hasPermission from "@/utils/permission";
import WebEndpointCreate from "@/components/WebEndpoints/WebEndpointCreate.vue";
import useWebEndpointsStore from "@/store/modules/web_endpoints";

const webEndpointsStore = useWebEndpointsStore();
const snackbar = useSnackbar();
const filter = ref("");
const showList = computed(() => webEndpointsStore.showWebEndpoints);
const showWebEndpointCreate = ref(false);

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
  }
};

defineExpose({ searchWebEndpoints });
</script>
