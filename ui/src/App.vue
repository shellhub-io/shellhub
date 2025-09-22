<template>
  <v-app :theme="theme">
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
import useLayoutStore from "@/store/modules/layout";
import "./assets/global.css";

const components = {
  AppLayout,
  LoginLayout,
};
const layoutStore = useLayoutStore();
const layout = computed(() => components[layoutStore.layout || ""]);
const theme = computed(() => layoutStore.theme);
</script>
