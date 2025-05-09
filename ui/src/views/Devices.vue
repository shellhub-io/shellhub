<template>
  <div
    class="d-flex flex-column justify-space-between align-center flex-sm-row"
    data-test="device-title"
  >
    <h1>Devices</h1>
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
      <TagSelector variant="device" v-if="isDeviceList" />
      <DeviceAdd />
    </div>
  </div>
  <div class="mt-2" v-if="show" data-test="device-table-component">
    <Device />
  </div>

  <BoxMessage
    v-if="!show"
    class="mt-2"
    type-message="device"
    data-test="boxMessageDevice-component"
  />
</template>

<script setup lang="ts">
import { computed, ref, onUnmounted } from "vue";
import { useRouter } from "vue-router";
import { useStore } from "../store";
import Device from "../components/Devices/Device.vue";
import DeviceAdd from "../components/Devices/DeviceAdd.vue";
import TagSelector from "../components/Tags/TagSelector.vue";
import BoxMessage from "../components/Box/BoxMessage.vue";
import useSnackbar from "@/helpers/snackbar";

const store = useStore();
const router = useRouter();
const snackbar = useSnackbar();
const filter = ref("");
const show = computed(() => store.getters["devices/getShowDevices"]);

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
    store.dispatch("devices/search", {
      page: store.getters["devices/getPage"],
      perPage: store.getters["devices/getPerPage"],
      filter: encodedFilter,
      status: store.getters["devices/getStatus"],
    });
  } catch {
    snackbar.showError("Failed to load devices.");
  }
};

const isDeviceList = computed(() => router.currentRoute.value.name === "DeviceList");

onUnmounted(async () => {
  await store.dispatch("devices/setFilter", "");
});
</script>
