<template>
  <fragment>
    <v-snackbar
      v-model="snackbar"
      :timeout="4000"
    >
      {{ message }}
    </v-snackbar>
  </fragment>
</template>

<script>

export default {
  name: 'SnackbarCopy',

  props: {
    mainContent: {
      type: String,
      default: '',
      required: true,
    },
  },

  computed: {
    snackbar: {
      get() {
        return this.$store.getters['modals/snackbarCopy'];
      },

      set() {
        this.$store.dispatch('modals/unsetShowStatusSnackbarCopy');
      },
    },

    message() {
      return `${this.mainContent} copied to clipboard.`;
    },
  },
};

</script>
