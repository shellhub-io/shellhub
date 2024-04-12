<template>
  <v-container class="pb-0 mb-0">
    <form @submit.prevent="createAccount">
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
        Do you have account ?
        <router-link
          class="ml-1"
          :to="{ name: 'login' }"
        >
          Login
        </router-link>
      </v-card-subtitle>
    </form>
    <AccountCreated
      :show="showMessage"
      :username="username"
      data-test="accountCreated-component"
    />
  </v-container>
</template>

<script setup lang="ts">
import { ref } from "vue";
import { useRouter } from "vue-router";
import { useField } from "vee-validate";
import * as yup from "yup";
import axios, { AxiosError } from "axios";
import { useStore } from "../store";
import AccountCreated from "../components/Account/AccountCreated.vue";

const store = useStore();
const router = useRouter();
const showPassword = ref(false);
const showConfirmPassword = ref(false);
const showMessage = ref(false);
const acceptMarketing = ref(false);
const acceptPrivacyPolicy = ref(false);

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
} = useField<string>(
  "username",
  yup
    .string()
    .required()
    .min(3)
    .max(32)
    .test(
      "username-error",
      "The username only accepts the lowercase letters and this special characters _, ., - and @.",
      (value) => {
        const regex = /^[a-z0-9_.@-\s]*$/;
        return regex.test(value || "");
      },
    )
    .test(
      "white-spaces",
      "The username cannot contain white spaces.",
      (value) => {
        const regex = /\s/;
        return !regex.test(value || "");
      },
    ),
  {
    initialValue: "",
  },
);

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
} = useField<string>(
  "password",
  yup
    .string()
    .required()
    .min(5, "Your password should be 5-32 characters long")
    .max(32, "Your password should be 5-32 characters long"),
  {
    initialValue: "",
  },
);

const {
  value: passwordConfirm,
  errorMessage: passwordConfirmError,
} = useField<string>(
  "passwordConfirm",
  yup
    .string()
    .required()
    .test(
      "passwords-match",
      "Passwords do not match",
      (value) => password.value === value,
    ),
  {
    initialValue: "",
  },
);

const hasErrors = () => !!(
  nameError.value
  || usernameError.value
  || emailError.value
  || passwordError.value
  || passwordConfirmError.value
  || !name.value
  || !username.value
  || !email.value
  || !password.value
  || !passwordConfirm.value
);

const createAccount = async () => {
  if (!hasErrors()) {
    try {
      await store.dispatch("users/signUp", {
        name: name.value,
        email: email.value,
        username: username.value,
        password: password.value,
        confirmPassword: passwordConfirm.value,
        emailMarketing: acceptMarketing.value,
      });

      showMessage.value = !showMessage.value;

      await router.push({ name: "ConfirmAccount", query: { username: username.value } });
    } catch (error: unknown) {
      if (axios.isAxiosError(error)) {
        const axiosError = error as AxiosError;
        const responseData = axiosError.response?.data;
        if (Array.isArray(responseData)) {
          if (responseData.includes("username")) setUsernameError("This username already exists");
          if (responseData.includes("name")) setNameError("This name is invalid!");
          if (responseData.includes("password")) setPasswordError("This password is invalid!");
          if (responseData.includes("email")) setEmailError("This email is invalid!");
        }
      }
    }
  }
};
</script>
