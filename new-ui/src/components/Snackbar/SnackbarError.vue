<template>
  <v-snackbar
    v-model="snackbar"
    :timeout="4000"
    color="#bd4147"
    location="top"
    variant="tonal"
    text
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
    typeMessage: {
      type: String,
      required: true,
    },

    mainContent: {
      type: String,
      default: "",
      required: false,
    },
  },
  setup(props) {
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
      switch (props.typeMessage) {
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

    return {
      snackbar,
      message,
    };
  },
});
</script>
