<template>
  <v-container class="pb-0 mb-0">
    <v-alert
      v-if="alertVisible"
      type="warning"
      variant="tonal"
      class="mb-4"
      data-test="user-status-alert"
    >
      Please create your account before accepting the namespace invitation.
    </v-alert>
    <form @submit.prevent="createAccount" v-if="!showMessage">
      <v-card-title class="text-center">Create Account</v-card-title>
      <v-container>
        <v-text-field
          color="primary"
          prepend-inner-icon="mdi-account"
          v-model="name"
          :error-messages="nameError"
          required
          label="Name"
          variant="underlined"
          data-test="name-text"
        />

        <v-text-field
          color="primary"
          prepend-inner-icon="mdi-account"
          v-model="username"
          :error-messages="usernameError"
          required
          label="Username"
          variant="underlined"
          data-test="username-text"
        />

        <v-text-field
          color="primary"
          prepend-inner-icon="mdi-email"
          v-model="email"
          :disabled="isEmailLocked"
          :error-messages="emailError"
          required
          label="Email"
          variant="underlined"
          data-test="email-text"
        />

        <v-text-field
          color="primary"
          prepend-inner-icon="mdi-lock"
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

        <v-text-field
          color="primary"
          prepend-inner-icon="mdi-lock"
          :append-inner-icon="showConfirmPassword ? 'mdi-eye' : 'mdi-eye-off'"
          v-model="passwordConfirm"
          :error-messages="passwordConfirmError"
          label="Confirm Password"
          required
          variant="underlined"
          data-test="password-confirm-text"
          :type="showConfirmPassword ? 'text' : 'password'"
          @click:append-inner="showConfirmPassword = !showConfirmPassword"
        />
      </v-container>

      <div>
        <v-checkbox
          v-model="acceptPrivacyPolicy"
          color="primary"
          hide-details
          data-test="accept-privacy-policy-checkbox"
        >
          <template #label>
            <span class="caption">
              I agree to the
              <a
                href="https://www.shellhub.io/privacy-policy"
                target="_blank"
                rel="noopener noreferrer"
              >Privacy
                Policy</a>
            </span>
          </template>
        </v-checkbox>
        <v-checkbox
          v-model="acceptMarketing"
          color="primary"
          hide-details
          data-test="accept-news-checkbox"
        >
          <template #label>
            <p>
              I accept to receive news and updates from ShellHub via
              email.
            </p>
          </template>
        </v-checkbox>
      </div>

      <v-card-actions class="justify-center">
        <v-btn
          :disabled="!acceptPrivacyPolicy"
          type="submit"
          data-test="create-account-btn"
          color="primary"
          :variant="acceptPrivacyPolicy ? 'elevated' : 'tonal'"
          block
        >
          SignUp
        </v-btn>
      </v-card-actions>

      <v-card-subtitle
        class="d-flex align-center justify-center pa-4 mx-auto"
        data-test="login-btn"
      >
        Do you have an account?
        <router-link
          class="ml-1"
          :to="{ name: 'Login' }"
        >
          Login
        </router-link>
      </v-card-subtitle>
    </form>
    <AccountCreated
      :show="showMessage"
      :message-kind="messageKind"
      :username="username"
      data-test="accountCreated-component"
    />
  </v-container>
</template>

<script setup lang="ts">
import { ref, onMounted, Ref, computed } from "vue";
import { useRouter, useRoute } from "vue-router";
import { useField } from "vee-validate";
import * as yup from "yup";
import axios, { AxiosError } from "axios";
import { useStore } from "../store";
import AccountCreated from "../components/Account/AccountCreated.vue";

const store = useStore();
const router = useRouter();
const route = useRoute();
const showPassword = ref(false);
const showConfirmPassword = ref(false);
const showMessage = ref(false);
const acceptMarketing = ref(false);
const acceptPrivacyPolicy = ref(false);
const isEmailLocked = ref(false);
const messageKind: Ref<"sig" | "normal"> = ref("normal");
const token = computed(() => store.getters["users/getSignToken"]);
const userStatus = computed(() => store.getters["namespaces/getUserStatus"]);

const alertVisible = computed(
  () => userStatus.value === "invited"
  || (route.query.redirect?.includes("/accept-invite") && !showMessage.value),
);

const sigValue = ref("");

const {
  value: name,
  errorMessage: nameError,
  setErrors: setNameError,
} = useField<string>("name", yup.string().required()
  .min(1, "Your name should be 1-64 characters long")
  .max(64, "Your name should be 1-64 characters long"), {
  initialValue: "",
});

const {
  value: username,
  errorMessage: usernameError,
  setErrors: setUsernameError,
} = useField<string>("username", yup.string().required().min(3).max(32), {
  initialValue: "",
});

const {
  value: email,
  errorMessage: emailError,
  setErrors: setEmailError,
} = useField<string>("email", yup.string().email().required(), {
  initialValue: "",
});

const {
  value: password,
  errorMessage: passwordError,
  setErrors: setPasswordError,
} = useField<string>("password", yup.string().required().min(5).max(32), {
  initialValue: "",
});

const {
  value: passwordConfirm,
  errorMessage: passwordConfirmError,
} = useField<string>("passwordConfirm", yup.string().required()
  .test("passwords-match", "Passwords do not match", (value) => password.value === value), {
  initialValue: "",
});

onMounted(() => {
  const emailQuery = route.query.email as string;
  sigValue.value = route.query.sig as string;

  if (emailQuery && sigValue.value) {
    email.value = emailQuery;
    isEmailLocked.value = true;
  }
});

const hasErrors = () => !!(
  nameError.value
  || usernameError.value
  || emailError.value
  || passwordError.value
  || passwordConfirmError.value
);

const handleAxiosError = (error: AxiosError) => {
  const responseData = error.response?.data;
  if (Array.isArray(responseData)) {
    if (responseData.includes("username")) setUsernameError("This username already exists");
    if (responseData.includes("name")) setNameError("This name is invalid!");
    if (responseData.includes("password")) setPasswordError("This password is invalid!");
    if (responseData.includes("email")) setEmailError("This email is invalid!");
  }
};

const createAccount = async () => {
  if (!hasErrors()) {
    try {
      const signUpData = {
        name: name.value,
        email: email.value,
        username: username.value,
        password: password.value,
        confirmPassword: passwordConfirm.value,
        emailMarketing: acceptMarketing.value,
        sig: sigValue.value,
      };

      await store.dispatch("users/signUp", signUpData);
      showMessage.value = true;

      if (!token.value) {
        await router.push({ name: "ConfirmAccount", query: { username: username.value } });
      }
      messageKind.value = "sig";
    } catch (error: unknown) {
      if (axios.isAxiosError(error)) {
        handleAxiosError(error);
      }
    }
  }
};

</script>
