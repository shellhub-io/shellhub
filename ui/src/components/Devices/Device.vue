<template>
  <div class="d-flex flex-wrap ga-2 mb-4 align-center justify-space-between">
    <div class="d-flex ga-2 align-center flex-wrap flex-grow-1">
      <v-btn-group
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
        label="Search by hostname"
        variant="outlined"
        color="primary"
        single-line
        hide-details
        prepend-inner-icon="mdi-magnify"
        density="compact"
        class="flex-grow-1"
        data-test="search-text"
        @update:model-value="updateDeviceListFilter"
      />
    </div>
    <TagSelector
      v-if="isDeviceList"
      variant="device"
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
import useDevicesStore from "@/store/modules/devices";

const devicesStore = useDevicesStore();
const route = useRoute();
const filter = ref("");

const states = [
  { to: "/devices", title: "Accepted" },
  { to: "/devices/pending", title: "Pending" },
  { to: "/devices/rejected", title: "Rejected" },
];

const isActive = (to: string) => {
  return route.path === to;
};

const isDeviceList = computed(() => route.name === "DeviceList");

const updateDeviceListFilter = () => {
  const base64DeviceFilter = [{
    type: "property",
    params: { name: "name", operator: "contains", value: filter.value },
  }];

  const encodedFilter = filter.value ? btoa(JSON.stringify(base64DeviceFilter)) : undefined;

  devicesStore.deviceListFilter = encodedFilter;
};
</script>
