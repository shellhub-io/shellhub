<template>
  <v-container>
    <v-row>
      <v-col align="center">
        <h3 data-test="title">Multi-factor Authentication</h3>
      </v-col>
    </v-row>
    <v-row class="mb-2">
      <v-col align="center">
        <h4 data-test="sub-title">An email has been sent to the primary and recovery mail.
          Please check your inbox and click the link we've provided to disable the MFA.</h4>
      </v-col>
    </v-row>
    <v-card-subtitle
      class="d-flex align-center justify-center pa-4 mx-auto pt-2"
      data-test="back-to-login"
    >
      Back to
      <router-link
        class="ml-1"
        :to="{ name: 'login' }"
        data-test="login-btn"
      >
        Login
      </router-link>
    </v-card-subtitle>
  </v-container>
</template>

<script setup lang="ts">
</script>
