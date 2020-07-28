<template>
  <fragment>
    <v-snackbar
      :timeout="4000"
      :value="error"
      color="#bd4147"
      absolute
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

  props: {
    error: {
      type: Boolean,
      required: true,
    },
  },

};

</script>
