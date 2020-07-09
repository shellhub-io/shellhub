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
      The request has succeeded.
    </v-snackbar>
  </fragment>
</template>

<script>

export default {
  name: 'SnackbarSuccess',

  computed: {
    snackbar: {
      get() {
        return this.$store.getters['modals/snackbarSuccess'];
      },

      set(value) {
        this.$store.dispatch('modals/showSnackbarSuccess', value);
      },
    },
  },
};

</script>
