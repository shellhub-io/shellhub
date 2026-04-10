<template>
  <v-container class="pb-0 my-0">
    <v-alert
      v-if="alertMessage"
      :type="alertType"
      variant="tonal"
      :text="alertMessage"
      class="mb-4"
      data-test="user-status-alert"
      :role="alertType === 'error' ? 'alert' : 'status'"
      :aria-live="alertType === 'error' ? 'assertive' : 'polite'"
    />
    <v-form @submit.prevent="setupAccount">
      <v-card-title
        class="text-center"
        data-test="welcome-title"
      >
        Welcome to ShellHub!
      </v-card-title>
      <v-window v-model="step">
        <v-window-item
          v-if="showOnboardingStep"
          :value="SetupStep.Onboarding"
        >
          <v-card-subtitle
            class="text-wrap text-center mb-4"
            data-test="subtitle-2"
          >
            Help us improve ShellHub by sharing your feedback
          </v-card-subtitle>

          <div style="position: relative; height:60dvh; overflow:auto;">
            <iframe
              :src="onboardingUrl"
              frameborder="0"
              style="position: absolute; left:0; top:0; width:100%; height:100%; border:0;"
            />
          </div>

          <v-card-actions class="mt-4">
            <v-spacer />
            <v-btn
              :disabled="!surveyCompleted"
              color="primary"
              variant="tonal"
              data-test="continue-btn"
              @click="step = SetupStep.Account"
            >
              Continue
            </v-btn>
          </v-card-actions>
        </v-window-item>

        <v-window-item :value="SetupStep.Account">
          <v-card-subtitle
            class="text-wrap text-center mb-3"
            data-test="subtitle-3"
          >
            Please complete the following form to set up your account with your personal information.
          </v-card-subtitle>
          <v-text-field
            v-model="name"
            color="primary"
            prepend-inner-icon="mdi-account"
            :error-messages="nameError"
            required
            label="Name"
            autocomplete="name"
            data-test="name-text"
          />

          <v-text-field
            v-model="username"
            color="primary"
            prepend-inner-icon="mdi-account"
            :error-messages="usernameError"
            required
            label="Username"
            autocomplete="username"
            data-test="username-text"
          />

          <v-text-field
            v-model="email"
            color="primary"
            prepend-inner-icon="mdi-email"
            :error-messages="emailError"
            required
            label="Email"
            autocomplete="email"
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
            autocomplete="new-password"
            data-test="password-text"
            :type="showPassword ? 'text' : 'password'"
            @click:append-inner="showPassword = !showPassword"
          />

          <v-text-field
            v-model="passwordConfirm"
            color="primary"
            prepend-inner-icon="mdi-lock"
            :append-inner-icon="showConfirmPassword ? 'mdi-eye' : 'mdi-eye-off'"
            :error-messages="passwordConfirmError"
            label="Confirm Password"
            required
            autocomplete="new-password"
            data-test="password-confirm-text"
            :type="showConfirmPassword ? 'text' : 'password'"
            @click:append-inner="showConfirmPassword = !showConfirmPassword"
          />

          <v-card-actions class="mt-4">
            <v-btn
              v-if="showOnboardingStep"
              color="primary"
              variant="text"
              @click="step = SetupStep.Onboarding"
            >
              Back
            </v-btn>
            <v-spacer />
            <v-btn
              :disabled="!isFormValid"
              type="submit"
              data-test="setup-account-btn"
              color="primary"
              variant="tonal"
              text="Create Account"
            />
          </v-card-actions>
        </v-window-item>
      </v-window>
    </v-form>
  </v-container>
</template>

<script setup lang="ts">
import { ref, computed } from "vue";
import { useRouter } from "vue-router";
import { useField } from "vee-validate";
import { useEventListener } from "@vueuse/core";
import * as yup from "yup";
import useUsersStore from "@/store/modules/users";
import { envVariables } from "@/envVariables";

enum SetupStep {
  Onboarding = 1,
  Account = 2,
}

const usersStore = useUsersStore();
const router = useRouter();

const showPassword = ref(false);
const showConfirmPassword = ref(false);
const alertMessage = ref("");
const alertType = ref<"success" | "error">("success");
const surveyCompleted = ref(false);

// Onboarding survey is only available in Community Edition
const showOnboardingStep = computed(() => envVariables.isCommunity && !!envVariables.onboardingUrl);

const step = ref<SetupStep>(showOnboardingStep.value ? SetupStep.Onboarding : SetupStep.Account);

const onboardingUrl = computed(() => {
  if (!envVariables.onboardingUrl) {
    return "";
  }

  const baseUrl = envVariables.onboardingUrl;
  const params = new URLSearchParams({
    consent_to_contact: "accepted",
    source: "self-hosted",
    embed: "true",
    instance_domain: window.location.hostname,
  });

  if (import.meta.env.DEV) {
    params.append("preview", "true");
  }

  return `${baseUrl}?${params.toString()}`;
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

// Listen for FormBricks survey completion
useEventListener(window, "message", (event: MessageEvent) => {
  // Verify the message is from FormBricks
  if (!envVariables.onboardingUrl) return;

  try {
    const formbricksOrigin = new URL(envVariables.onboardingUrl).origin;
    if (event.origin !== formbricksOrigin) {
      return;
    }
  } catch {
    return;
  }

  // Check if the survey was completed
  // FormBricks sends the completion event as a simple string
  if (event.data === "formbricksSurveyCompleted") {
    surveyCompleted.value = true;
  }
});

const setupAccount = async () => {
  if (isFormValid.value) {
    try {
      const setupData = {
        name: name.value,
        username: username.value,
        email: email.value,
        password: password.value,
      };

      await usersStore.setup(setupData);

      alertType.value = "success";
      alertMessage.value = "Successfully created your account. Redirecting to login...";
      setTimeout(() => { void router.push({ name: "Login" }); }, 3000);
    } catch {
      alertType.value = "error";
      alertMessage.value = "An error occurred. Please try again.";
    }
  }
};
</script>
