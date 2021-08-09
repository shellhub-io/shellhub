<template>
  <v-layout
    class="grey lighten-4"
    align-center
    justify-center
  >
    <div class="text-center">
      <v-overlay :value="overlay">
        <v-progress-circular
          indeterminate
          size="64"
        />
      </v-overlay>
    </div>

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
        v-if="!showMessage"
        color="transparent"
        class="elevation-0"
      >
        <ValidationObserver
          ref="obs"
          v-slot="{ passes }"
        >
          <v-card-title class="justify-center">
            Sign up for free
          </v-card-title>

          <v-card-text>
            <ValidationProvider
              v-slot="{ errors }"
              ref="providerName"
              name="Priority"
              vid="name"
              rules="required"
            >
              <v-text-field
                v-model="newUser.name"
                prepend-icon="person"
                label="Name"
                type="text"
                :error-messages="errors"
                data-test="name-text"
              />
            </ValidationProvider>

            <ValidationProvider
              v-slot="{ errors }"
              ref="providerUsername"
              name="Priority"
              vid="username"
              rules="required"
            >
              <v-text-field
                v-model="newUser.username"
                prepend-icon="person"
                label="Username"
                type="text"
                :error-messages="errors"
                data-test="username-text"
              />
            </ValidationProvider>

            <ValidationProvider
              v-slot="{ errors }"
              ref="providerEmail"
              name="Priority"
              vid="email"
              rules="required|email"
            >
              <v-text-field
                v-model="newUser.email"
                prepend-icon="email"
                label="Email"
                type="text"
                :error-messages="errors"
                data-test="email-text"
              />
            </ValidationProvider>

            <ValidationProvider
              v-slot="{ errors }"
              ref="providerPassword"
              name="Priority"
              rules="required|password|comparePasswords:@currentPassword"
              vid="password"
            >
              <v-text-field
                id="password"
                v-model="newUser.password"
                prepend-icon="lock"
                label="Password"
                type="password"
                :error-messages="errors"
                data-test="password-text"
              />
            </ValidationProvider>

            <ValidationProvider
              v-slot="{ errors }"
              ref="providerConfirmPassword"
              rules="required|confirmed:password"
              name="confirmPassword"
            >
              <v-text-field
                id="confirmpassword"
                v-model="newUser.confirmPassword"
                prepend-icon="lock"
                label="Confirm password"
                type="password"
                :error-messages="errors"
                data-test="confirmPassword-text"
              />
            </ValidationProvider>
          </v-card-text>

          <v-card-actions class="justify-center">
            <v-btn
              type="submit"
              color="primary"
              data-test="login-btn"
              @click="passes(signUp)"
            >
              SIGN UP
            </v-btn>
          </v-card-actions>

          <v-card-subtitle class="d-flex align-center justify-center pa-4 mx-auto">
            Already have an account?

            <router-link
              class="ml-1"
              :to="{ name: 'login' }"
            >
              Log in
            </router-link>
          </v-card-subtitle>
        </ValidationObserver>
      </v-card>

      <AccountCreated
        :show="showMessage"
        :email="newUser.email"
        data-test="accountCreated-component"
      />
    </v-flex>
  </v-layout>
</template>

<script>

import {
  ValidationObserver,
  ValidationProvider,
} from 'vee-validate';

import AccountCreated from '@/components/account/AccountCreated';

export default {
  name: 'SignUp',

  components: {
    ValidationProvider,
    ValidationObserver,
    AccountCreated,
  },

  data() {
    return {
      newUser: {
        name: '',
        username: '',
        email: '',
        password: '',
        confirmPassword: '',
      },
      showMessage: false,
      delay: 500,
      overlay: false,
    };
  },

  watch: {
    overlay() {
      setTimeout(() => {
        this.overlay = false;
        this.$store.dispatch('snackbar/showSnackbarSuccessAction', this.$success.addUser);
      }, this.delay);
    },
  },

  methods: {
    async signUp() {
      try {
        await this.$store.dispatch('users/signUp', this.newUser);
        this.overlay = !this.overlay;
        this.showMessage = !this.showMessage;
      } catch (error) {
        // Invalid username and/or password
        if (error.response.status === 400) {
          this.$refs.obs.setErrors({
            username: ['The username must be alphanumeric'],
          });
        } else if (error.response.status === 409) {
          error.response.data.forEach((n) => {
            if (n.Field === 'username') {
              this.$refs.obs.setErrors({
                username: ['This username is already taken'],
              });
            } else if (n.Field === 'email') {
              this.$refs.obs.setErrors({
                email: ['This email is already taken'],
              });
            } else if (n.Field === 'password') {
              this.$refs.obs.setErrors({
                password: ['This email is invalid'],
              });
            }
          });
        } else {
          this.$store.dispatch('snackbar/showSnackbarErrorAction', this.$errors.snackbar.addUser);
        }
      }
    },
  },
};

</script>
