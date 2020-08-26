<template>
  <fragment>
    <v-snackbar
      v-model="snackbar"
      :timeout="4000"
      color="#bd4147"
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
  name: 'SnackbarError',

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
        return this.$store.getters['modals/snackbarError'];
      },

      set() {
        this.$store.dispatch('modals/unsetShowStatusSnackbarError');
      },
    },

    message() {
      switch (this.typeMessage) {
      case 'loading':
        return `Loading the ${this.mainContent} has failed, please try again.`;
      case 'action':
        return `The ${this.mainContent} request has failed, please try again.`;
      default:
        return 'The request has failed, please try again.';
      }
    },
  },
};

</script>
