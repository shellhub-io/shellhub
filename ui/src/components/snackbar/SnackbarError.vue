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
      The request has failed, please try again.
    </v-snackbar>
  </fragment>
</template>

<script>

export default {
  name: 'SnackbarError',

  computed: {
    snackbar: {
      get() {
        return this.$store.getters['modals/snackbarError'];
      },

      set(value) {
        this.$store.dispatch('modals/showSnackbarError', value);
      },
    },
  },
};

</script>
