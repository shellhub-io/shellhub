<template>
  <div class="d-flex flex-column justify-space-between align-center flex-md-row">
    <h1>Namespaces</h1>
    <div class="w-50">
      <v-text-field
        label="Search by name"
        variant="underlined"
        color="primary"
        single-line
        hide-details
        v-model.trim="filter"
        v-on:keyup="searchNamespaces"
        append-inner-icon="mdi-magnify"
        density="comfortable"
      />
    </div>
    <div class="mt-sm-4">
      <NamespaceExport />
    </div>
  </div>
  <v-card class="mt-2">
    <NamespaceList />
  </v-card>
</template>

<script lang="ts">
import { defineComponent, ref } from "vue";
import NamespaceList from "../components/Namespace/NamespaceList.vue";
import { useStore } from "../store";
import NamespaceExport from "../components/Namespace/NamespaceExport.vue";

export default defineComponent({
  setup() {
    const store = useStore();

    const filter = ref("");

    const searchNamespaces = () => {
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

      store.dispatch("namespaces/search", {
        perPage: store.getters["namespaces/perPage"],
        page: store.getters["namespaces/page"],
        filter: encodedFilter,
      });
    };

    return {
      filter,
      searchNamespaces,
    };
  },
  components: { NamespaceList, NamespaceExport },
});
</script>
