<template>
  <v-app :theme>
    <component
      :is="layout"
      :data-test="`${layout}-component`" />
  </v-app>
</template>

<script setup lang="ts">
import { computed } from "vue";
import SimpleLayout from "./layouts/SimpleLayout.vue";
import AppLayout from "./layouts/AppLayout.vue";
import useLayoutStore from "./store/modules/layout";

const components = {
  SimpleLayout,
  AppLayout,
};

const layoutStore = useLayoutStore();
const theme = computed(() => layoutStore.theme);
const layout = computed(() => components[layoutStore.layout]);
</script>
