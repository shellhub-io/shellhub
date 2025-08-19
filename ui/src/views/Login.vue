<template>
  <v-container>
    <v-alert
      v-if="alertVisible"
      type="warning"
      variant="tonal"
      class="mb-4"
      data-test="user-status-alert"
    >
      {{ alertMessage }}
    </v-alert>
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
        :title="invalid.title + (invalid.timeout ? countdownTimer : '')"
        :text="invalid.msg"
        @click:close="!invalidCredentials"
        closable
        variant="tonal"
        class="mb-4"
        data-test="invalid-login-alert"
      />
    </v-slide-y-reverse-transition>
    <v-slide-y-reverse-transition>
      <v-alert
        v-model="isCountdownFinished"
        type="success"
        title="Your timeout has finished"
        text="Please try to log back in."
        closable
        variant="tonal"
        class="mb-4"
        data-test="invalid-login-alert"
      />
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
          :disabled="!ssoStatus.local && envVariables.isEnterprise"
          :rules="rules"
          required
          label="Username or email address"
          data-test="username-text"
        />

        <v-text-field
          color="primary"
          prepend-inner-icon="mdi-lock"
          :append-inner-icon="showPassword ? 'mdi-eye' : 'mdi-eye-off'"
          v-model="password"
          :disabled="!ssoStatus.local && envVariables.isEnterprise"
          :rules="rules"
          label="Password"
          required
          data-test="password-text"
          :type="showPassword ? 'text' : 'password'"
          @click:append-inner="showPassword = !showPassword"
        />
        <v-card-actions class="justify-center pa-0">
          <v-btn
            :disabled="!validForm || (!ssoStatus.local && envVariables.isEnterprise)"
            data-test="login-btn"
            color="primary"
            :variant="validForm ? 'elevated' : 'tonal'"
            block
            type="submit"
          >
            Login
          </v-btn>
        </v-card-actions>

      </v-col>
    </v-form>
    <v-col v-if="cloudEnvironment">
      <v-card-subtitle
        class="d-flex align-center justify-center pa-4 mx-auto pt-0 pb-0"
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
    <div v-if="ssoStatus.saml && envVariables.isEnterprise" data-test="or-divider-sso">
      <v-row class="mb-2">
        <v-col class="mr-1">
          <v-divider />
        </v-col>
        <v-card-subtitle>OR</v-card-subtitle>
        <v-col class="ml-1">
          <v-divider />
        </v-col>
      </v-row>
      <v-col
        class="d-flex align-center justify-center"
      >
        <v-btn
          @click="redirectToSaml()"
          color="primary"
          class="bg-primary"
          prepend-icon="mdi-cloud-sync-outline"
          size="large"
          data-test="sso-btn"
        >Login with SSO</v-btn>
      </v-col>
    </div>
  </v-container>
</template>
<script setup lang="ts">
import { onMounted, ref, computed, reactive, watch } from "vue";
import { useRoute, useRouter } from "vue-router";
import axios, { AxiosError } from "axios";
import { useStore } from "../store";
import isCloudEnvironment from "../utils/cloudUtils";
import handleError from "../utils/handleError";
import useSnackbar from "../helpers/snackbar";
import useCountdown from "@/utils/countdownTimeout";
import { envVariables } from "@/envVariables";
import useAuthStore from "@/store/modules/auth";
import useNamespacesStore from "@/store/modules/namespaces";

const store = useStore();
const route = useRoute();
const router = useRouter();
const snackbar = useSnackbar();
const authStore = useAuthStore();
const namespacesStore = useNamespacesStore();
const showPassword = ref(false);
const loginToken = ref(false);
const invalid = reactive({ title: "", msg: "", timeout: false });
const username = ref("");
const password = ref("");
const rules = [(v: string) => v ? true : "This is a required field"];
const validForm = ref(false);
const cloudEnvironment = isCloudEnvironment();
const invalidCredentials = ref(false);
const isCountdownFinished = ref(false);
const isMfaEnabled = computed(() => authStore.isMfaEnabled);
const loginTimeout = computed(() => authStore.loginTimeout);
const isLoggedIn = computed(() => authStore.isLoggedIn);
const ssoStatus = computed(() => store.getters["users/getSystemInfo"].authentication);
const samlUrl = computed(() => store.getters["users/getSamlURL"]);
// Alerts for user status on accept namespace invitation logic
const userStatus = computed(() => namespacesStore.userStatus);

const cameFromAcceptInvite = computed(() => isLoggedIn.value === false && route.query.redirect?.includes("/accept-invite"));

const missingAssertions = route.query.missing_assertions;

const alertMessage = computed(() => {
  if (userStatus.value === "not-confirmed") {
    return "Your account is not confirmed, please confirm it before attempting to accept the namespace invite.";
  }
  if (cameFromAcceptInvite.value) {
    return "Please login before accepting any namespace invitation.";
  }
  if (missingAssertions) {
    return "The SSO configuration is incomplete due to missing required mappings. Please contact your administrator to resolve this issue.";
  }
  return "";
});

const alertVisible = computed(() => userStatus.value === "not-confirmed" || cameFromAcceptInvite.value || missingAssertions);

// Logic for wrong login countdown
const { startCountdown, countdown } = useCountdown();

const countdownTimer = ref("");

watch(countdown, (newValue) => {
  countdownTimer.value = newValue;
  if (countdownTimer.value === "0 seconds") {
    invalidCredentials.value = false;
    isCountdownFinished.value = true;
  }
});

const redirectToSaml = async () => {
  await store.dispatch("users/fetchSamlUrl");
  window.location.replace(samlUrl.value);
};

onMounted(async () => {
  if (!route.query.token) {
    return;
  }

  loginToken.value = true;

  await store.dispatch("stats/clear");
  namespacesStore.namespaceList = [];
  authStore.logout();
  await authStore.loginWithToken(route.query.token as string);

  window.location.href = "/";
});

const login = async () => {
  try {
    await authStore.login({ username: username.value, password: password.value });

    const redirectPath = route.query.redirect ? route.query.redirect.toString() : "/";
    // eslint-disable-next-line @typescript-eslint/no-unused-vars
    const { redirect, ...cleanedQuery } = route.query;

    if (isMfaEnabled.value === true) {
      await router.push({ name: "MfaLogin" });
      localStorage.setItem("name", username.value);
    } else {
      await router.push({ path: redirectPath, query: cleanedQuery });
    }
  } catch (error: unknown) {
    isCountdownFinished.value = false;
    if (axios.isAxiosError(error)) {
      const axiosError = error as AxiosError;
      switch (axiosError.response?.status) {
        case 401:
          invalidCredentials.value = true;
          Object.assign(invalid, {
            title: "Invalid login credentials",
            msg: "Your password is incorrect or this account doesn't exist.",
            timeout: false,
          });
          break;
        case 403:
          router.push({ name: "ConfirmAccount", query: { username: username.value } });
          break;
        case 429:
          startCountdown(loginTimeout.value);
          invalidCredentials.value = true;
          Object.assign(invalid, {
            title: "Your account is blocked for ",
            msg: "There was too many failed login attempts. Please wait to try again.",
            timeout: true,
          });
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
  ssoStatus,
});
</script>
