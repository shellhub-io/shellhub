<template>
  <fragment>
    <v-snackbar
      :timeout="4000"
      :value="ok"
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
  name: 'SnackbarError',

  props: {
    ok: {
      type: Boolean,
      required: true,
    },
    message: {
      type: String,
      default: 'The request was a success.',
    },
  },

};

</script>
