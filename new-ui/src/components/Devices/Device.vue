<template>
  <v-card class="bg-v-theme-surface">
    <v-tabs background-color="secondary" stacked color="primary">
      <v-tab to="/devices"> Device List </v-tab>
      <v-tab to="/devices/pending"> Pending </v-tab>

      <v-tab to="/devices/rejected"> Rejected </v-tab>
    </v-tabs>

    <v-divider />
  </v-card>

  <v-card>
    <router-view />
  </v-card>
</template>

<script lang="ts">
import { defineComponent, ref } from "vue";
import DeviceList from "./DeviceList.vue";
import DevicePendingList from "./DevicePendingList.vue";
import DeviceRejectedList from "./DeviceRejectedList.vue";

export default defineComponent({
  name: "Device",
  setup() {
    const tab = ref("list");
    return {
      tab,
    };
  },
  components: { DeviceList, DevicePendingList, DeviceRejectedList },
});
</script>
