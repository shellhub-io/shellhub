<template>
  <v-app>
    <v-main>
      <v-container class="full-height d-flex justify-center align-center" fluid>
        <v-row align="center" justify="center">
          <v-col cols="12" sm="8" md="4">
            <v-card theme="dark" class="pa-6 bg-v-theme-surface" rounded="lg">
              <v-card-title class="d-flex justify-center align-center mt-4">
                <v-img
                  :src="Logo"
                  max-width="220"
                  alt="ShellHub logo, a cloud with a shell in your base write ShellHub in the right side"
                />
              </v-card-title>

              <v-card-title class="d-flex justify-center">
                Verification Account
              </v-card-title>

              <v-card-text
                v-if="verifyActivationProcessingStatus === 'processing'"
                class="d-flex align-center justify-center"
                data-test="processing-cardText"
              >
                Processing activation.
              </v-card-text>

              <v-card-text
                v-if="verifyActivationProcessingStatus === 'success'"
                class="d-flex align-center justify-center text-center"
                data-test="success-cardText"
              >
                Congrats and welcome to ShellHub.
              </v-card-text>

              <v-card-text
                v-if="verifyActivationProcessingStatus === 'failed'"
                class="d-flex align-center justify-center text-center"
                data-test="failed-cardText"
              >
                There was a problem activating your account. Go to the login
                page, login to receive another email with the activation link.
              </v-card-text>

              <v-card-text
                v-if="verifyActivationProcessingStatus === 'failed-token'"
                class="d-flex align-center justify-center text-center"
                data-test="failed-cardText"
              >
                Your account activation token has expired. Go to the login page,
                login to receive another email with the activation link.
              </v-card-text>

              <v-card-subtitle
                class="d-flex align-center justify-center pa-4 mx-auto pt-2"
                data-test="isCloud-card"
              >
                Back to
                <router-link class="ml-1" :to="{ name: 'login' }">
                  Login
                </router-link>
              </v-card-subtitle>
            </v-card>
          </v-col>
        </v-row>
      </v-container>
    </v-main>
  </v-app>
</template>

<script lang="ts">
import { computed, defineComponent, onMounted, ref, watch } from "vue";
import { useRoute, useRouter } from "vue-router";
import Logo from "../assets/logo-inverted.png";
import {
  INotificationsError,
  INotificationsSuccess,
} from "../interfaces/INotifications";
import { useStore } from "../store";

export default defineComponent({
  setup() {
    const store = useStore();
    const router = useRouter();
    const route = useRoute();

    const activationProcessingStatus = ref("processing");

    const verifyActivationProcessingStatus = computed(() => {
      return activationProcessingStatus.value;
    });

    onMounted(() => {
      validationAccount(route.query);
    });

    const validationAccount = async (data: any) => {
      try {
        await store.dispatch("users/validationAccount", data);
        store.dispatch(
          "snackbar/showSnackbarSuccessAction",
          INotificationsSuccess.validationAccount
        );

        activationProcessingStatus.value = "success";
        setTimeout(() => router.push({ path: "/login" }), 4000);
      } catch (error: any) {
        store.dispatch(
          "snackbar/showSnackbarErrorAction",
          INotificationsError.validationAccount
        );
        if (error && error.response) {
          switch (error.response.status) {
            case 400:
              activationProcessingStatus.value = "failed";
              break;
            case 404:
              activationProcessingStatus.value = "failed-token";
              break;
            default:
              activationProcessingStatus.value = "failed";
              break;
          }
        } else {
          activationProcessingStatus.value = "failed";
        }
      }
    };

    return {
      Logo,
      verifyActivationProcessingStatus,
    };
  },
});
</script>

<style>
.full-height {
  height: 100vh;
}

.v-field__append-inner {
  cursor: pointer;
}
</style>
