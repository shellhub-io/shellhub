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
            <v-img src="@/assets/logo.png" />
            <span class="overline">BETA</span>
          </v-flex>
        </v-layout>
      </v-container>

      <v-card
        color="transparent"
        class="elevation-0"
      >
        <ValidationObserver
          ref="obs"
          v-slot="{ passes }"
        >
          <v-card-title class="justify-center">
            Forgot your password
          </v-card-title>

          <v-card-text>
            <div class="d-flex align-center justify-center mb-6">
              Please enter the email address you'd like your password reset information send to
            </div>

            <ValidationProvider
              v-slot="{ errors }"
              ref="providerEmail"
              name="Priority"
              vid="email"
              rules="required|email"
            >
              <v-text-field
                v-model="email"
                prepend-icon="email"
                label="Email"
                type="text"
                :error-messages="errors"
                data-test="email-text"
              />
            </ValidationProvider>
          </v-card-text>

          <v-card-actions class="justify-center">
            <v-btn
              type="submit"
              color="primary"
              data-test="login-btn"
              @click="passes(recoverPassword)"
            >
              RESET PASSWORD
            </v-btn>
          </v-card-actions>

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
        </ValidationObserver>
      </v-card>
    </v-flex>
  </v-layout>
</template>

<script>

import {
  ValidationObserver,
  ValidationProvider,
} from 'vee-validate';

export default {
  name: 'ForgotPassword',

  components: {
    ValidationProvider,
    ValidationObserver,
  },

  data() {
    return {
      email: '',
    };
  },

  methods: {
    async recoverPassword() {
      try {
        await this.$store.dispatch('users/recoverPassword', this.email);

        this.$store.dispatch('snackbar/showSnackbarSuccessAction', this.$success.recoverPassword);
      } catch {
        this.$store.dispatch('snackbar/showSnackbarErrorAction', this.$errors.snackbar.recoverPassword);
      }
    },
  },
};

</script>
