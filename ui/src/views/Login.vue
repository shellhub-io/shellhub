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

        <v-form @submit.prevent="login()">
          <v-card-text>
            <v-text-field
              v-model="username"
              :error="this.$store.getters['auth/authStatus'] == 'error'"
              prepend-icon="person"
              label="Username"
              type="text"
              data-cy="username-text"
            />
            <v-text-field
              id="password"
              v-model="password"
              :error="this.$store.getters['auth/authStatus'] == 'error'"
              prepend-icon="lock"
              label="Password"
              type="password"
              data-cy="password-text"
            />
          </v-card-text>

          <v-card-actions>
            <v-spacer />
            <v-btn
              type="submit"
              color="primary"
              data-cy="login-btn"
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
    login() {
      this.$store
        .dispatch('auth/login', {
          username: this.username,
          password: this.password,
        })
        .then(() => {
          if (this.$route.query.redirect) {
            this.$router.push(this.$route.query.redirect);
          } else {
            this.$router.push('/');
          }
        });
    },
  },
};

</script>
