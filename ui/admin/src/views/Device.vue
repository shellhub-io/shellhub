<template>
  <div class="d-flex flex-column justify-space-between align-center flex-sm-row">
    <h1>Devices</h1>
    <v-spacer />
    <div class="w-50">
      <v-text-field
        label="Search by hostname"
        variant="underlined"
        color="primary"
        single-line
        hide-details
        v-model.trim="filter"
        v-on:keyup="searchDevices"
        append-inner-icon="mdi-magnify"
        density="comfortable"
      />
    </div>
    <v-spacer />
  </div>
  <v-card class="mt-2">
    <DeviceList />
  </v-card>
</template>

<script lang="ts">
import { defineComponent, ref } from "vue";
import { useStore } from "../store";
import DeviceList from "../components/Device/DeviceList.vue";

export default defineComponent({
  setup() {
    const store = useStore();

    const filter = ref("");

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
          perPage: store.getters["devices/perPage"],
          page: store.getters["devices/page"],
          filter: encodedFilter,
        });
      } catch {
        store.dispatch("snackbar/showSnackbarErrorDefault");
      }
    };

    return {
      filter,
      searchDevices,
    };
  },
  components: { DeviceList },
});
</script>
