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
      <v-card
        v-if="showMessage && !$env.isCloud"
        data-test="unknownReason-card"
      >
        <v-card-text>
          <v-card-title class="justify-center">
            Activate Account
          </v-card-title>

          <div class="d-flex align-center justify-center mb-6">
            The account is not active for an unknown reason.
          </div>
        </v-card-text>
      </v-card>

      <v-card
        v-if="!showMessage"
        class="pa-6"
      >
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

        <v-form
          @submit.prevent="login()"
        >
          <v-card-text class="pb-0">
            <v-text-field
              v-model="username"
              :error="$store.getters['auth/authStatus'] == 'error'"
              prepend-icon="person"
              label="Username"
              type="text"
              data-test="username-text"
            />
            <v-text-field
              id="password"
              v-model="password"
              :error="$store.getters['auth/authStatus'] == 'error'"
              prepend-icon="lock"
              :append-icon="showPassword? 'mdi-eye': 'mdi-eye-off'"
              label="Password"
              :type="showPassword ? 'text': 'password'"
              data-test="password-text"
              @click:append="showPassword = !showPassword"
              @keyup.enter="login()"
            />
          </v-card-text>

          <v-card-actions class="justify-center">
            <v-btn
              type="submit"
              color="primary"
              data-test="login-btn"
              block
            >
              LOGIN
            </v-btn>
          </v-card-actions>

          <v-card-subtitle
            v-if="$env.isCloud"
            class="d-flex align-center justify-center pa-4 mx-auto pt-8 pb-0"
            data-test="forgotPassword-card"
          >
            Forgot your
            <router-link
              class="ml-1"
              :to="{ name: 'forgotPassword' }"
            >
              Password?
            </router-link>
          </v-card-subtitle>

          <v-card-subtitle
            v-if="$env.isCloud"
            class="d-flex align-center justify-center pa-4 mx-auto"
            data-test="isCloud-card"
          >
            Don't have an account?

            <router-link
              class="ml-1"
              :to="{ name: 'signUp' }"
            >
              Sign up here
            </router-link>
          </v-card-subtitle>
        </v-form>
      </v-card>

      <AccountCreated
        v-if="$env.isCloud"
        :show="showMessage"
        :username="username"
        data-test="accountCreated-component"
      />
    </v-flex>
  </v-layout>
</template>

<script>

import AccountCreated from '@/components/account/AccountCreated';

export default {
  name: 'LoginView',

  components: {
    AccountCreated,
  },

  data() {
    return {
      username: '',
      password: '',
      error: false,
      showPassword: false,
      showMessage: false,
    };
  },

  computed: {
    getStatusDarkMode() {
      return this.$store.getters['layout/getStatusDarkMode'];
    },
  },

  async created() {
    if (this.$route.query.token) {
      this.$store.dispatch('layout/setLayout', 'simpleLayout');

      await this.$store.dispatch('auth/logout');
      await this.$store.dispatch('auth/loginToken', this.$route.query.token);

      this.$store.dispatch('auth/loginToken', this.$route.query.token).then(() => {
        this.$store.dispatch('layout/setLayout', 'appLayout');
        this.$router.push({ name: 'dashboard' }).catch(() => {});
      });
    }
  },

  methods: {
    async login() {
      try {
        await this.$store
          .dispatch('auth/login', {
            username: this.username,
            password: this.password,
          });

        if (this.$route.query.redirect) {
          await this.$router.push(this.$route.query.redirect);
        } else {
          await this.$router.push('/');
        }
        this.$store.dispatch('layout/setLayout', 'appLayout');
      } catch (error) {
        switch (true) {
        case (error.response.status === 401): {
          this.$store.dispatch('snackbar/showSnackbarErrorIncorrect', this.$errors.snackbar.loginFailed);
          break;
        }
        case (error.response.status === 403): {
          this.showMessage = !this.showMessage;
          break;
        }
        default: {
          this.$store.dispatch('snackbar/showSnackbarErrorDefault');
        }
        }
      }
    },
  },
};

</script>
