<template>
  <v-container>
    <v-alert
      v-if="loginToken"
      data-test="loadingToken-alert"
      class="pa-6 bg-v-theme-surface"
    >
      <div class="text-center">
        <p>Logging the token in...</p>

        <v-progress-linear
          indeterminate
          color="cyan"
          class="mt-2"
        />
      </div>
    </v-alert>
    <form @submit.prevent="login">
      <v-col>
        <v-text-field
          color="primary"
          prepend-icon="mdi-account"
          v-model="username"
          :rules="rules"
          required
          label="Username or email address"
          variant="underlined"
          data-test="username-text"
        />

        <v-text-field
          color="primary"
          prepend-icon="mdi-lock"
          :append-inner-icon="showPassword ? 'mdi-eye' : 'mdi-eye-off'"
          v-model="password"
          :rules="rules"
          label="Password"
          required
          variant="underlined"
          data-test="password-text"
          :type="showPassword ? 'text' : 'password'"
          @click:append-inner="showPassword = !showPassword"
        />
        <v-card-actions class="justify-center">
          <v-btn
            data-test="login-btn"
            color="primary"
            variant="tonal"
            block
            @click="login"
            type="submit"
          >
            LOGIN
          </v-btn>
        </v-card-actions>

      </v-col>
    </form>
    <v-col>
      <v-card-subtitle
        v-if="cloudEnvironment"
        class="d-flex align-center justify-center pa-4 mx-auto pt-4 pb-0"
        data-test="forgotPassword-card"
      >
        Forgot your
        <router-link
          class="ml-1"
          :to="{ name: 'ForgotPassword' }"
        >
          Password?
        </router-link>
      </v-card-subtitle>

      <v-card-subtitle
        v-if="cloudEnvironment"
        class="d-flex align-center justify-center pa-4 mx-auto"
        data-test="isCloud-card"
      >
        Don't have an account?

        <router-link
          class="ml-1"
          :to="{ name: 'SignUp' }"
        >
          Sign up here
        </router-link>
      </v-card-subtitle>
    </v-col>
  </v-container>
</template>
<script setup lang="ts">
import { onMounted, ref } from "vue";
import { useRoute, useRouter } from "vue-router";
import axios, { AxiosError } from "axios";
import { useStore } from "../store";
import isCloudEnvironment from "../utils/cloudUtils";
import handleError from "../utils/handleError";
import useSnackbar from "../helpers/snackbar";

const showPassword = ref(false);
const loginToken = ref(false);
const username = ref("");
const password = ref("");
const rules = [(v: string) => v ? true : "This is a required field"];
const store = useStore();
const route = useRoute();
const router = useRouter();
const snackbar = useSnackbar();
const cloudEnvironment = isCloudEnvironment();

// Test purpose references
const requireAccountConfirm = ref(false);
const invalidCredentials = ref(false);

onMounted(async () => {
  if (!route.query.token) {
    return;
  }

  loginToken.value = true;
  await store.dispatch("stats/clear");
  await store.dispatch("namespaces/clearNamespaceList");
  await store.dispatch("auth/logout");
  await store.dispatch("auth/loginToken", route.query.token);
  await router.push("/");
});

const login = async () => {
  try {
    await store.dispatch("auth/login", { username: username.value, password: password.value });
    router.push(route.query.redirect ? route.query.redirect.toString() : "/");
  } catch (error: unknown) {
    if (axios.isAxiosError(error)) {
      const axiosError = error as AxiosError;
      switch (axiosError.response?.status) {
        case 401:
          snackbar.showError("Invalid username or password. Please try again.");
          invalidCredentials.value = true;
          break;
        case 403:
          await router.push({ name: "ConfirmAccount", query: { username: username.value } });
          requireAccountConfirm.value = true;
          break;
        default:
          snackbar.showError("Something went wrong in our server. Please try again later.");
          handleError(error);
      }
      return;
    }

    snackbar.showError("Something went wrong. Please try again later.");
    handleError(error);
  }
};

defineExpose({
  invalidCredentials,
  requireAccountConfirm,
  router,
});
</script>
