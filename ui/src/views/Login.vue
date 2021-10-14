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
        v-if="!showMessage"
        color="transparent"
        class="elevation-0"
      >
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
              label="Password"
              type="password"
              data-test="password-text"
              @keyup.enter="login()"
            />
          </v-card-text>

          <v-card-actions class="justify-center">
            <v-btn
              type="submit"
              color="primary"
              data-test="login-btn"
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
  name: 'Login',

  components: {
    AccountCreated,
  },

  data() {
    return {
      username: '',
      password: '',
      error: false,
      showMessage: false,
    };
  },

  created() {
    this.$store.dispatch('layout/setLayout', 'simpleLayout');

    if (this.$route.query.token) {
      this.$store.dispatch('auth/logout');

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

        this.$store.dispatch('layout/setLayout', 'appLayout');

        if (this.$route.query.redirect) {
          this.$router.push(this.$route.query.redirect);
        } else {
          this.$router.push('/');
        }
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
