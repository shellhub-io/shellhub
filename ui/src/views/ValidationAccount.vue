<template>
  <v-layout
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

        <v-card-text
          v-if="verifyActivationProcessingStatus === 'processing'"
          class="d-flex align-center justify-center"
          data-test="processing-cardText"
        >
          Processing activation.
        </v-card-text>

        <v-card-text
          v-if="verifyActivationProcessingStatus === 'success'"
          class="d-flex align-center justify-center"
          data-test="success-cardText"
        >
          Congrats and welcome to ShellHub.
        </v-card-text>

        <v-card-text
          v-if="verifyActivationProcessingStatus === 'failed'"
          class="d-flex align-center justify-center"
          data-test="failed-cardText"
        >
          There was a problem activating your account. Go to the login page, login to receive
          another email with the activation link.
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

  data() {
    return {
      activationProcessingStatus: 'processing',
    };
  },

  computed: {
    verifyActivationProcessingStatus() {
      return this.activationProcessingStatus;
    },
  },

  created() {
    this.validationAccount(this.$route.query);
  },

  methods: {
    async validationAccount(data) {
      try {
        await this.$store.dispatch('users/validationAccount', data);
        this.$store.dispatch('snackbar/showSnackbarSuccessAction', this.$success.validationAccount);

        this.activationProcessingStatus = 'success';
        setTimeout(() => this.$router.push({ path: '/login' }), 4000);
      } catch {
        this.activationProcessingStatus = 'failed';
        this.$store.dispatch('snackbar/showSnackbarErrorAction', this.$errors.snackbar.validationAccount);
      }
    },
  },
};

</script>
