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
            <v-img src="@/assets/logo.png" />
            <span class="overline">BETA</span>
          </v-flex>
        </v-layout>
      </v-container>

      <v-card class="elevation-12 mt-10">
        <v-toolbar
          dark
          color="primary"
        >
          <v-toolbar-title>Login</v-toolbar-title>
          <v-spacer />
        </v-toolbar>

        <v-form
          @submit.prevent="login()"
          @keyup.native.enter="submit"
        >
          <v-card-text>
            <v-text-field
              v-model="username"
              :error="this.$store.getters['auth/authStatus'] == 'error'"
              prepend-icon="person"
              label="Username"
              type="text"
              data-test="username-text"
            />
            <v-text-field
              id="password"
              v-model="password"
              :error="this.$store.getters['auth/authStatus'] == 'error'"
              prepend-icon="lock"
              label="Password"
              type="password"
              data-test="password-text"
            />
          </v-card-text>

          <v-card-actions>
            <v-spacer />
            <v-btn
              type="submit"
              color="primary"
              data-test="login-btn"
            >
              Submit
            </v-btn>
          </v-card-actions>
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
        this.$router.push('/');
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
      } catch {
        this.$store.dispatch('modals/showSnackbarError', true);
      }
    },
  },
};

</script>
