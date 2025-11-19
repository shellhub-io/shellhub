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
    <form
      v-if="!showMessage"
      @submit.prevent="createAccount"
    >
      <v-card-title class="text-center">
        Create Account
      </v-card-title>
      <v-container>
        <v-text-field
          v-model="name"
          color="primary"
          prepend-inner-icon="mdi-account"
          :error-messages="nameError"
          required
          label="Name"
          variant="underlined"
          data-test="name-text"
        />

        <v-text-field
          v-model="username"
          color="primary"
          prepend-inner-icon="mdi-account"
          :error-messages="usernameError"
          required
          label="Username"
          variant="underlined"
          data-test="username-text"
        />

        <v-text-field
          v-model="email"
          color="primary"
          prepend-inner-icon="mdi-email"
          :disabled="isEmailLocked"
          :error-messages="emailError"
          required
          label="Email"
          variant="underlined"
          data-test="email-text"
        />

        <v-text-field
          v-model="password"
          color="primary"
          prepend-inner-icon="mdi-lock"
          :append-inner-icon="showPassword ? 'mdi-eye' : 'mdi-eye-off'"
          :error-messages="passwordError"
          label="Password"
          required
          variant="underlined"
          data-test="password-text"
          :type="showPassword ? 'text' : 'password'"
          @click:append-inner="showPassword = !showPassword"
          @update:model-value="handlePasswordChange"
        />

        <v-text-field
          v-model="passwordConfirm"
          color="primary"
          prepend-inner-icon="mdi-lock"
          :append-inner-icon="showConfirmPassword ? 'mdi-eye' : 'mdi-eye-off'"
          :error-messages="passwordConfirmError"
          label="Confirm Password"
          required
          variant="underlined"
          data-test="password-confirm-text"
          :type="showConfirmPassword ? 'text' : 'password'"
          @click:append-inner="showConfirmPassword = !showConfirmPassword"
          @update:model-value="handlePasswordChange"
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
          :disabled="hasErrors()"
          type="submit"
          data-test="create-account-btn"
          color="primary"
          :variant="!hasErrors() ? 'elevated' : 'tonal'"
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
import AccountCreated from "../components/Account/AccountCreated.vue";
import useNamespacesStore from "@/store/modules/namespaces";
import useUsersStore from "@/store/modules/users";

const namespacesStore = useNamespacesStore();
const usersStore = useUsersStore();
const router = useRouter();
const route = useRoute();
const showPassword = ref(false);
const showConfirmPassword = ref(false);
const showMessage = ref(false);
const acceptMarketing = ref(false);
const acceptPrivacyPolicy = ref(false);
const isEmailLocked = ref(false);
const messageKind: Ref<"sig" | "normal"> = ref("normal");
const userStatus = computed(() => namespacesStore.userStatus);

const alertVisible = computed(
  () => (userStatus.value === "invited" || route.query.redirect?.includes("/accept-invite")) && !showMessage.value,
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
} = useField<string>(
  "username",
  yup
    .string()
    .required("Username is required")
    .min(3, "Username must be at least 3 characters")
    .max(32, "Username must not exceed 32 characters")
    .matches(/^[a-z0-9-_.@]+$/, "Username can only contain lowercase letters and numbers"),
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
} = useField<string>("password", yup.string().required("This field is required").min(5).max(32), {
  initialValue: "",
});

const {
  value: passwordConfirm,
  errorMessage: passwordConfirmError,
  setErrors: setPasswordConfirmError,
} = useField<string>("passwordConfirm", yup.string().required("This field is required")
  .test("passwords-match", "Passwords do not match", (value) => password.value === value), {
  initialValue: "",
});

const handlePasswordChange = () => {
  if (!passwordConfirm.value || !password.value) return;

  if (password.value !== passwordConfirm.value) {
    setPasswordConfirmError("Passwords do not match");
    return;
  }

  setPasswordConfirmError("");
};

const hasErrors = () => !!(
  nameError.value
  || usernameError.value
  || emailError.value
  || passwordError.value
  || passwordConfirmError.value
  || !acceptPrivacyPolicy.value
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

      const token = await usersStore.signUp(signUpData);
      showMessage.value = true;

      if (!token) {
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

onMounted(() => {
  const emailQuery = route.query.email as string;
  sigValue.value = route.query.sig as string;

  if (emailQuery && sigValue.value) {
    email.value = emailQuery;
    isEmailLocked.value = true;
  }
});
</script>
