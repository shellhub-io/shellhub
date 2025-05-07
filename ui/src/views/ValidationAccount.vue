<template>
  <v-container>
    <v-card-title class="d-flex justify-center" data-test="verification-title">
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
      data-test="failed-token-cardText"
    >
      Your account activation token has expired. Go to the login page,
      login to receive another email with the activation link.
    </v-card-text>

    <v-card-subtitle
      class="d-flex align-center justify-center pa-4 mx-auto pt-2"
      data-test="isCloud-card"
    >
      Back to
      <router-link
        class="ml-1"
        :to="{ name: 'Login' }"
        data-test="login-btn"
      >
        Login
      </router-link>
    </v-card-subtitle>
  </v-container>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from "vue";
import { useRoute, useRouter } from "vue-router";
import axios from "axios";
import { useStore } from "../store";
import handleError from "@/utils/handleError";
import useSnackbar from "@/helpers/snackbar";

const store = useStore();
const router = useRouter();
const route = useRoute();
const snackbar = useSnackbar();

const activationProcessingStatus = ref("processing");

const verifyActivationProcessingStatus = computed(() => activationProcessingStatus.value);

const validationAccount = async (data: unknown) => {
  try {
    await store.dispatch("users/validationAccount", data);
    snackbar.showSuccess("Your account has been activated successfully.");
    // Only set to "success" if validation is successful
    activationProcessingStatus.value = "success";
    setTimeout(() => router.push({ path: "/login" }), 4000);
  } catch (error: unknown) {
    snackbar.showError("There was a problem activating your account.");
    if (error && axios.isAxiosError(error) && error.response) {
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
    handleError(error);
  }
};

onMounted(() => {
  validationAccount(route.query);
});

defineExpose({ activationProcessingStatus });
</script>
