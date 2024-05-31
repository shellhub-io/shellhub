<template>
  <v-dialog
    v-model="dialog"
    transition="dialog-bottom-transition"
    width="700"
    persistent
  >
    <v-card class="bg-v-theme-surface" data-test="card-dialog">
      <v-card-title class="text-h5 pa-4 bg-primary" data-test="dialog-title"> Multi-Factor Authentication Enabled </v-card-title>
      <v-container data-test="dialog-text">
        <v-row class="mb-2">
          <v-col>
            <h4>
              In case you lose access to all your MFA credentials,
              we'll need a recovery email to verify your identity
              and reset your account access.
            </h4>
            <p class="mt-2">
              To ensure you can recover your account if you lose access to your MFA credentials, please associate a recovery email.
            </p>
          </v-col>
        </v-row>
        <v-row>
          <v-col align="center">
            <v-text-field
              width="400"
              v-model="recoveryEmail"
              label="Recovery Email"
              :error-messages="recoveryEmailError"
              required
              variant="underlined"
              data-test="recovery-email-text"
            />
          </v-col>
        </v-row>
        <v-card-actions>
          <v-spacer />
          <v-btn
            :disabled="!recoveryEmail || !!recoveryEmailError"
            variant="text"
            color="primary"
            data-test="save-btn"
            @click="updateUserData"
          >
            Save Recovery Email
          </v-btn>
        </v-card-actions>
      </v-container>
    </v-card>
  </v-dialog>
</template>

<script setup lang="ts">
import * as yup from "yup";
import { useField } from "vee-validate";
import { computed, ref } from "vue";
import axios, { AxiosError } from "axios";
import { useStore } from "@/store";
import { INotificationsSuccess } from "@/interfaces/INotifications";
import handleError from "@/utils/handleError";

const dialog = ref(false);
const store = useStore();
const email = computed(() => store.getters["auth/email"]);
const {
  value: recoveryEmail,
  errorMessage: recoveryEmailError,
  setErrors: setRecoveryEmailError,
} = useField<string>(
  "recoveryEmail",
  yup
    .string()
    .email("Please enter a valid email address")
    .required("Recovery email is required")
    .test(
      "not-same-as-email",
      "Recovery email must not be the same as your current email",
      (value) => value !== email.value,
    ),
  {
    initialValue: "",
  },
);

const updateUserData = async () => {
  const data = {
    id: store.getters["auth/id"],
    recovery_email: recoveryEmail.value,
  };

  try {
    await store.dispatch("users/patchData", data);
    store.dispatch("auth/changeUserData", data);
    store.dispatch(
      "snackbar/showSnackbarSuccessAction",
      INotificationsSuccess.profileData,
    );
  } catch (error) {
    if (axios.isAxiosError(error)) {
      const axiosError = error as AxiosError;
      switch (axiosError.response?.status) {
        case 409:
          setRecoveryEmailError("This recovery email is already in use");
          break;
        case 400:
          setRecoveryEmailError("This recovery email is invalid");
          break;
        default:
          handleError(error);
      }
    }
  }
};

defineExpose({ dialog, recoveryEmailError });
</script>
