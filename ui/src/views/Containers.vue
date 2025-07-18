<template>
  <div
    class="d-flex flex-column justify-space-between align-center flex-sm-row"
    data-test="device-title"
  >
    <h1>Containers</h1>
    <v-col md="6">
      <v-text-field
        v-if="show"
        label="Search by hostname"
        variant="outlined"
        color="primary"
        single-line
        hide-details
        v-model.trim="filter"
        v-on:keyup="searchDevices"
        prepend-inner-icon="mdi-magnify"
        density="compact"
        data-test="search-text"
      />
    </v-col>

    <div class="d-flex" data-test="device-header-component-group">
      <TagSelector variant="container" v-if="isContainerList" />
      <ContainerAdd />
    </div>

  </div>
  <div class="mt-2" v-if="show" data-test="device-table-component">
    <Containers />
  </div>

  <NoItemsMessage
    v-else
    class="mt-2"
    item="Containers"
    icon="mdi-server"
    data-test="no-items-message-component"
  >
    <template #content>
      <p>In order to register a container on ShellHub, you need to configure a Docker Connector.</p>
      <p>To view and connect to your containers in ShellHub, please add a Docker Engine connector.
        This will allow you to connect to your Docker Engine and see all your containers here.</p>
    </template>
    <template #action>
      <ContainerAdd />
    </template>
  </NoItemsMessage>
</template>

<script setup lang="ts">
import { computed, ref, onUnmounted } from "vue";
import { useRouter } from "vue-router";
import { useStore } from "../store";
import Containers from "../components/Containers/Container.vue";
import TagSelector from "../components/Tags/TagSelector.vue";
import NoItemsMessage from "../components/NoItemsMessage.vue";
import ContainerAdd from "../components/Containers/ContainerAdd.vue";
import useSnackbar from "@/helpers/snackbar";

const store = useStore();
const router = useRouter();
const filter = ref("");
const show = computed(() => store.getters["container/getShowContainers"]);
const snackbar = useSnackbar();

const searchDevices = () => {
  let encodedFilter = "";

  if (filter.value) {
    const filterToEncodeBase64 = [
      {
        type: "property",
        params: { name: "name", operator: "contains", value: filter.value },
      },
    ];
    encodedFilter = btoa(JSON.stringify(filterToEncodeBase64));
  }

  try {
    store.dispatch("container/search", {
      page: store.getters["container/getPage"],
      perPage: store.getters["container/getPerPage"],
      filter: encodedFilter,
      status: store.getters["container/getStatus"],
    });
  } catch {
    snackbar.showError("An error occurred while searching for containers.");
  }
};

const isContainerList = computed(() => router.currentRoute.value.name === "ContainerList");

onUnmounted(async () => {
  await store.dispatch("container/setFilter", "");
});
</script>
