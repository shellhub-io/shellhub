<template>
  <v-container>
    <v-card-title class="text-center">
      Forgot your password
    </v-card-title>
    <v-card-text class="text-center mt-2">
      Please insert the e-mail associated to the account you'd like
      to request an password reset for
    </v-card-text>
    <form @submit.prevent="sendEmail">
      <v-container>
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
            data-test="login-btn"
            color="primary"
            variant="tonal"
            block
            @click="sendEmail"
          >
            RESET PASSWORD
          </v-btn>
        </v-card-actions>

        <v-card-subtitle
          class="d-flex align-center justify-center pa-4 mx-auto"
          data-test="isCloud-card"
        >
          Back to
          <router-link
            class="ml-1"
            :to="{ name: 'login' }"
          >
            Login
          </router-link>
        </v-card-subtitle>
      </v-container>
    </form>
  </v-container>
</template>

<script lang="ts">
import { defineComponent } from "vue";
import { useField } from "vee-validate";
import * as yup from "yup";
import { useStore } from "../store";
import {
  INotificationsError,
  INotificationsSuccess,
} from "../interfaces/INotifications";
import handleError from "@/utils/handleError";

export default defineComponent({
  setup() {
    const store = useStore();

    const { value: account, errorMessage: accountError } = useField<
      string | undefined
    >(
      "account",
      yup
        .string()
        .required()
        .min(3)
        .max(255)
        .test(
          "account-error",
          "The field only accepts the special characters _, ., - and @.",
          (value) => {
            const regex = /^[a-zA-Z0-9_.@-\s]*$/;
            return regex.test(value || "");
          },
        )
        .test(
          "white-spaces",
          "The field cannot contain white spaces.",
          (value) => {
            const regex = /\s/;
            return !regex.test(value || "");
          },
        ),
      {
        initialValue: "",
      },
    );

    const sendEmail = async () => {
      if (!accountError.value) {
        try {
          await store.dispatch("users/recoverPassword", account.value);
          store.dispatch(
            "snackbar/showSnackbarSuccessAction",
            INotificationsSuccess.recoverPassword,
          );
        } catch (error: unknown) {
          store.dispatch(
            "snackbar/showSnackbarErrorAction",
            INotificationsError.recoverPassword,
          );
          handleError(error);
        }
      }
    };

    return {
      account,
      accountError,
      sendEmail,
      store,
    };
  },
});
</script>
