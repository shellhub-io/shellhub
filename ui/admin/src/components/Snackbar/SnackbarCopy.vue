<template>
  <v-snackbar
    v-model="snackbar"
    :timeout="2000"
    location="top"
    color="#F9F3EE"
    variant="tonal"
    transition="slide-x-transition"
  >
    {{ message }}
  </v-snackbar>
</template>

<script setup lang="ts">
import { computed } from "vue";
import useSnackbarStore from "@admin/store/modules/snackbar";

const props = defineProps({
  mainContent: {
    type: String,
    default: "",
    required: true,
  },
});
const snackbarStore = useSnackbarStore();

const snackbar = computed({
  get() {
    return snackbarStore.getSnackbarCopy;
  },
  set() {
    snackbarStore.unsetShowStatusSnackbarCopy();
  },
});

const message = computed(() => `${props.mainContent} copied to clipboard.`);

defineExpose({ snackbar, message });
</script>
