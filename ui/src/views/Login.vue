<template>
<v-layout align-center justify-center>
  
  <v-flex xs12 sm8 md4>
    
    <div dark color="primary" class="logo">
      <div class="text-center">
        <v-icon>mdi-console</v-icon>
        <h2 style="font-family: monospace">ShellHub</h2>
        <span class="overline">beta</span>
      </div>
    </div>

    <v-card class="elevation-12">
      <v-toolbar dark color="primary">
        <v-toolbar-title>Login</v-toolbar-title>
        <v-spacer></v-spacer>
      </v-toolbar>
      
      <v-form @submit.prevent="login()" >
        <v-card-text>
          <v-text-field :error="this.$store.getters['auth/authStatus'] == 'error'" prepend-icon="person" v-model="username" label="Username" type="text"></v-text-field>
          <v-text-field :error="this.$store.getters['auth/authStatus'] == 'error'" id="password" prepend-icon="lock" v-model="password" label="Password" type="password"></v-text-field>
        </v-card-text>
        
        <v-card-actions>
          <v-spacer></v-spacer>
          <v-btn type="submit" color="primary">Submit</v-btn>
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
      password: null
    };
  },

  methods: {
    login() {
      this.$store
        .dispatch('auth/login', {
          username: this.username,
          password: this.password
        })
        .then(() => {
          if (this.$route.query.redirect) {
            this.$router.push(this.$route.query.redirect);
          } else {
            this.$router.push('/');
          }
        });
    }
  }
};
</script>

<style lang="stylus">
.logo{
  margin-bottom: 15px;
}
</style>
