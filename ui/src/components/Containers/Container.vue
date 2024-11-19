<template>
  <v-btn-group
    divided
    class="mb-4 border"
  >
    <v-btn
      v-for="state in states"
      :to="state.to"
      variant="flat"
      :active="isActive(state.to)"
      active-color="secondary"
      class="bg-background"
      v-bind:key="state.to"
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
  { to: "/containers", title: "Accepted" },
  { to: "/containers/pending", title: "Pending" },
  { to: "/containers/rejected", title: "Rejected" },
];

const isActive = (to: string) => {
  const route = useRoute();
  return route.path === to;
};

</script>
