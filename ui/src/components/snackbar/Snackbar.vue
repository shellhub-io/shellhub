<template>
  <fragment>
    <snackbarSuccess
      :type-message="typeMessage"
    />
    <snackbarError
      :type-message="typeMessage"
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

  props: {
    typeMessage: {
      type: String,
      default: 'default',
      validator: (value) => ['default'].includes(value),
    },
  },

};

</script>
