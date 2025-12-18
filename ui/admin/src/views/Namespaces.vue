<template>
  <PageHeader
    icon="mdi-cloud-braces"
    title="Namespaces"
    overline="Namespace Management"
    description="Track every tenant, search by name, and export namespace data for audits."
    icon-color="primary"
  >
    <template #actions>
      <NamespaceExport />
    </template>

    <v-text-field
      v-model.trim="filter"
      label="Search by name"
      color="primary"
      class="w-100 w-md-50"
      single-line
      hide-details
      append-inner-icon="mdi-magnify"
      density="compact"
      @keyup="searchNamespaces"
    />
  </PageHeader>
  <NamespaceList />
</template>

<script setup lang="ts">
import { ref } from "vue";
import PageHeader from "@/components/PageHeader.vue";
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
