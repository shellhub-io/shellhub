<template>
  <fragment>
    <snackbarSuccess />
    <snackbarError />
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
};

</script>
