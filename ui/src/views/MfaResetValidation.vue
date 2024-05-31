<template>
  <v-container>
    <v-card-title class="d-flex justify-center" data-test="verification-title">
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
      class="mt-1"
      v-model="primaryEmail"
      label="Primary Email Code"
      :error-messages="emailError"
      required
      variant="outlined"
      data-test="email-text"
    />

    <v-text-field
      class="mt-1"
      v-model="recoveryEmail"
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
        :to="{ name: 'login' }"
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
import {
  INotificationsError,
  INotificationsSuccess,
} from "../interfaces/INotifications";
import { useStore } from "../store";
import handleError from "../utils/handleError";

const store = useStore();
const router = useRouter();
const route = useRoute();
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
    await store.dispatch("auth/resetMfa", {
      id: route.query.id,
      main_email_code: primaryEmail.value,
      recovery_email_code: recoveryEmail.value });
    store.dispatch(
      "snackbar/showSnackbarSuccessAction",
      INotificationsSuccess.cancelMfa,
    );

    disableProcessingStatus.value = "success";
    setTimeout(() => router.push({ path: "/" }), 5000);
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
    store.dispatch(
      "snackbar/showSnackbarErrorAction",
      INotificationsError.cancelMfa,
    );
    handleError(error);
  }
};

defineExpose({ disableProcessingStatus });
</script>
