<template>
  <v-container>
    <v-card-title class="d-flex justify-center" data-test="title">
      Account Activation Required
    </v-card-title>

    <v-card-text class="d-flex align-center justify-center text-center" data-test="subtitle">
      Thank you for registering an account on ShellHub.
      An email was sent with a confirmation link. You need to click on the link to activate your account.
      If you haven't received the email, click on Resend Email button.
    </v-card-text>

    <v-card-actions class="justify-center">
      <v-btn
        type="submit"
        color="primary"
        variant="tonal"
        block
        data-test="resendEmail-btn"
        @click="resendEmail()"
      >
        Resend Email
      </v-btn>
    </v-card-actions>

    <v-card-subtitle
      class="d-flex align-center justify-center pa-4 mx-auto"
      data-test="isCloud-card"
    >
      Back to
      <router-link
        class="ml-1"
        :to="{ name: 'login' }"
      >
        Login
      </router-link>
    </v-card-subtitle>
  </v-container>
</template>

<script setup lang="ts">
import { useRoute, useRouter } from "vue-router";
import { useStore } from "../store";
import {
  INotificationsSuccess,
  INotificationsError,
} from "../interfaces/INotifications";
import handleError from "@/utils/handleError";

const store = useStore();
const router = useRouter();
const route = useRoute();

const resendEmail = async () => {
  try {
    await store.dispatch("users/resendEmail", route.query.username);
    store.dispatch(
      "snackbar/showSnackbarSuccessAction",
      INotificationsSuccess.resendEmail,
    );
    await router.push({ name: "login" });
  } catch (error) {
    store.dispatch(
      "snackbar/showSnackbarErrorDefault",
      INotificationsError.resendEmail,
    );
    handleError(error);
  }
};
</script>
