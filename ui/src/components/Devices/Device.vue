<template>
  <v-btn-group
    color="a"
    divided
    class="mb-4 border"
  >
    <v-btn
      v-for="state in states"
      :key="state.to"
      :to="state.to"
      variant="flat"
      :active="isActive(state.to)"
      active-color="secondary"
      class="bg-background"
    >
      {{ state.title }}
    </v-btn>
  </v-btn-group>

  <div>
    <router-view />
  </div>
</template>

<script setup lang="ts">
import { useRoute } from "vue-router";

const states = [
  { to: "/devices", title: "Accepted" },
  { to: "/devices/pending", title: "Pending" },
  { to: "/devices/rejected", title: "Rejected" },
];

const isActive = (to: string) => {
  const route = useRoute();
  return route.path === to;
};

</script>
