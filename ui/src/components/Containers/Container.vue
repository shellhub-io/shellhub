<template>
  <div class="d-flex flex-wrap ga-2 mb-4 align-center justify-space-between">
    <div class="d-flex ga-2 align-center flex-wrap flex-grow-1">
      <v-btn-group
        color="a"
        divided
        density="default"
        class="border"
      >
        <v-btn
          v-for="state in states"
          :key="state.to"
          :to="state.to"
          variant="flat"
          :active="isActive(state.to)"
          active-color="secondary"
          class="bg-background"
        >
          {{ state.title }}
        </v-btn>
      </v-btn-group>
      <v-text-field
        v-model.trim="filter"
        label="Search by name"
        variant="outlined"
        color="primary"
        single-line
        hide-details
        prepend-inner-icon="mdi-magnify"
        density="compact"
        class="flex-grow-1"
        data-test="search-text"
        @update:model-value="updateContainerListFilter"
      />
    </div>
    <TagSelector
      v-if="isContainerList"
      variant="container"
    />
  </div>

  <div>
    <router-view />
  </div>
</template>

<script setup lang="ts">
import { computed, ref } from "vue";
import { useRoute } from "vue-router";
import TagSelector from "../Tags/TagSelector.vue";
import useContainersStore from "@/store/modules/containers";

const containersStore = useContainersStore();
const route = useRoute();
const filter = ref("");

const states = [
  { to: "/containers", title: "Accepted" },
  { to: "/containers/pending", title: "Pending" },
  { to: "/containers/rejected", title: "Rejected" },
];

const isActive = (to: string) => {
  return route.path === to;
};

const isContainerList = computed(() => route.name === "ContainerList");

const updateContainerListFilter = () => {
  const base64ContainerFilter = [{
    type: "property",
    params: { name: "name", operator: "contains", value: filter.value },
  }];

  const encodedFilter = filter.value ? btoa(JSON.stringify(base64ContainerFilter)) : undefined;

  containersStore.containerListFilter = encodedFilter;
};
</script>
