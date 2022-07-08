<template>
  <v-snackbar
    v-model="snackbar"
    :timeout="2000"
    location="top"
    :color="color"
    variant="tonal"
    transition="slide-x-transition"
  >
    {{ message }}
  </v-snackbar>
</template>

<script lang="ts">
import { computed, defineComponent } from "vue";
import { useStore } from "../../store";

export default defineComponent({
  props: {
    mainContent: {
      type: String,
      default: "",
      required: true,
    },
  },
  setup(props) {
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

    const message = computed(() => `${props.mainContent} copied to clipboard.`);

    return {
      color,
      snackbar,
      message,
    };
  },
});
</script>
