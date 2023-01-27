<template>
  <v-app>
    <v-main>
      <v-container class="full-height d-flex justify-center align-center" fluid>
        <v-row align="center" justify="center">
          <v-col cols="12" sm="8" md="4">
            <v-card theme="dark" class="pa-6 bg-v-theme-surface" rounded="lg">
              <v-card-title class="d-flex justify-center align-center mt-4">
                <v-img
                  :src="Logo"
                  max-width="220"
                  alt="logo do ShellHub, uma nuvem de com a escrita ShellHub Admin ao lado"
                />
              </v-card-title>
              <v-container>
                <v-card-title class="text-center"
                >Forgot your password</v-card-title
                >
                <v-card-text class="text-center mt-2">
                  Please insert the e-mail associated to the account you'd like to request an password reset for
                </v-card-text>
                <form @submit.prevent="sendEmail">
                  <v-container>
                    <v-text-field
                      color="primary"
                      prepend-icon="mdi-email"
                      v-model="email"
                      :error-messages="emailError"
                      required
                      label="Email"
                      variant="underlined"
                      data-test="email-text"
                    />
                    <v-card-actions class="justify-center">
                      <v-btn
                        type="submit"
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
                      <router-link class="ml-1" :to="{ name: 'login' }">
                        Login
                      </router-link>
                    </v-card-subtitle>
                  </v-container>
                </form>
              </v-container>
            </v-card>
          </v-col>
        </v-row>
      </v-container>
    </v-main>
  </v-app>
</template>

<script lang="ts">
import { defineComponent } from "vue";
import { useField } from "vee-validate";
import * as yup from "yup";
import Logo from "../assets/logo-inverted.png";
import { useStore } from "../store";
import {
  INotificationsError,
  INotificationsSuccess,
} from "../interfaces/INotifications";

export default defineComponent({
  setup() {
    const store = useStore();

    const { value: email, errorMessage: emailError } = useField<
      string | undefined
    >("name", yup.string().email().required(), { initialValue: "" });

    const sendEmail = () => {
      if (!emailError.value) {
        try {
          store.dispatch("users/recoverPassword", email.value);
          store.dispatch(
            "snackbar/showSnackbarSuccessAction",
            INotificationsSuccess.recoverPassword,
          );
        } catch (error: any) {
          store.dispatch(
            "snackbar/showSnackbarErrorAction",
            INotificationsError.recoverPassword,
          );
          throw new Error(error);
        }
      }
    };

    return {
      Logo,
      email,
      emailError,
      sendEmail,
      store,
    };
  },
});
</script>

<style>
.full-height {
  height: 100vh;
}

.v-field__append-inner {
  cursor: pointer;
}
</style>
