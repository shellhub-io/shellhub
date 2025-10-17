<template>
  <div class="d-flex flex-column justify-space-between align-center flex-md-row mb-2">
    <h1>Namespaces</h1>
    <v-spacer />
    <v-text-field
      label="Search by name"
      color="primary"
      class="w-50"
      single-line
      hide-details
      v-model.trim="filter"
      v-on:keyup="searchNamespaces"
      append-inner-icon="mdi-magnify"
      density="compact"
    />
    <v-spacer />
    <NamespaceExport />
  </div>
  <NamespaceList />
</template>

<script setup lang="ts">
import { ref } from "vue";
import useNamespacesStore from "@admin/store/modules/namespaces";
import NamespaceList from "../components/Namespace/NamespaceList.vue";
import NamespaceExport from "../components/Namespace/NamespaceExport.vue";

const namespacesStore = useNamespacesStore();

const filter = ref("");

const searchNamespaces = async () => {
  const filterToEncodeBase64 = [{
    type: "property",
    params: { name: "name", operator: "contains", value: filter.value },
  }];

  const encodedFilter = filter.value ? btoa(JSON.stringify(filterToEncodeBase64)) : "";

  namespacesStore.setFilter(encodedFilter);

  await namespacesStore.fetchNamespaceList({ filter: encodedFilter, page: 1 });
};

defineExpose({ filter });
</script>
