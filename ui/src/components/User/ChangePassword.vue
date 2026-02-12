<template>
  <FormDialog
    v-model="showDialog"
    title="Change Password"
    icon="mdi-lock"
    confirm-text="Save Password"
    cancel-text="Cancel"
    :confirm-disabled="hasUpdatePasswordError"
    confirm-data-test="change-password-btn"
    cancel-data-test="close-btn"
    data-test="password-change-dialog"
    @close="close"
    @cancel="close"
    @confirm="updatePassword"
  >
    <div class="px-6 pt-4">
      <v-text-field
        v-model="currentPassword"
        label="Current password"
        :append-icon="showCurrentPassword ? 'mdi-eye' : 'mdi-eye-off'"
        :type="showCurrentPassword ? 'text' : 'password'"
        class="mb-4"
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
        data-test="new-password-input"
        @update:model-value="handleNewPasswordChange"
        @click:append="showNewPassword = !showNewPassword"
      />

      <v-text-field
        v-model="newPasswordConfirm"
        label="Confirm new password"
        :append-icon="showConfirmPassword ? 'mdi-eye' : 'mdi-eye-off'"
        :type="showConfirmPassword ? 'text' : 'password'"
        class="mb-4"
        :error-messages="newPasswordConfirmError"
        required
        data-test="confirm-new-password-input"
        @update:model-value="handleNewPasswordChange"
        @click:append="showConfirmPassword = !showConfirmPassword"
      />
    </div>
  </FormDialog>
</template>

<script setup lang="ts">
import * as yup from "yup";
import { useField } from "vee-validate";
import { computed, ref } from "vue";
import axios from "axios";
import FormDialog from "@/components/Dialogs/FormDialog.vue";
import handleError from "@/utils/handleError";
import useSnackbar from "@/helpers/snackbar";
import useAuthStore from "@/store/modules/auth";
import useUsersStore from "@/store/modules/users";

const authStore = useAuthStore();
const usersStore = useUsersStore();
const snackbar = useSnackbar();
const showDialog = defineModel<boolean>({ required: true });

const {
  value: currentPassword,
  errorMessage: currentPasswordError,
  setErrors: setCurrentPasswordError,
  resetField: resetCurrentPassword,
} = useField<string>("currentPassword", yup.string().required("This field is required"), {
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
    .required("This field is required")
    .min(5, "Your password should be 5-32 characters long")
    .max(32, "Your password should be 5-32 characters long"),
  { initialValue: "" },
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
    .required("This field is required")
    .test("passwords-match", "Passwords do not match", (value) => newPassword.value === value),
  { initialValue: "" },
);

const showCurrentPassword = ref(false);
const showNewPassword = ref(false);
const showConfirmPassword = ref(false);
const { name, email, username, recoveryEmail } = authStore;

const handleNewPasswordChange = () => {
  if (!newPasswordConfirm.value || !newPassword.value) return;

  if (newPassword.value !== newPasswordConfirm.value) {
    setNewPasswordConfirmError("Passwords do not match");
    return;
  }

  setNewPasswordConfirmError("");
};

const close = () => {
  showDialog.value = false;
  resetCurrentPassword();
  resetNewPassword();
  resetNewPasswordConfirm();
};

const hasUpdatePasswordError = computed(() => !!currentPasswordError.value
  || !!newPasswordError.value
  || !!newPasswordConfirmError.value
  || !currentPassword.value
  || !newPassword.value
  || !newPasswordConfirm.value);

const updatePassword = async () => {
  if (hasUpdatePasswordError.value) return;

  const data = {
    name,
    username,
    email,
    recovery_email: recoveryEmail,
    currentPassword: currentPassword.value,
    newPassword: newPassword.value,
  };

  try {
    await usersStore.patchPassword(data);
    snackbar.showSuccess("Password updated successfully.");
    close();
  } catch (error: unknown) {
    if (axios.isAxiosError(error)) {
      if (error.response?.status === 403) {
        setNewPasswordError("Your password doesn't match");
        setNewPasswordConfirmError("Your password doesn't match");
      } else if (error.response?.status === 400) {
        setCurrentPasswordError("Your current password is incorrect");
      }
    }
    snackbar.showError("An error occurred while updating the password.");
    handleError(error);
  }
};
</script>
