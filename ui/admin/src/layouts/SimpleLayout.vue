<template>
  <router-view :key="currentRoute.value.path" />
</template>

<script lang="ts">
import { computed } from "vue";
import { useRouter } from "vue-router";

export default {
  name: "SimpleLayout",
  setup() {
    const router = useRouter();

    const currentRoute = computed(() => router.currentRoute);

    return {
      currentRoute,
    };
  },
};
</script>
