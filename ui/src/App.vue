<template>
  <v-app>
    <Snackbar data-test="snackbar-component" />

    <component
      :is="layout"
      :data-test="layout+'-component'"
    />
  </v-app>
</template>

<script>

import AppLayout from '@/layouts/AppLayout';
import SimpleLayout from '@/layouts/SimpleLayout';
import Snackbar from '@/components/snackbar/Snackbar';

export default {
  name: 'App',

  // Define as many layouts you want for the application
  components: {
    appLayout: AppLayout,
    simpleLayout: SimpleLayout,
    Snackbar,
  },

  computed: {
    layout() {
      return this.$store.getters['layout/getLayout'];
    },

    isLoggedIn() {
      return this.$store.getters['auth/isLoggedIn'];
    },

    hasLoggedID() {
      return this.$store.getters['auth/id'] !== '';
    },
  },

  async created() {
    if (!this.isLoggedIn) {
      this.$store.dispatch('layout/setLayout', 'simpleLayout');
    }

    if (!this.hasLoggedID && this.isLoggedIn) {
      try {
        await this.$store.dispatch('auth/logout');

        this.$store.dispatch('layout/setLayout', 'simpleLayout');
        this.$router.push('/login');

        this.$store.dispatch('snackbar/showSnackbarSuccessAction', this.$success.namespaceReload);
      } catch {
        this.$store.dispatch('snackbar/showSnackbarErrorAction', this.$error.namespaceLoad);
      }
    }
  },
};

</script>
