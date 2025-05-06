<template>
  <v-snackbar
    v-model="show"
    location="top"
    :timeout="4000"
    :color="color"
    transition="slide-x-transition"
  >
    <p class="w-100 text-center">{{ message }}</p>
  </v-snackbar>
</template>

<script setup lang="ts">
import { computed } from "vue";
import { plugin } from "@/plugins/snackbar";

const show = computed(() => plugin.getShow());

const message = computed(() => plugin.getMessage());
const type = computed(() => plugin.getType());
const color = computed(() => type.value);
</script>
