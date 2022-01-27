<template>
  <v-layout
    align-center
    justify-center
  >
    <v-flex
      xs12
      sm8
      md4
      lg3
      xl2
    >
      <v-card class="pa-6">
        <v-container>
          <v-layout
            align-center
            justify-center
            column
          >
            <v-flex class="text-center primary--text">
              <v-img
                v-if="getStatusDarkMode"
                src="@/assets/logo-inverted.png"
                max-width="220"
              />

              <v-img
                v-else
                src="@/assets/logo.png"
                max-width="220"
              />
            </v-flex>
          </v-layout>
        </v-container>

        <v-card-title class="justify-center">
          Reset your password
        </v-card-title>

        <ValidationObserver
          ref="obs"
          v-slot="{ passes }"
        >
          <v-card-text>
            <div class="d-flex align-center justify-center mb-6">
              Please enter your new password you would like.
            </div>

            <ValidationProvider
              v-slot="{ errors }"
              ref="providerNewPassword"
              name="Priority"
              rules="required|password|comparePasswords:@currentPassword"
              vid="newPassword"
            >
              <v-text-field
                v-model="newPassword"
                type="password"
                label="New password"
                class="mb-4"
                :error-messages="errors"
                required
                data-test="newPassword-text"
              />
            </ValidationProvider>

            <ValidationProvider
              v-slot="{ errors }"
              ref="providerConfirmPassword"
              rules="required|confirmed:newPassword"
              name="confirm"
            >
              <v-text-field
                v-model="newPasswordConfirm"
                label="Confirm new password"
                type="password"
                class="mb-4"
                :error-messages="errors"
                required
                data-test="confirmNewPassword-text"
              />
            </ValidationProvider>
          </v-card-text>

          <v-card-actions class="justify-center">
            <v-btn
              type="submit"
              color="primary"
              data-test="login-btn"
              @click="passes(updatePassword)"
            >
              UPDATE PASSWORD
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
  name: 'ForgotPasswordView',

  components: {
    ValidationProvider,
    ValidationObserver,
  },

  data() {
    return {
      newPassword: '',
      newPasswordConfirm: '',
      data: {},
    };
  },

  computed: {
    getStatusDarkMode() {
      return this.$store.getters['layout/getStatusDarkMode'];
    },
  },

  async created() {
    this.data = {
      id: this.$route.query.id,
      token: this.$route.query.token,
    };
  },

  methods: {
    async updatePassword() {
      try {
        this.data.password = this.newPassword;
        await this.$store.dispatch('users/updatePassword', this.data);

        this.$store.dispatch('snackbar/showSnackbarSuccessAction', this.$success.updatingAccount);
      } catch {
        this.$store.dispatch('snackbar/showSnackbarErrorAction', this.$errors.snackbar.updatingAccount);
      }
    },
  },
};

</script>
