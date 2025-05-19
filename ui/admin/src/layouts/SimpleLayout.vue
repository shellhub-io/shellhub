<template>
  <Snackbar />
  <router-view :key="currentRoute.value.path" />
</template>

<script setup lang="ts">
import { computed } from "vue";
import { useRouter } from "vue-router";
import Snackbar from "@/components/Snackbar/Snackbar.vue";

const router = useRouter();

const currentRoute = computed(() => router.currentRoute);
</script>
