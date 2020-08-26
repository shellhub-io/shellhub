<template>
  <fragment>
    <snackbarSuccess
      :type-message="message.typeMessage"
    />
    <snackbarError
      :type-message="message.typeMessage"
      :main-content="message.typeContent"
    />
  </fragment>
</template>

<script>

import snackbarSuccess from '@/components/snackbar/SnackbarSuccess';
import snackbarError from '@/components/snackbar/SnackbarError';

export default {
  name: 'Snackbar',

  components: {
    snackbarSuccess,
    snackbarError,
  },

  computed: {
    message() {
      return this.$store.getters['modals/SnackbarMessageAndContentType'];
    },
  },
};

</script>
