<template>
  <v-row class="align-center">
    <v-col cols="12" sm="4" md="3">
      <h1 class="text-center text-sm-left">Web Endpoints</h1>
    </v-col>

    <v-col cols="12" sm="4" md="6">
      <v-text-field
        label="Search by Address"
        variant="outlined"
        color="primary"
        single-line
        hide-details
        v-model.trim="filter"
        v-on:keyup="searchWebEndpoints"
        prepend-inner-icon="mdi-magnify"
        density="compact"
        data-test="search-text"
      />
    </v-col>

    <v-col cols="12" sm="4" md="3" class="d-flex justify-center justify-sm-end mt-2 mt-sm-0">
      <v-btn
        @click="showWebEndpointCreate = true"
        color="primary"
        variant="elevated"
        aria-label="Tunnel Create Dialog"
        @keypress.enter="showWebEndpointCreate = true"
        data-test="tunnel-create-dialog-btn"
        :disabled="!hasAuthorizationCreateWebEndpoint"
      >
        Create Web Endpoint
      </v-btn>
    </v-col>
  </v-row>

  <div v-if="showList" class="mt-2" data-test="web-endpoints-table-component">
    <WebEndpointList />
  </div>

  <NoItemsMessage
    v-else
    class="mt-2"
    item="Web Endpoints"
    icon="mdi-web"
    data-test="no-items-message-component"
  >
    <template #action>
      <v-btn
        @click="showWebEndpointCreate = true"
        color="primary"
        tabindex="0"
        variant="elevated"
        aria-label="Tunnel Create Dialog"
        @keypress.enter="showWebEndpointCreate = true"
        data-test="tunnel-create-dialog-btn"
        :disabled="!hasAuthorizationCreateWebEndpoint"
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
import { actions, authorizer } from "@/authorizer";
import WebEndpointCreate from "@/components/WebEndpoints/WebEndpointCreate.vue";
import useAuthStore from "@/store/modules/auth";
import useWebEndpointsStore from "@/store/modules/web_endpoints";

const authStore = useAuthStore();
const webEndpointsStore = useWebEndpointsStore();
const snackbar = useSnackbar();
const filter = ref("");
const showList = computed(() => webEndpointsStore.showWebEndpoints);
const showWebEndpointCreate = ref(false);

const hasAuthorizationCreateWebEndpoint = () => {
  const { role } = authStore;
  return !!role && hasPermission(authorizer.role[role], actions.tunnel.create);
};

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
