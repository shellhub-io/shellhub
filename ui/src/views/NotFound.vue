<template>
  <v-container>
    <v-row align="center" justify="center">
      <v-col cols="12" md="6" class="text-center">
        <h1 class="text-h3 font-weight-bold mt-6 mb-4">Whoops!</h1>
        <v-row align="center" justify="center" class="mb-4">
          <v-icon size="110" color="primary">mdi-cloud-off-outline</v-icon>
          <h1 class="text-h1 font-weight-bold mt-4 mb-2 text-primary">404</h1>
        </v-row>
        <p class="mb-5 font-weight-bold text-h3">Page not found</p>
        <p class="mb-4 font-weight-bold text-h6">The requested URL was not found on the server. You can go back to the home
          by clicking the button below.</p>
        <v-btn color="primary" @click="goToHome" data-test="home-btn">Go to Home</v-btn>
      </v-col>
    </v-row>
  </v-container>
</template>

<script setup lang="ts">
import { useRouter } from "vue-router";

const router = useRouter();

const goToHome = () => {
  router.push({ name: "Home" });
};

defineExpose({ goToHome });
</script>
