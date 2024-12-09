<template>
  <v-container class="pb-0 mb-0">
    <v-alert
      v-if="alertMessage"
      :type="alertType"
      variant="tonal"
      :text="alertMessage"
      class="mb-4"
      data-test="user-status-alert"
    />
    <form @submit.prevent="setupAccount">
      <v-card-title class="text-center" data-test="welcome-title">Welcome to ShellHub!</v-card-title>
      <v-window v-model="el">
        <v-window-item :value="1">
          <v-card-subtitle style="white-space: normal;" data-test="subtitle-1">
            To set up your account, please run <code>/bin/setup</code> in your terminal to generate a signature.
            Use the generated signature in the "Sign" field below to proceed.
          </v-card-subtitle>
          <v-container>
            <v-text-field
              color="primary"
              prepend-inner-icon="mdi-key"
              v-model="sign"
              :disabled="!!hasQuery"
              :error-messages="signError"
              required
              label="Sign"
              variant="underlined"
              data-test="sign-text"
            />
            <v-btn
              :disabled="!hasSign"
              type="submit"
              data-test="sign-btn"
              color="primary"
              @click="el = 2"
              variant="tonal"
              block
            >
              Setup
            </v-btn>
          </v-container>
        </v-window-item>
        <v-window-item :value="2">
          <v-card-subtitle class="d-inline-block text-center" style="white-space: normal;" data-test="subtitle-2">
            Please complete the following form to set up your account with your personal information.
          </v-card-subtitle>
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

          <v-card-actions class="justify-center">
            <v-btn
              :disabled="!isFormValid"
              type="submit"
              data-test="setup-account-btn"
              color="primary"
              variant="tonal"
              block
            >
              Create Account
            </v-btn>

          </v-card-actions>
        </v-window-item>
      </v-window>
    </form>
  </v-container>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from "vue";
import { useRouter, useRoute } from "vue-router";
import { useField } from "vee-validate";
import * as yup from "yup";
import { useStore } from "../store";

const store = useStore();
const router = useRouter();
const route = useRoute();
const showPassword = ref(false);
const showConfirmPassword = ref(false);
const alertMessage = ref("");
const alertType = ref<"warning" | "success" | "info" | "error">("warning");
const el = ref<number>(1);
const hasQuery = computed(() => route.query.sign as string);

const {
  value: sign,
  errorMessage: signError,
} = useField<string>("sign", yup.string().required(), {
  initialValue: route.query.sign as string,
});

const {
  value: name,
  errorMessage: nameError,
} = useField<string>("name", yup.string().required()
  .min(1, "Your name should be 1-64 characters long")
  .max(64, "Your name should be 1-64 characters long"), {
  initialValue: "",
});

const {
  value: username,
  errorMessage: usernameError,
} = useField<string>("username", yup.string()
  .required("Username is required")
  .min(3, "Username must be at least 3 characters")
  .max(32, "Username must not exceed 32 characters")
  .matches(/^[a-z0-9-_.@]+$/, "Username can only contain lowercase letters, numbers, and certain symbols"), {
  initialValue: "",
});

const {
  value: email,
  errorMessage: emailError,
} = useField<string>("email", yup.string().email().required(), {
  initialValue: "",
});

const {
  value: password,
  errorMessage: passwordError,
} = useField<string>("password", yup.string().required()
  .min(5, "Password must be at least 5 characters long")
  .max(32, "Password must not exceed 32 characters"), {
  initialValue: "",
});

const {
  value: passwordConfirm,
  errorMessage: passwordConfirmError,
} = useField<string>("passwordConfirm", yup.string().required()
  .test("passwords-match", "Passwords do not match", (value) => password.value === value), {
  initialValue: "",
});

const hasSign = computed(() => !!sign.value);

const isFormValid = computed(() => (
  name.value
    && username.value
    && email.value
    && password.value
    && passwordConfirm.value
    && !nameError.value
    && !usernameError.value
    && !emailError.value
    && !passwordError.value
    && !passwordConfirmError.value
));

onMounted(() => {
  if (hasQuery.value) {
    el.value = 2;
  }
});

const setupAccount = async () => {
  if (isFormValid.value) {
    try {
      const setupData = {
        sign: sign.value,
        name: name.value,
        username: username.value,
        email: email.value,
        password: password.value,
      };

      await store.dispatch("users/setup", setupData);

      alertType.value = "success";
      alertMessage.value = "Successfully created your account. Redirecting to login...";
      setTimeout(() => router.push({ name: "Login" }), 3000);
    } catch (error) {
      alertType.value = "error";
      alertMessage.value = "An error occurred. please check if the sign matches the same in ./bin/setup and try again.";
    }
  }
};

defineExpose({ el });
</script>
