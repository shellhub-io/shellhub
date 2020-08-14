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
      {{ message[typeMessage] }}
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
  },

  data() {
    return {
      message: {
        default: 'The request has failed, please try again.',
      },
    };
  },

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
