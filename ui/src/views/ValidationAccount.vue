<template>
  <v-layout
    class="grey lighten-4"
    align-center
    justify-center
  >
    <v-flex
      xs12
      sm8
      md4
    >
      <v-container>
        <v-layout
          align-center
          justify-center
          column
        >
          <v-flex class="text-center primary--text">
            <v-img
              src="@/assets/logo.png"
            />
            <span class="overline">BETA</span>
          </v-flex>
        </v-layout>
      </v-container>

      <v-card
        color="transparent"
        class="elevation-0"
        data-test="accountCreated-card"
      >
        <v-card-title class="justify-center">
          Verification Account
        </v-card-title>

        <v-card-text class="d-flex align-center justify-center">
          Congrats and welcome to ShellHub.
        </v-card-text>

        <v-card-subtitle
          class="d-flex align-center justify-center pa-4 mx-auto pt-2"
          data-test="isCloud-card"
        >
          Back to
          <router-link
            class="ml-1"
            :to="{ name: 'login' }"
          >
            Login
          </router-link>
        </v-card-subtitle>
      </v-card>
    </v-flex>
  </v-layout>
</template>

<script>

export default {
  name: 'AccountCreatedView',

  created() {
    this.validationAccount(this.$route.query);
  },

  methods: {
    async validationAccount(data) {
      try {
        await this.$store.dispatch('users/validationAccount', data);
        this.$store.dispatch('snackbar/showSnackbarSuccessAction', this.$success.validationAccount);

        setTimeout(() => this.$router.push({ path: '/login' }), 4000);
      } catch {
        this.$store.dispatch('snackbar/showSnackbarErrorAction', this.$errors.snackbar.validationAccount);
      }
    },
  },
};

</script>
