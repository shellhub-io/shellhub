<template>
  <v-snackbar
    v-model="snackbar"
    :timeout="2000"
    location="top"
    :color="color"
    transition="slide-x-transition"
  >
    {{ message }}
  </v-snackbar>
</template>

<script setup lang="ts">
import { computed } from "vue";
import { useStore } from "@/store";

const { mainContent } = defineProps({
  mainContent: {
    type: String,
    default: "",
    required: true,
  },
});

const store = useStore();

const color = computed(() => store.getters["layout/getStatusDarkMode"] === "dark" ? "#F9F3EE" : "#1E1E1E");

const snackbar = computed({
  get() {
    return store.getters["snackbar/snackbarCopy"];
  },
  set() {
    store.dispatch("snackbar/unsetShowStatusSnackbarCopy");
  },
});

const message = computed(() => `${mainContent} copied to clipboard.`);
</script>
