<template>
  <WindowDialog
    v-model="showDialog"
    transition="dialog-bottom-transition"
    persistent
    title="Multi-Factor Authentication Enabled"
    description="Add a recovery email to secure your account access"
    icon="mdi-email-plus-outline"
    icon-color="primary"
    :show-close-button="false"
  >
    <div class="pa-6 d-flex flex-column align-center">
      <p class="text-justify mb-4 px-1">
        In case you lose access to all your MFA credentials,
        we'll need a recovery email to verify your identity
        and reset your account access.
        To ensure you can recover your account if you lose
        access to your MFA credentials, please associate a
        recovery email.
      </p>
      <v-text-field
        width="400"
        class="mx-auto"
        v-model="recoveryEmail"
        label="Recovery Email"
        :error-messages="recoveryEmailError"
        hide-details="auto"
        required
        data-test="recovery-email-text"
      />

    </div>

    <template #footer>
      <v-spacer />
      <v-card-actions>
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
    </template>
  </WindowDialog>
</template>

<script setup lang="ts">
import * as yup from "yup";
import { useField } from "vee-validate";
import { computed } from "vue";
import axios, { AxiosError } from "axios";
import handleError from "@/utils/handleError";
import useSnackbar from "@/helpers/snackbar";
import WindowDialog from "@/components/Dialogs/WindowDialog.vue";
import useAuthStore from "@/store/modules/auth";
import useUsersStore from "@/store/modules/users";
import { IUserPatch } from "@/interfaces/IUser";

const showDialog = defineModel<boolean>({ required: true });
const authStore = useAuthStore();
const usersStore = useUsersStore();
const snackbar = useSnackbar();
const email = computed(() => authStore.email);
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
    recovery_email: recoveryEmail.value,
  } as IUserPatch;

  try {
    await usersStore.patchData(data);
    authStore.updateUserData(data);
    snackbar.showSuccess("Recovery email updated successfully.");
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

defineExpose({ showDialog, recoveryEmailError });
</script>
