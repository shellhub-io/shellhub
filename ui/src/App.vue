<template>
  <v-app>
    <component :is="layout" :data-test="layout + '-component'" />
  </v-app>
</template>

<script lang="ts">
import { defineComponent, computed } from "vue";
import SimpleLayout from "./layouts/SimpleLayout.vue";
import AppLayout from "./layouts/AppLayout.vue";
import { useStore } from "./store";

export default defineComponent({
  name: "App",
  components: {
    appLayout: AppLayout,
    simpleLayout: SimpleLayout,
  },
  setup() {
    const store = useStore();
    const layout = computed(() => store.getters["layout/getLayout"]);

    return {
      layout,
    };
  },
});
</script>
