<template>
  <v-container>
    <v-card-title
      class="d-flex justify-center"
      data-test="title"
    >
      Account Activation Required
    </v-card-title>

    <v-card-text
      class="d-flex align-center justify-center text-center"
      data-test="subtitle"
      role="status"
      aria-live="polite"
    >
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
        :disabled="!canResend"
        data-test="resend-email-btn"
        text="Resend Email"
        @click="resendEmail()"
      />
    </v-card-actions>

    <v-card-subtitle
      class="d-flex align-center justify-center pa-4 mx-auto"
      data-test="back-to-login-link"
    >
      Back to
      <router-link
        class="ml-1"
        :to="{ name: 'Login' }"
      >
        Login
      </router-link>
    </v-card-subtitle>
  </v-container>
</template>

<script setup lang="ts">
import { computed } from "vue";
import { useRoute, useRouter } from "vue-router";
import handleError from "@/utils/handleError";
import useSnackbar from "@/helpers/snackbar";
import useUsersStore from "@/store/modules/users";

const usersStore = useUsersStore();
const router = useRouter();
const route = useRoute();
const snackbar = useSnackbar();
const canResend = computed(() => typeof route.query.username === "string" && route.query.username.length > 0);

const resendEmail = async () => {
  try {
    await usersStore.resendEmail(route.query.username as string);
    snackbar.showSuccess("The email has been sent.");
    await router.push({ name: "Login" });
  } catch (error) {
    snackbar.showError("An error occurred while sending the email. Please try again.");
    handleError(error);
  }
};
</script>
