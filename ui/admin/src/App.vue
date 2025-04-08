<template>
  <v-app>
    <component
      :is="layout"
      :data-test="`${layout}-component`" />
  </v-app>
</template>

<script setup lang="ts">
import { computed } from "vue";
import SimpleLayout from "./layouts/SimpleLayout.vue";
import AppLayout from "./layouts/AppLayout.vue";
import { useStore } from "./store";

const components = {
  SimpleLayout,
  AppLayout,
};

const store = useStore();

const layout = computed(() => components[store.getters["layout/getLayout"]]);
</script>
