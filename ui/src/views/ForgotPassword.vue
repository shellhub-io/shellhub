<template>
  <v-container>
    <v-col>
      <v-slide-y-reverse-transition>
        <v-card-text v-if="postSuccessful" class="text-center" data-test="success-text">
          <strong>An email with password reset instructions has been sent to your registered email address. Please check your inbox.
          </strong>
        </v-card-text>
      </v-slide-y-reverse-transition>
    </v-col>

    <v-container>
      <div v-if="!postSuccessful">
        <v-card-title class="text-center" data-test="title-text">
          Forgot your password
        </v-card-title>
        <v-card-text class="text-center" data-test="body-text">
          Please insert the email associated with the account you'd like to request a password reset for
        </v-card-text>
        <v-form v-model="validForm" @submit.prevent="sendEmail">

          <v-text-field
            color="primary"
            prepend-icon="mdi-account"
            v-model="account"
            :error-messages="accountError"
            required
            label="Username or email address"
            variant="underlined"
            data-test="account-text"
          />
          <v-card-actions class="justify-center">
            <v-btn
              :disabled="!validForm"
              data-test="forgotPassword-btn"
              color="primary"
              variant="tonal"
              block
              @click="sendEmail"
            >
              RESET PASSWORD
            </v-btn>
          </v-card-actions>

        </v-form>
      </div>
      <v-card-subtitle
        class="d-flex align-center justify-center pa-4 mx-auto"
        data-test="isCloud-card"
      >
        Back to
        <router-link class="ml-1 text-decoration-none" :to="{ name: 'login' }">
          <strong>Login</strong>
        </router-link>
      </v-card-subtitle>
    </v-container>

  </v-container>
</template>

<script setup lang="ts">
import { ref } from "vue";
import { useField } from "vee-validate";
import * as yup from "yup";
import { useStore } from "../store";
import handleError from "../utils/handleError";

const store = useStore();
const postSuccessful = ref(false);
const validForm = ref(false);

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
  if (accountError.value) {
    return;
  }

  try {
    await store.dispatch("users/recoverPassword", account.value);
    postSuccessful.value = true;
  } catch (error) {
    handleError(error);
  }
};
</script>
