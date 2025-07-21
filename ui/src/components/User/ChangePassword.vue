<template>
  <BaseDialog v-model="showDialog" @click:outside="close">
    <v-card data-test="password-change-card" class="bg-v-theme-surface">
      <v-card-title class="text-h5 pa-4 bg-primary" data-test="title">
        Change Password
      </v-card-title>

      <v-card-text class="mt-4 mb-3 pb-1">
        <div class="mt-4 pl-4 pr-4">
          <v-text-field
            v-model="currentPassword"
            label="Current password"
            :append-icon="showCurrentPassword ? 'mdi-eye' : 'mdi-eye-off'"
            :type="showCurrentPassword ? 'text' : 'password'"
            class="mb-4"
            variant="underlined"
            :error-messages="currentPasswordError"
            required
            data-test="password-input"
            @click:append="showCurrentPassword = !showCurrentPassword"
          />

          <v-text-field
            v-model="newPassword"
            label="New password"
            :append-icon="showNewPassword ? 'mdi-eye' : 'mdi-eye-off'"
            :type="showNewPassword ? 'text' : 'password'"
            class="mb-4"
            :error-messages="newPasswordError"
            required
            variant="underlined"
            data-test="new-password-input"
            @click:append="showNewPassword = !showNewPassword"
          />

          <v-text-field
            v-model="newPasswordConfirm"
            label="Confirm new password"
            :append-icon="showConfirmPassword ? 'mdi-eye' : 'mdi-eye-off'"
            :type="showConfirmPassword ? 'text' : 'password'"
            class="mb-4"
            variant="underlined"
            :error-messages="newPasswordConfirmError"
            required
            data-test="confirm-new-password-input"
            @click:append="showConfirmPassword = !showConfirmPassword"
          />
        </div>
      </v-card-text>

      <v-card-actions>
        <v-spacer />

        <v-btn variant="text" data-test="close-btn" @click="close">
          Cancel
        </v-btn>

        <v-btn
          color="primary"
          variant="text"
          data-test="change-password-btn"
          :disabled="hasUpdatePasswordError"
          @click="updatePassword()"
        >
          Save Password
        </v-btn>
      </v-card-actions>
    </v-card>
  </BaseDialog>
</template>

<script setup lang="ts">
import * as yup from "yup";
import { useField } from "vee-validate";
import { computed, ref } from "vue";
import axios, { AxiosError } from "axios";
import { useStore } from "@/store";
import handleError from "@/utils/handleError";
import useSnackbar from "@/helpers/snackbar";
import BaseDialog from "../BaseDialog.vue";

const store = useStore();
const snackbar = useSnackbar();
const showDialog = defineModel({ default: false });

const {
  value: currentPassword,
  errorMessage: currentPasswordError,
  resetField: resetCurrentPassword,
} = useField<string>("currentPassword", yup.string().required(), {
  initialValue: "",
});

const {
  value: newPassword,
  errorMessage: newPasswordError,
  setErrors: setNewPasswordError,
  resetField: resetNewPassword,
} = useField<string>(
  "newPassword",
  yup
    .string()
    .required()
    .min(5, "Your password should be 5-32 characters long")
    .max(32, "Your password should be 5-32 characters long"),
  {
    initialValue: "",
  },
);

const {
  value: newPasswordConfirm,
  errorMessage: newPasswordConfirmError,
  setErrors: setNewPasswordConfirmError,
  resetField: resetNewPasswordConfirm,
} = useField<string>(
  "newPasswordConfirm",
  yup
    .string()
    .required()
    .test(
      "passwords-match",
      "Passwords do not match",
      (value) => newPassword.value === value,
    ),
  {
    initialValue: "",
  },
);

const showCurrentPassword = ref(false);
const showNewPassword = ref(false);
const showConfirmPassword = ref(false);
const name = computed(() => store.getters["auth/currentName"]);
const email = computed(() => store.getters["auth/email"]);
const username = computed(() => store.getters["auth/currentUser"]);
const recoveryEmail = computed(() => store.getters["auth/recoveryEmail"]);

const close = () => {
  showDialog.value = false;
  resetCurrentPassword();
  resetNewPassword();
  resetNewPasswordConfirm();
};

const hasUpdatePasswordError = computed(() => (
  Boolean(currentPasswordError.value)
        || Boolean(newPasswordError.value)
        || Boolean(newPasswordConfirmError.value)
        || newPassword.value === ""
        || newPasswordConfirm.value === ""
        || currentPassword.value === ""
));

const updatePassword = async () => {
  if (!hasUpdatePasswordError.value) {
    const data = {
      name: name.value,
      username: username.value,
      email: email.value,
      recovery_email: recoveryEmail.value,
      currentPassword: currentPassword.value,
      newPassword: newPassword.value,
    };

    try {
      await store.dispatch("users/patchPassword", data);
      snackbar.showSuccess("Password updated successfully.");
      close();
    } catch (error: unknown) {
      if (axios.isAxiosError(error)) {
        const axiosError = error as AxiosError;
        if (axiosError.response?.status === 403) {
          setNewPasswordError("Your password doesn't match");
          setNewPasswordConfirmError("Your password doesn't match");
          snackbar.showError("An error occurred while updating the password.");
        }
      } else {
        snackbar.showError("An error occurred while updating the password.");
        handleError(error);
      }
    }
  }
};

defineExpose({ showDialog, newPasswordConfirmError, newPasswordError });
</script>
