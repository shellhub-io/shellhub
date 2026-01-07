<template>
  <v-container>
    <v-card-title
      class="d-flex justify-center"
      data-test="verification-title"
    >
      MFA Disabling
    </v-card-title>
    <v-row>
      <v-col data-test="verification-subtitle">
        <h4>Please, paste the codes we've sent on your primary and recovery email</h4>
      </v-col>
    </v-row>
    <v-row class="mb-3">
      <v-col>
        <v-alert
          v-if="verifyDisableProcessingStatus === 'success'"
          class="d-flex align-center justify-center text-center"
          type="success"
          data-test="verification-success"
        >
          You have successfully disabled Multi-Factor Authentication (MFA).
          You will be redirected to the application in 5 seconds.
        </v-alert>

        <v-alert
          v-if="verifyDisableProcessingStatus === 'failed'"
          class="d-flex align-center justify-center text-center"
          type="error"
          data-test="verification-error"
        >
          {{ errorMsg }}
        </v-alert>
      </v-col>
    </v-row>

    <v-text-field
      v-model="primaryEmail"
      class="mt-1"
      label="Primary Email Code"
      :error-messages="emailError"
      required
      variant="outlined"
      data-test="email-text"
    />

    <v-text-field
      v-model="recoveryEmail"
      class="mt-1"
      label="Recovery Email Code"
      :error-messages="recoveryEmailError"
      required
      variant="outlined"
      data-test="recovery-email-text"
    />

    <v-btn
      class="d-flex align-center justify-center pa-4 mx-auto mb-2"
      :disabled="!primaryEmail || !recoveryEmail || !!recoveryEmailError || !!emailError"
      variant="elevated"
      color="primary"
      data-test="save-mail-btn"
      @click="validationAccount"
    >
      Disable MFA
    </v-btn>

    <v-card-subtitle
      class="d-flex align-center justify-center pa-4 mx-auto pt-2"
      data-test="back-to-login"
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
import * as yup from "yup";
import { computed, ref } from "vue";
import { useRoute, useRouter } from "vue-router";
import { useField } from "vee-validate";
import axios, { AxiosError } from "axios";
import handleError from "../utils/handleError";
import useSnackbar from "@/helpers/snackbar";
import useAuthStore from "@/store/modules/auth";

const authStore = useAuthStore();
const router = useRouter();
const route = useRoute();
const snackbar = useSnackbar();
const errorMsg = ref("");

const {
  value: primaryEmail,
  errorMessage: emailError,
} = useField<string>("email", yup.string().required(), {
  initialValue: "",
});

const {
  value: recoveryEmail,
  errorMessage: recoveryEmailError,
} = useField<string>("recoverEmail", yup.string().required(), {
  initialValue: "",
});

const disableProcessingStatus = ref("");

const verifyDisableProcessingStatus = computed(() => disableProcessingStatus.value);

const validationAccount = async () => {
  try {
    await authStore.resetMfa({
      id: route.query.id as string,
      main_email_code: primaryEmail.value,
      recovery_email_code: recoveryEmail.value,
    });

    snackbar.showSuccess("Successfully disabled multi-factor authentication");

    disableProcessingStatus.value = "success";
    setTimeout(() => void router.push({ path: "/" }), 5000);
  } catch (error: unknown) {
    disableProcessingStatus.value = "failed";
    if (axios.isAxiosError(error)) {
      const axiosError = error as AxiosError;
      switch (axiosError.response?.status) {
        case 404:
          errorMsg.value = "The ID sent in your request is invalid, please check the button available in your email and try again.";
          break;
        case 403:
          errorMsg.value = "The recovery codes are incorrect, please check the codes and try again.";
          break;
        default:
          errorMsg.value = "There was a problem disabling your MFA, please try again later.";
          break;
      }
    }
    snackbar.showError("Failed to disable multi-factor authentication");
    handleError(error);
  }
};
</script>
