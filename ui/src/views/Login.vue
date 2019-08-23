<template>
<v-layout align-center justify-center>
    <v-flex xs12 sm8 md4>
        <v-card class="elevation-12">
            <v-toolbar dark color="primary">
                <v-toolbar-title>Login</v-toolbar-title>
                <v-spacer></v-spacer>
            </v-toolbar>
            <v-card-text>
                <v-form>
                    <v-text-field :error="this.$store.getters['auth/authStatus'] == 'error'" prepend-icon="person" v-model="username" label="Username" type="text"></v-text-field>
                    <v-text-field :error="this.$store.getters['auth/authStatus'] == 'error'" id="password" prepend-icon="lock" v-model="password" label="Password" type="password"></v-text-field>
                </v-form>
            </v-card-text>
            <v-card-actions>
                <v-spacer></v-spacer>
                <v-btn color="primary" @click='login'>Login</v-btn>
            </v-card-actions>
        </v-card>
    </v-flex>
</v-layout>
</template>

<script>
export default {
  data() {
    return {
      username: null,
      password: null
    };
  },

  methods: {
    login() {
      this.$store
        .dispatch("auth/login", {
          username: this.username,
          password: this.password
        })
        .then(() => {
          if (this.$route.query.redirect) {
            this.$router.push(this.$route.query.redirect);
          } else {
            this.$router.push("/");
          }
        });
    }
  }
};
</script>
