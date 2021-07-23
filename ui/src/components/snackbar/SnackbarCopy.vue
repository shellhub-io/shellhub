<template>
  <fragment>
    <v-snackbar
      v-model="snackbar"
      :timeout="4000"
      data-test="message-snackbar"
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
        return this.$store.getters['snackbar/snackbarCopy'];
      },

      set() {
        this.$store.dispatch('snackbar/unsetShowStatusSnackbarCopy');
      },
    },

    message() {
      return `${this.mainContent} copied to clipboard.`;
    },
  },
};

</script>
