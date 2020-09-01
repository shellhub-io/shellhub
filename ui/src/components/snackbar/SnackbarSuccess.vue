<template>
  <fragment>
    <v-snackbar
      v-model="snackbar"
      :timeout="4000"
      color="#4caf50"
      top
      outlined
      text
    >
      {{ message }}
    </v-snackbar>
  </fragment>
</template>

<script>

export default {
  name: 'SnackbarSuccess',

  props: {
    typeMessage: {
      type: String,
      required: true,
    },

    mainContent: {
      type: String,
      default: '',
      required: false,
    },
  },

  computed: {
    snackbar: {
      get() {
        return this.$store.getters['modals/snackbarSuccess'];
      },

      set() {
        this.$store.dispatch('modals/unsetShowStatusSnackbarSuccess');
      },
    },

    message() {
      switch (this.typeMessage) {
      case 'action':
        return `The ${this.mainContent} has succeeded.`;
      default:
        return 'The request has succeeded.';
      }
    },
  },
};

</script>
