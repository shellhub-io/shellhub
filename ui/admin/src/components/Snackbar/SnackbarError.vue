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
import useSnackbarStore from "@admin/store/modules/snackbar";

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

const snackbarStore = useSnackbarStore();

const snackbar = computed({
  get() {
    return snackbarStore.getSnackbarError;
  },
  set() {
    snackbarStore.unsetShowStatusSnackbarError();
  },
});

const message = computed(() => {
  switch (props.typeMessage) {
    case "custom":
      return props.mainContent;
    case "loading":
      return `Loading the ${props.mainContent} has failed, please try again.`;
    case "action":
      return `The ${props.mainContent} request has failed, please try again.`;
    case "licenseRequired":
      return `The ${props.mainContent} request has failed, license required.`;
    default:
      return "The request has failed, please try again.";
  }
});

defineExpose({ snackbar, message });
</script>
