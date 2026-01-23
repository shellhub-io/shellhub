<template>
  <WindowDialog
    v-model="showDialog"
    :title="currentStepConfig.title"
    :description="currentStepConfig.description"
    :icon="currentStepConfig.icon"
    icon-color="warning"
    data-test="dialog"
    @close="close"
  >
    <v-window
      v-model="el"
      class="pa-6"
    >
      <v-slide-y-reverse-transition>
        <v-alert
          v-model="showAlert"
          :text="alertMessage"
          type="error"
          closable
          variant="tonal"
          class="mb-4 align-self-stretch"
          data-test="alert-message"
          role="alert"
          aria-live="assertive"
        />
      </v-slide-y-reverse-transition>
      <v-window-item :value="1">
        <v-otp-input
          v-model="verificationCode"
          data-test="verification-code"
          required
          label="Verification Code"
          class="mb-4"
          autocomplete="one-time-code"
          @keyup.enter="verificationCode ? disableMfa() : false"
        />

        <p class="text-subtitle-2 text-center">
          If you lost your MFA TOTP Provider and want to use your recovery code,
          <span
            tag="button"
            class="text-primary cursor-pointer text-decoration-underline"
            data-test="use-recovery-code-btn"
            @click="goToNextStep"
            @keyup.enter="goToNextStep"
          >
            click here
          </span>
        </p>
      </v-window-item>

      <v-window-item :value="2">
        <v-text-field
          v-model="recoveryCode"
          color="primary"
          class="mx-auto mt-2"
          required
          label="Recovery Code"
          autocomplete="one-time-code"
          data-test="recovery-code"
          width="400"
          @keyup.enter="recoveryCode ? disableMfa() : false"
        />

        <p class="text-subtitle-2 text-center">
          If you lost your recovery codes, we'll send you an e-mail to continue
          the MFA disable.
          <span
            tag="button"
            class="text-primary cursor-pointer text-decoration-underline"
            data-test="recover-email-btn"
            @click="requestMail"
            @keyup.enter="requestMail"
          >
            Click here</span>.
        </p>
      </v-window-item>

      <v-window-item :value="3">
        <div class="text-center">
          <v-icon
            icon="mdi-email-check-outline"
            size="80"
            color="success"
            class="mb-4"
          />
        </div>

        <p
          data-test="sub-title"
          class="mb-4 text-center text-body-1 font-weight-bold"
        >
          An email has been sent to {{ userMail }}. Please check your inbox and
          click the link we've provided to disable MFA.
        </p>
      </v-window-item>
    </v-window>

    <template #footer>
      <v-spacer />
      <v-card-actions>
        <v-btn
          data-test="close-btn"
          @click="close"
        >
          Close
        </v-btn>
        <v-btn
          v-if="el === 1"
          :disabled="!verificationCode"
          data-test="verify-btn"
          color="primary"
          @click="disableMfa"
        >
          Verify
        </v-btn>
        <v-btn
          v-else-if="el === 2"
          :disabled="!recoveryCode"
          data-test="recover-btn"
          color="primary"
          @click="disableMfa"
        >
          Recover Account
        </v-btn>
      </v-card-actions>
    </template>
  </WindowDialog>
</template>

<script setup lang="ts">
import { computed, ref } from "vue";
import axios, { AxiosError } from "axios";
import handleError from "@/utils/handleError";
import useSnackbar from "@/helpers/snackbar";
import WindowDialog from "@/components/Dialogs/WindowDialog.vue";
import useAuthStore from "@/store/modules/auth";

const authStore = useAuthStore();
const snackbar = useSnackbar();
const verificationCode = ref("");
const recoveryCode = ref("");
const el = ref<1 | 2 | 3>(1);
const showAlert = ref(false);
const alertMessage = ref("");
const showDialog = defineModel<boolean>({ required: true });
const userMail = computed(() => localStorage.getItem("email"));

const stepConfig = {
  1: {
    title: "Disable Multi-Factor Authentication",
    description: "Verify your identity using your authenticator app",
    icon: "mdi-shield-remove-outline",
  },
  2: {
    title: "Use Recovery Code",
    description: "Enter one of your backup recovery codes",
    icon: "mdi-shield-key-outline",
  },
  3: {
    title: "Email Verification Sent",
    description: "Check your email to complete MFA removal",
    icon: "mdi-email-check-outline",
  },
};

const currentStepConfig = computed(() => stepConfig[el.value]);

const disableMfa = async () => {
  try {
    const params = {
      1: { code: verificationCode.value },
      2: { recovery_code: recoveryCode.value },
    }[el.value as 1 | 2];

    await authStore.disableMfa(params);
    snackbar.showSuccess("MFA disabled successfully.");
    showDialog.value = false;
  } catch (error) {
    if (axios.isAxiosError(error)) {
      const axiosError = error as AxiosError;
      showAlert.value = true;
      switch (axiosError.response?.status) {
        case 403:
          alertMessage.value
            = "The verification code sent in your MFA verification is invalid, please try again.";
          break;
        default:
          alertMessage.value
            = "An error occurred during your MFA verification, try again later.";
          handleError(error);
      }
      return;
    }
    handleError(error);
  }
};

const goToNextStep = () => {
  el.value++;
  alertMessage.value = "";
  showAlert.value = false;
};

const requestMail = async () => {
  try {
    await authStore.requestMfaReset();
    goToNextStep();
  } catch (error) {
    if (axios.isAxiosError(error)) {
      showAlert.value = true;
      alertMessage.value
        = "An error occurred sending your recovery mail, please try again later.";
    }
    handleError(error);
  }
};

const close = () => {
  showDialog.value = false;
  recoveryCode.value = "";
  verificationCode.value = "";
  alertMessage.value = "";
  showAlert.value = false;
  el.value = 1;
};
</script>
