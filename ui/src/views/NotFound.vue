<template>
  <v-container>
    <v-row align="center" justify="center">
      <v-col cols="12" md="6" class="text-center">
        <h1 class="text-h2 font-weight-bold mt-6 mb-4">Whoops!</h1>
        <v-row align="center" justify="center" class="mb-4">
          <v-icon size="120" color="primary">mdi-bug</v-icon>
          <h1 class="text-h1 font-weight-bold mt-4 mb-2">404</h1>
        </v-row>
        <p class="mb-5 font-weight-bold text-h3">Page not found</p>
        <p class="mb-2 font-weight-bold">The requested URL was not found on the server. You can go back to the dashboard
          by clicking the button below.</p>
        <br>
        <v-btn color="primary" @click="$router.push({ name: 'Dashboard' })">Go to Dashboard</v-btn>
      </v-col>
    </v-row>
  </v-container>
</template>

<script lang="ts">
import { defineComponent } from "vue";

export default defineComponent({
  name: "NotFound",
});
</script>
