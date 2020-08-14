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
      {{ message[typeMessage] }}
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
  },

  data() {
    return {
      message: {
        default: 'The request has succeeded.',
      },
    };
  },

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
