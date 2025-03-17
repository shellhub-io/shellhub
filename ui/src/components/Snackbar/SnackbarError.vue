<template>
  <v-snackbar
    v-model="snackbar"
    :timeout="4000"
    color="#bd4147"
    location="top"
    transition="slide-x-transition"
  >
    {{ message }}
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
    return store.getters["snackbar/snackbarError"];
  },
  set() {
    store.dispatch("snackbar/unsetShowStatusSnackbarError");
  },
});

const message = computed(() => {
  switch (typeMessage) {
    case "loading":
      return `Loading the ${mainContent} has failed, please try again.`;
    case "action":
      return `The ${mainContent} request has failed, please try again.`;
    case "licenseRequired":
      return `The ${mainContent} request has failed, license required.`;
    default:
      return "The request has failed, please try again.";
  }
});
</script>
