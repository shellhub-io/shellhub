<template>
  <v-app>
    <v-main>
      <v-container class="full-height d-flex justify-center align-center" fluid>
        <v-row align="center" justify="center">
          <v-col cols="12" sm="8" md="4">
            <v-card theme="dark" class="pa-6" rounded="lg">
              <v-card-title class="d-flex justify-center align-center mt-4">
                <v-img
                  :src="Logo"
                  max-width="220"
                  alt="logo do ShellHub, uma nuvem de com a escrita ShellHub Admin ao lado"
                />
                <span class="mt-6 text-overline">Admin</span>
              </v-card-title>
              <v-card-item>
                <SnackbarComponent />
                <form @submit.prevent="login">
                  <v-container>
                    <v-text-field
                      color="primary"
                      prepend-icon="mdi-account"
                      v-model="username"
                      :error-messages="usernameError"
                      required
                      label="Username"
                      variant="underlined"
                      data-test="username-text"
                    />

                    <v-text-field
                      color="primary"
                      prepend-icon="mdi-lock"
                      :append-inner-icon="showPassword ? 'mdi-eye' : 'mdi-eye-off'"
                      v-model="password"
                      :error-messages="passwordError"
                      label="Password"
                      required
                      variant="underlined"
                      data-test="password-text"
                      :type="showPassword ? 'text' : 'password'"
                      @click:append-inner="showPassword = !showPassword"
                    />
                    <v-card-actions class="justify-center">
                      <v-btn
                        type="submit"
                        data-test="login-btn"
                        color="primary"
                        variant="tonal"
                        block
                        @click="login"
                      >
                        LOGIN
                      </v-btn>
                    </v-card-actions>
                  </v-container>
                </form>
              </v-card-item>
            </v-card>
          </v-col>
        </v-row>
      </v-container>
    </v-main>
  </v-app>
</template>

<script setup lang="ts">
import { ref } from "vue";
import { useField } from "vee-validate";
import * as yup from "yup";
import { useRoute, useRouter } from "vue-router";
import useAuthStore from "@admin/store/modules/auth";
import useSnackbar from "@/helpers/snackbar";
import Logo from "../assets/logo-inverted.png";
import { createNewClient } from "../api/http";

const showPassword = ref(false);
const snackbar = useSnackbar();
const authStore = useAuthStore();
const route = useRoute();
const router = useRouter();

const { value: username, errorMessage: usernameError } = useField<string | undefined>(
  "name",
  yup.string().required(),
  { initialValue: "" },
);

const { value: password, errorMessage: passwordError } = useField<string | undefined>(
  "password",
  yup.string().required(),
  { initialValue: "" },
);

const hasErrors = () => {
  if (usernameError.value || passwordError.value) {
    return true;
  }

  return false;
};

const login = async () => {
  if (!hasErrors() && username.value && password.value) {
    try {
      await authStore.login({
        username: username.value,
        password: password.value,
      });
      createNewClient();
      if (route.query.redirect) {
        router.push(`${route.query.redirect}`);
      } else {
        router.push("/");
      }
    } catch (error) {
      snackbar.showError("Failed to log in. Please check your credentials and try again.");
    }
  } else {
    snackbar.showError("Failed to log in. Please check your credentials and try again.");
  }
};

defineExpose({ username, password, usernameError, passwordError });
</script>

<style>
.full-height {
  height: 100vh;
}

.v-field__append-inner {
  cursor: pointer;
}
</style>
