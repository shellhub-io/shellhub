<template>
  <v-app>
    <v-main>
      <v-container
        class="full-height d-flex justify-center align-center"
        fluid
      >
        <v-row
          align="center"
          justify="center"
        >
          <v-col
            cols="12"
            sm="8"
            md="4"
          >
            <v-card class="pa-6 bg-v-theme-surface">
              <v-card-title class="d-flex justify-center align-center mt-4">
                <v-img
                  :src="Logo"
                  max-width="220"
                  alt=""
                />
                <span class="mt-6 text-overline">Admin</span>
              </v-card-title>
              <form @submit.prevent="login">
                <v-text-field
                  v-model="username"
                  color="primary"
                  prepend-inner-icon="mdi-account"
                  :error-messages="usernameError"
                  required
                  label="Username"
                  data-test="username-text"
                />

                <v-text-field
                  v-model="password"
                  color="primary"
                  prepend-inner-icon="mdi-lock"
                  :append-inner-icon="showPassword ? 'mdi-eye' : 'mdi-eye-off'"
                  :error-messages="passwordError"
                  label="Password"
                  required
                  data-test="password-text"
                  :type="showPassword ? 'text' : 'password'"
                  @click:append-inner="showPassword = !showPassword"
                />
                <v-card-actions>
                  <v-btn
                    class="w-100 text-center"
                    type="submit"
                    data-test="login-btn"
                    color="primary"
                    variant="elevated"
                    text="Login"
                    @click="login"
                  />
                </v-card-actions>
              </form>
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
import Logo from "../assets/logo-inverted.svg";
import { createNewAdminClient } from "@/api/http";

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
      createNewAdminClient();
      if (route.query.redirect) {
        await router.push(route.query.redirect as string);
      } else {
        await router.push("/");
      }
    } catch {
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
