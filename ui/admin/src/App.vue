<template>
  <v-app :theme>
    <AppLayout
      data-test="app-layout"
    />
  </v-app>
</template>

<script setup lang="ts">
import { computed } from "vue";
import AppLayout from "./layouts/AppLayout.vue";
import useLayoutStore from "@/store/modules/layout";
import "@/assets/global.css";

const layoutStore = useLayoutStore();
const theme = computed(() => layoutStore.theme);
</script>
