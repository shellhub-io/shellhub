<template>
  <v-snackbar
    v-model="snackbar"
    location="top"
    :timeout="4000"
    color="#4caf50"
    transition="slide-x-transition"
  >
    <p class="w-100 text-center">{{ message }}</p>
  </v-snackbar>
</template>

<script setup lang="ts">
import { computed } from "vue";
import { useStore } from "../../store";

const { mainContent, typeMessage } = defineProps({
  typeMessage: {
    type: String,
    required: true,
  },

  mainContent: {
    type: String,
    default: "",
    required: false,
  },
});

const store = useStore();

const snackbar = computed({
  get() {
    return store.getters["snackbar/snackbarSuccess"];
  },
  set() {
    store.dispatch("snackbar/unsetShowStatusSnackbarSuccess");
  },
});

const message = computed(() => {
  switch (typeMessage) {
    case "action":
      return `The ${mainContent} has succeeded.`;
    case "no-content":
      return "There is no content to export";
    default:
      return "The request has succeeded.";
  }
});
</script>
