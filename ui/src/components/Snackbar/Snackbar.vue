<template>
  <snackbar-sucess :type-message="message.typeMessage" :main-content="message.typeContent" />

  <snackbar-error :type-message="message.typeMessage" :main-content="message.typeContent" />

  <snackbar-copy :main-content="message.typeContent" />
</template>

<script lang="ts">
import { defineComponent, computed } from "vue";
import SnackbarSucess from "./SnackbarSucess.vue";
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
  components: { SnackbarSucess, SnackbarError, SnackbarCopy },
});
</script>
