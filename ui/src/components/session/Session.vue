<template>
  <fragment>
    <div class="d-flex pa-0 align-center">
      <h1>Sessions</h1>

      <v-spacer />
      <v-spacer />
    </div>

    <v-card class="mt-2">
      <router-view
        v-if="hasSession"
      />

      <BoxMessageSession
        v-if="showBoxMessage"
        type-message="session"
        data-test="BoxMessageSession-component"
      />
    </v-card>
  </fragment>
</template>

<script>

import BoxMessageSession from '@/components/box/BoxMessage';

export default {
  name: 'Session',

  components: {
    BoxMessageSession,
  },

  data() {
    return {
      show: false,
    };
  },

  computed: {
    hasSession() {
      return this.$store.getters['sessions/getNumberSessions'] > 0;
    },

    showBoxMessage() {
      return !this.hasSession && this.show;
    },

    isLoggedIn() {
      return this.$store.getters['auth/isLoggedIn'];
    },
  },

  async created() {
    if (this.isLoggedIn) {
      try {
        this.$store.dispatch('boxs/setStatus', true);
        this.$store.dispatch('sessions/resetPagePerpage');

        await this.$store.dispatch('sessions/refresh');
        this.show = true;
      } catch {
        this.$store.dispatch('snackbar/showSnackbarErrorLoading', this.$errors.snackbar.sessionList);
      }
    }
  },
};
</script>
