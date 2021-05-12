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
        <v-form
          @submit.prevent="login()"
        >
          <v-card-text>
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
            class="d-flex align-center justify-center pa-4 mx-auto"
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
    </v-flex>
  </v-layout>
</template>

<script>

export default {
  name: 'Login',

  data() {
    return {
      username: null,
      password: null,
      error: false,
    };
  },

  created() {
    if (this.$route.query.token) {
      this.$store.dispatch('auth/logout');
      this.$store.dispatch('auth/loginToken', this.$route.query.token).then(() => {
        this.$store.dispatch('notifications/fetch');
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
          this.$router.push(this.$route.query.redirect);
        } else {
          this.$router.push('/');
        }
      } catch (error) {
        switch (true) {
        case (error.response.status === 401): {
          this.$store.dispatch('snackbar/showSnackbarErrorIncorrect', this.$errors.loginFailed);
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
