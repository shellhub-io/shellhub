<template>
  <v-container>
    <v-row
      align="center"
      justify="center"
    >
      <v-col
        cols="12"
        md="6"
        class="text-center"
      >
        <h1
          class="text-h3 font-weight-bold mt-6 mb-4"
          data-test="whoops-heading"
        >
          Whoops!
        </h1>
        <v-row
          align="center"
          justify="center"
          class="mb-4"
        >
          <v-icon
            size="110"
            color="primary"
            data-test="error-icon"
          >
            mdi-cloud-off-outline
          </v-icon>
          <h1
            class="text-h1 font-weight-bold mt-4 mb-2 text-primary"
            data-test="404-heading"
          >
            404
          </h1>
        </v-row>
        <p
          class="mb-5 font-weight-bold text-h3"
          data-test="not-found-message"
        >
          Page not found
        </p>
        <p
          class="mb-4 font-weight-bold text-h6"
          data-test="help-text"
        >
          The requested URL was not found on the server. You can go back to the home
          by clicking the button below.
        </p>
        <v-btn
          color="primary"
          data-test="home-btn"
          text="Go to Home"
          @click="goToHome"
        />
      </v-col>
    </v-row>
  </v-container>
</template>

<script setup lang="ts">
import { useRouter } from "vue-router";

const router = useRouter();

const goToHome = async () => { await router.push({ name: "Home" }); };
</script>
