<template>
  <v-container class="text-center">
    <v-slide-y-reverse-transition v-if="wasEmailSent">
      <v-card-text data-test="success-text">
        <strong>An email with password reset instructions has been sent to your registered email address. Please check your inbox.
        </strong>
      </v-card-text>
    </v-slide-y-reverse-transition>

    <div v-else>
      <v-card-title data-test="title-text">
        Forgot your password
      </v-card-title>
      <v-card-text data-test="body-text">
        Please insert the email associated with the account you'd like to request a password reset for
      </v-card-text>
      <v-form
        v-model="isFormValid"
        @submit.prevent="sendEmail"
      >
        <v-text-field
          v-model="account"
          color="primary"
          prepend-inner-icon="mdi-account"
          :error-messages="accountError"
          required
          label="Username or email address"
          variant="underlined"
          class="text-left"
          data-test="account-text"
        />
        <v-card-actions class="pa-0">
          <v-btn
            :disabled="!isFormValid || !account"
            data-test="forgot-password-btn"
            color="primary"
            variant="elevated"
            type="submit"
            block
          >
            RESET PASSWORD
          </v-btn>
        </v-card-actions>
      </v-form>
    </div>
    <v-card-text
      class="pa-4"
      data-test="back-to-login"
    >
      Back to
      <router-link
        class="text-decoration-none"
        :to="{ name: 'Login' }"
      >
        <strong>Login</strong>
      </router-link>
    </v-card-text>
  </v-container>
</template>

<script setup lang="ts">
import { ref } from "vue";
import { useField } from "vee-validate";
import * as yup from "yup";
import handleError from "../utils/handleError";
import useUsersStore from "@/store/modules/users";
import useSnackbar from "@/helpers/snackbar";

const snackbar = useSnackbar();
const usersStore = useUsersStore();
const wasEmailSent = ref(false);
const isFormValid = ref(false);

const { value: account, errorMessage: accountError } = useField<string | undefined>(
  "account",
  yup
    .string()
    .required()
    .min(3)
    .max(255)
    .test("account-error", "The field only accepts the special characters _, ., -, and @.", (value) => {
      const regex = /^[a-zA-Z0-9_.@-\s]*$/;
      return regex.test(value || "");
    })
    .test("white-spaces", "The field cannot contain white spaces.", (value) => {
      const regex = /\s/;
      return !regex.test(value || "");
    }),
  {
    initialValue: "",
  },
);

const sendEmail = async () => {
  if (accountError.value) return;

  try {
    await usersStore.recoverPassword(account.value as string);
    wasEmailSent.value = true;
  } catch (error) {
    snackbar.showError("Failed to send password reset email. Please ensure the email/username is correct and try again.");
    handleError(error);
  }
};
</script>
