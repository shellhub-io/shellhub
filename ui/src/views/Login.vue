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
    <v-slide-y-reverse-transition>
      <v-alert
        v-model="invalidCredentials"
        type="error"
        closable
        variant="tonal"
        class="mb-4"
      >
        <strong>Invalid login credentials:</strong>
        Your password is incorrect or this account doesn't exists.
      </v-alert>
    </v-slide-y-reverse-transition>

    <v-form
      v-model="validForm"
      @submit.prevent="login"
      data-test="form"
    >
      <v-col>
        <v-text-field
          color="primary"
          prepend-inner-icon="mdi-account"
          v-model="username"
          :rules="rules"
          required
          label="Username or email address"
          variant="underlined"
          data-test="username-text"
        />

        <v-text-field
          color="primary"
          prepend-inner-icon="mdi-lock"
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
        <v-card-actions class="justify-center pa-0">
          <v-btn
            :disabled="!validForm"
            data-test="login-btn"
            color="primary"
            :variant="validForm ? 'elevated' : 'tonal'"
            block
            type="submit"
          >
            LOGIN
          </v-btn>
        </v-card-actions>

      </v-col>
    </v-form>
    <v-col v-if="cloudEnvironment">
      <v-card-subtitle
        class="d-flex align-center justify-center pa-4 mx-auto pt-4 pb-0"
        data-test="forgotPassword-card"
      >
        Did you
        <router-link
          class="ml-1"
          :to="{ name: 'ForgotPassword' }"
        >
          Forgot your Password?
        </router-link>
      </v-card-subtitle>

      <v-card-subtitle
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
import { onMounted, ref, computed } from "vue";
import { useRoute, useRouter } from "vue-router";
import axios, { AxiosError } from "axios";
import { useStore } from "../store";
import isCloudEnvironment from "../utils/cloudUtils";
import handleError from "../utils/handleError";
import useSnackbar from "../helpers/snackbar";

const store = useStore();
const route = useRoute();
const router = useRouter();
const snackbar = useSnackbar();

const showPassword = ref(false);
const loginToken = ref(false);
const username = ref("");
const password = ref("");
const rules = [(v: string) => v ? true : "This is a required field"];
const validForm = ref(false);
const cloudEnvironment = isCloudEnvironment();
const invalidCredentials = ref(false);
const isMfa = computed(() => store.getters["auth/isMfa"]);

onMounted(async () => {
  if (!route.query.token) {
    return;
  }
  loginToken.value = true;
  await store.dispatch("stats/clear");
  await store.dispatch("namespaces/clearNamespaceList");
  await store.dispatch("auth/logout");
  await store.dispatch("auth/loginToken", route.query.token);
});

const login = async () => {
  try {
    await store.dispatch("auth/login", { username: username.value, password: password.value });
    if (isMfa.value === true) {
      router.push({ name: "MfaLogin" });
    } else {
      router.push(route.query.redirect ? route.query.redirect.toString() : "/");
    }
  } catch (error: unknown) {
    if (axios.isAxiosError(error)) {
      const axiosError = error as AxiosError;
      switch (axiosError.response?.status) {
        case 401:
          invalidCredentials.value = true;
          break;
        case 403:
          router.push({ name: "ConfirmAccount", query: { username: username.value } });
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
  validForm,
});
</script>
