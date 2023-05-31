<template>
  <v-app>
    <component
      :is="layout"
      :data-test="layout + '-component'"
    />
  </v-app>
</template>

<script setup lang="ts">
import { computed } from "vue";
import LoginLayout from "./layouts/LoginLayout.vue";
import AppLayout from "./layouts/AppLayout.vue";
import { useStore } from "./store";

import "./assets/global.css";

const components = {
  AppLayout,
  LoginLayout,
};
const store = useStore();
const layout = computed(() => components[store.getters["layout/getLayout"]]);
</script>
