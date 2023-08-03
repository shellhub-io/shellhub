<template>
  <v-app>
    <snackbar-component />
    <SnackbarNew />
    <v-main class="d-flex justify-center align-center">
      <v-container
        class="full-height d-flex justify-center align-center"
        fluid
      >
        <v-row
          align="center"
          justify="center"
        >
          <v-col
            cols="12"
            sm="8"
            md="4"
            xl="3"
          >
            <v-card
              theme="dark"
              class="pa-6 bg-v-theme-surface"
              rounded="lg"
            >
              <v-card-title class="d-flex justify-center align-center mt-4">
                <v-img
                  :src="Logo"
                  max-width="220"
                  alt="ShellHub logo, a cloud with a shell in your base write ShellHub in the right side"
                />
              </v-card-title>
              <p
                v-if="!envVariables.isEnterprise && !envVariables.isCloud"
                class="text-caption text-center text-md font-weight-bolad"
              >Community Edition
              </p>
              <router-view :key="currentRoute.value.path" />
            </v-card>
          </v-col>
        </v-row>
      </v-container>
    </v-main>
  </v-app>
</template>

<script setup lang="ts">
import { computed } from "vue";
import { useRouter } from "vue-router";
import Logo from "../assets/logo-inverted.png";
import SnackbarNew from "@/components/Snackbar/SnackbarNew.vue";
import { envVariables } from "../envVariables";

const router = useRouter();
const currentRoute = computed(() => router.currentRoute);
</script>
