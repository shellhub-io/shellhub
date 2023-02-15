<template>
  <div
    class="d-flex flex-column justify-space-between align-center flex-sm-row"
  >
    <h1>Devices</h1>
    <div class="w-50">
      <v-text-field
        v-if="hasDevice"
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
    </div>

    <div class="d-flex mt-4">
      <TagSelector v-if="isDeviceList" data-test="tagSelector-component" />
      <DeviceAdd />
    </div>
  </div>
  <v-card class="mt-2" v-if="hasDevice">
    <Device />
  </v-card>

  <BoxMessage
    v-if="showMessageBox"
    class="mt-2"
    type-message="device"
    data-test="boxMessageDevice-component"
  />
</template>

<script lang="ts">
import { defineComponent, onMounted, computed, ref, onUnmounted } from "vue";
import { useRouter } from "vue-router";
import { useStore } from "../store";
import Device from "../components/Devices/Device.vue";
import DeviceAdd from "../components/Devices/DeviceAdd.vue";
import TagSelector from "../components/Tags/TagSelector.vue";
import BoxMessage from "../components/Box/BoxMessage.vue";

export default defineComponent({
  name: "Devices",
  setup() {
    const store = useStore();
    const router = useRouter();
    const filter = ref("");
    const show = ref(false);
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
        store.dispatch("snackbar/showSnackbarErrorDefault");
      }
    };

    const hasDevice = computed(() => (
      store.getters["stats/stats"].registered_devices > 0
        || store.getters["stats/stats"].pending_devices > 0
        || store.getters["stats/stats"].rejected_devices > 0
    ));

    const isDeviceList = computed(() => router.currentRoute.value.name === "listDevices");

    const showMessageBox = computed(() => !hasDevice.value && show.value);

    onMounted(async () => {
      try {
        await store.dispatch("stats/get");
        show.value = true;
      } catch (error: any) {
        if (error.response.status === 403) {
          store.dispatch("snackbar/showSnackbarErrorAssociation");
        } else {
          store.dispatch("snackbar/showSnackbarErrorDefault");
        }
        throw new Error(error);
      }
    });

    onUnmounted(async () => {
      await store.dispatch("devices/setFilter", null);
    });

    return {
      filter,
      searchDevices,
      hasDevice,
      isDeviceList,
      showMessageBox,
    };
  },
  components: { Device, DeviceAdd, TagSelector, BoxMessage },
});
</script>
