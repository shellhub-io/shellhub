<template>
  <div
    class="d-flex flex-column justify-space-between align-center flex-sm-row"
    data-test="device-title"
  >
    <h1>Containers</h1>
    <v-col md="6" sm="12">
      <v-text-field
        v-if="show"
        label="Search by hostname"
        variant="underlined"
        color="primary"
        single-line
        hide-details
        v-model.trim="filter"
        v-on:keyup="searchDevices"
        append-inner-icon="mdi-magnify"
        density="comfortable"
        data-test="search-text"
      />
    </v-col>

    <div class="d-flex mt-4" data-test="device-header-component-group">
      <TagSelector v-if="isContainerList" />
    </div>
  </div>
  <v-card :loading="loading" class="mt-2" v-if="show" data-test="device-table-component">
    <Containers />
  </v-card>

  <BoxMessage
    v-if="!show"
    :loading="loading"
    class="mt-2"
    type-message="container"
    data-test="boxMessageDevice-component"
  >
    <template v-slot:container v-if="envVariables.hasConnector">
      <ConnectorAdd @update="refresh" />
    </template>
  </BoxMessage>
</template>

<script setup lang="ts">
import { computed, ref, onUnmounted } from "vue";
import { useRouter } from "vue-router";
import { useStore } from "../store";
import { envVariables } from "../envVariables";
import ConnectorAdd from "../components/Connector/ConnectorAdd.vue";
import Containers from "../components/Containers/Container.vue";
import TagSelector from "../components/Tags/TagSelector.vue";
import BoxMessage from "../components/Box/BoxMessage.vue";

const store = useStore();
const router = useRouter();
const filter = ref("");
const show = computed(() => store.getters["devices/getNumberDevices"] > 0);
const loading = ref(false);

const filterToEncodeBase64 = [
  {
    type: "property",
    params: {
      name: "info.platform",
      operator: "eq",
      value: "connector",
    },
  },
  {
    type: "property",
    params: { name: "name", operator: "contains", value: filter.value },
  },
];

const searchDevices = () => {
  let encodedFilter = "";

  if (filter.value) {
    encodedFilter = btoa(JSON.stringify(filterToEncodeBase64));
  }

  try {
    store.dispatch("devices/search", {
      page: store.getters["devices/getPage"],
      perPage: store.getters["devices/getPerPage"],
      filter: encodedFilter,
      status: store.getters["devices/getStatus"],
    });
  } catch {
    store.dispatch("snackbar/showSnackbarErrorDefault");
  }
};

const isContainerList = computed(() => router.currentRoute.value.name === "listContainers");

const refresh = async () => {
  loading.value = true;
  setTimeout(() => {
    store.dispatch("devices/fetch", {
      page: store.getters["devices/getPage"],
      perPage: store.getters["devices/getPerPage"],
      filter: btoa(JSON.stringify(filterToEncodeBase64)),
      committable: false,
    });
    loading.value = false;
  }, 10000);
};

onUnmounted(async () => {
  await store.dispatch("devices/setFilter", "");
});

</script>
