<template>
  <fragment>
    <snackbarSuccess
      :type-message="message.typeMessage"
      :main-content="message.typeContent"
      data-test="snackbarSuccess-component"
    />
    <snackbarError
      :type-message="message.typeMessage"
      :main-content="message.typeContent"
      data-test="snackbarError-component"
    />
    <snackbarCopy
      :main-content="message.typeContent"
      data-test="snackbarCopy-component"
    />
  </fragment>
</template>

<script>

import snackbarSuccess from '@/components/snackbar/SnackbarSuccess';
import snackbarError from '@/components/snackbar/SnackbarError';
import snackbarCopy from '@/components/snackbar/SnackbarCopy';

export default {
  name: 'SnackbarComponent',

  components: {
    snackbarSuccess,
    snackbarError,
    snackbarCopy,
  },

  computed: {
    message() {
      return this.$store.getters['snackbar/snackbarMessageAndContentType'];
    },
  },
};

</script>
