<template>
  <snackbar-sucess
    :type-message="message.typeMessage"
    :main-content="message.typeContent"
  />

  <snackbar-error
    :type-message="message.typeMessage"
    :main-content="message.typeContent"
  />

  <snackbar-copy
    :main-content="message.typeContent"
  />
</template>

<script setup lang="ts">
import { computed } from "vue";
import useSnackbarStore from "@admin/store/modules/snackbar";
import SnackbarSucess from "./SnackbarSucess.vue";
import SnackbarError from "./SnackbarError.vue";
import SnackbarCopy from "./SnackbarCopy.vue";

const snackbarStore = useSnackbarStore();

const message = computed(() => snackbarStore.getSnackbarMessageAndContentType);

defineExpose({ message });
</script>
