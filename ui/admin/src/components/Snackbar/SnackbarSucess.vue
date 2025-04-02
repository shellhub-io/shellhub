<template>
  <v-snackbar
    v-model="snackbar"
    location="top"
    :timeout="4000"
    color="#4caf50"
    variant="tonal"
    transition="slide-x-transition"
  >
    <p class="w-100 text-center">{{ message }}</p>
  </v-snackbar>
</template>

<script setup lang="ts">
import { computed } from "vue";
import { useStore } from "../../store";

const props = defineProps({
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
  switch (props.typeMessage) {
    case "action":
      return `The ${props.mainContent} has succeeded.`;
    case "no-content":
      return "There is no content to export";
    default:
      return "The request has succeeded.";
  }
});

defineExpose({ snackbar, message });
</script>
