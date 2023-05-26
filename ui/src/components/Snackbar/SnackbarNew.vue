<template>
  <v-snackbar
    v-model="show"
    location="top"
    :timeout="4000"
    :color="type"
    ransition="slide-x-transition"
  >
    {{ message }}
  </v-snackbar>
</template>

<script setup lang="ts">
import { ref } from "vue";
import { useStore } from "../../store";

const store = useStore();
const show = ref(false);
const message = ref("");
const type = ref("");

store.subscribe((mutation, state) => {
  if (mutation.type === "snackbar/showMessage") {
    message.value = state.snackbar.message;
    type.value = state.snackbar.type;
    show.value = true;
  }
});
</script>
