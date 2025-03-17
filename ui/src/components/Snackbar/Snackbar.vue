<template>
  <snackbar-success :type-message="message.typeMessage" :main-content="message.typeContent" />

  <snackbar-error :type-message="message.typeMessage" :main-content="message.typeContent" />

  <snackbar-copy :main-content="message.typeContent" />
</template>

<script lang="ts">
import { defineComponent, computed } from "vue";
import SnackbarSuccess from "./SnackbarSuccess.vue";
import SnackbarError from "./SnackbarError.vue";
import SnackbarCopy from "./SnackbarCopy.vue";
import { useStore } from "../../store";

export default defineComponent({
  setup() {
    const store = useStore();

    const message = computed(() => store.getters["snackbar/snackbarMessageAndContentType"]);

    return {
      message,
    };
  },
  components: { SnackbarSuccess, SnackbarError, SnackbarCopy },
});
</script>
