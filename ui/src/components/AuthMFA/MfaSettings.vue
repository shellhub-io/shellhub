<template>
  <WindowDialog
    v-model="showDialog"
    @close="close"
    transition="dialog-bottom-transition"
    :title="currentStepConfig.title"
    :description="currentStepConfig.description"
    :icon="currentStepConfig.icon"
    icon-color="primary"
    data-test="dialog"
  >
    <v-window v-model="step" class="pa-6">
      <v-window-item :value="1">
        <p class="text-justify mb-4 px-1">
          In case you lose access to all your MFA credentials,
          we'll need a recovery email to verify your identity
          and reset your account access.
          To ensure you can recover your account if you lose
          access to your MFA credentials, please associate a
          recovery email.
        </p>
        <v-text-field
          width="400"
          class="mx-auto"
          v-model="recoveryEmail"
          label="Recovery Email"
          :error-messages="recoveryEmailError"
          hide-details="auto"
          required
          data-test="recovery-email-text"
        />

      </v-window-item>
      <v-window-item :value="2">
        <v-alert
          type="warning"
          :icon="false"
          data-test="alert-first-page"
          text="Please tick the box below when you're confident you've saved your recovery codes.
                Without them, you won't be able to get back into your account if you lose your MFA device.
                Keep in mind that the codes will change if you come back to this page."
        />
        <v-card class="my-4 pa-3 border overflow-x-auto">
          <div class="recovery-codes-grid">
            <span
              v-for="code in recoveryCodes"
              :key="code"
              class="py-2 text-center font-weight-medium"
              data-test="recovery-codes"
            >
              {{ code }}
            </span>
          </div>
        </v-card>

        <v-row class="my-4" justify="center">
          <v-btn
            @click="downloadRecoveryCodes"
            color="primary"
            tabindex="0"
            variant="text"
            prepend-icon="mdi-download-box-outline"
            class="mr-3"
            data-test="download-recovery-codes-btn"
          >
            Download
          </v-btn>
          <CopyWarning :copied-item="'Recovery codes'">
            <template #default="{ copyText }">
              <v-btn
                @click="copyText(recoveryCodes.join('\n'))"
                color="primary"
                tabindex="0"
                variant="text"
                prepend-icon="mdi-content-copy"
                data-test="copy-recovery-codes-btn"
              >
                Copy
              </v-btn>
            </template>
          </CopyWarning>
        </v-row>

        <v-checkbox
          v-model="checkbox"
          data-test="checkbox-recovery"
          label="I have saved my recovery codes and I want to continue the MFA Setup"
          @click="checkbox === true"
          hide-details
        />
      </v-window-item>
      <v-window-item :value="3" class="text-center">
        <qrcode-vue :value="mfaQRCode" :size="250" level="L" render-as="svg" :margin="2" />
        <v-text-field
          readonly
          class="my-4 text-left"
          label="Secret"
          v-model="secret"
          hint="Alternatively, you can use this secret to enable in your MFA App if you cannot read the QR code."
          persistent-hint
        />
        <div class="text-left">
          <p class="mb-2"><strong>Step 1:</strong> Scan the QR code above or manually enter the Secret Key provided
            into your preferred TOTP (Time-Based One-Time Password) provider.</p>
          <p class="my-2"><strong>Step 2:</strong> Enter the 6-digit code from your TOTP provider after signing in.</p>
        </div>
        <v-alert
          v-if="errorMessage"
          type="error"
          data-test="error-alert"
          :text="errorMessage"
        />
        <v-otp-input
          data-test="verification-code"
          required
          v-model="verificationCode"
          @keyup.enter="verificationCode ? enableMfa() : false"
          label="Verification Code"
        />
      </v-window-item>
      <v-window-item :value="4">
        <div class="text-center">
          <v-icon
            icon="mdi-cloud-lock-outline"
            color="green"
            size="100"
            class="mx-auto mb-6 green-cloud"
            data-test="green-cloud-icon"
          />
          <p class="text-h5 font-weight-medium">Congratulations! You've successfully enabled MFA.</p>
        </div>
        <p class="ml-3 mt-3 font-weight-medium">Your account is now more secure with:</p>
        <ul class="ml-8 pt-0" data-test="congratulation-bullet-point">
          <li>Two-step verification adding an extra layer of protection.</li>
          <li>Reduced risk of unauthorized access even if your password is compromised.</li>
          <li>Enhanced security against phishing attacks and identity theft.</li>
        </ul>
      </v-window-item>
    </v-window>
    <template #footer>
      <v-spacer />
      <v-card-actions>
        <v-btn v-if="step === 3" data-test="back-btn" @click="step--">Back</v-btn>
        <v-btn v-else data-test="close-btn" @click="close">Close</v-btn>
        <v-btn
          v-if="step === 1"
          :disabled="!recoveryEmail || !!recoveryEmailError"
          color="primary"
          data-test="disable-btn"
          @click="updateRecoveryEmail"
        >
          Save Recovery Email
        </v-btn>
        <v-btn v-else-if="step === 2" :disabled="!checkbox" color="primary" data-test="next-btn" @click="goToNextStep">
          Next Step
        </v-btn>
        <v-btn
          v-else-if="step === 3"
          :disabled="!verificationCode"
          color="primary"
          data-test="verify-btn"
          @click="enableMfa"
        >
          Verify
        </v-btn>
      </v-card-actions>
    </template>
  </WindowDialog>
</template>

<script setup lang="ts">
import * as yup from "yup";
import { useField } from "vee-validate";
import { ref, computed, watch } from "vue";
import QrcodeVue from "qrcode.vue";
import axios, { AxiosError } from "axios";
import handleError from "@/utils/handleError";
import useSnackbar from "@/helpers/snackbar";
import CopyWarning from "@/components/User/CopyWarning.vue";
import WindowDialog from "@/components/Dialogs/WindowDialog.vue";
import useAuthStore from "@/store/modules/auth";
import useUsersStore from "@/store/modules/users";
import { IUserPatch } from "@/interfaces/IUser";

const authStore = useAuthStore();
const usersStore = useUsersStore();
const snackbar = useSnackbar();
const step = ref<number>(1);
const emit = defineEmits(["update:recovery-email"]);
const showDialog = defineModel({ default: false });
const mfaQRCode = ref("");
const secret = ref("");
const recoveryCodes = ref<Array<string>>([]);
const email = computed(() => authStore.email);
const hasRecoveryEmail = computed(() => !!authStore.recoveryEmail);
const verificationCode = ref("");
const checkbox = ref(false);
const errorAlert = ref(false);
const errorMessage = ref("");

const stepConfig = {
  1: {
    title: "Recovery Email Setup",
    description: "Add a recovery email to secure your MFA process",
    icon: "mdi-email-plus-outline",
  },
  2: {
    title: "Save Your Recovery Codes",
    description: "Download and securely store your backup codes",
    icon: "mdi-shield-key-outline",
  },
  3: {
    title: "Configure MFA Device",
    description: "Scan QR code and verify your authenticator app",
    icon: "mdi-qrcode-scan",
  },
  4: {
    title: "MFA Setup Complete",
    description: "Your account is now secured with multi-factor authentication!",
    icon: "mdi-shield-check-outline",
  },
};

const currentStepConfig = computed(() => stepConfig[step.value as keyof typeof stepConfig]);

const {
  value: recoveryEmail,
  errorMessage: recoveryEmailError,
  setErrors: setRecoveryEmailError,
} = useField<string>(
  "recoveryEmail",
  yup
    .string()
    .email("Please enter a valid email address")
    .required("Recovery email is required")
    .test(
      "not-same-as-email",
      "Recovery email must not be the same as your current email",
      (value) => value !== email.value,
    ),
  {
    initialValue: "",
  },
);

const generateMfa = async () => {
  try {
    const data = await authStore.generateMfa();
    mfaQRCode.value = data.link; // QR Code
    secret.value = data.secret;
    recoveryCodes.value = data.recovery_codes;
    checkbox.value = false;
  } catch (error) {
    handleError(error);
  }
};

const goToNextStep = async () => {
  step.value++;
  if (step.value === 2) await generateMfa();
};

const updateRecoveryEmail = async () => {
  try {
    await usersStore.patchData({
      recovery_email: recoveryEmail.value,
    } as IUserPatch);
    authStore.recoveryEmail = recoveryEmail.value;
    snackbar.showSuccess("Recovery email updated successfully.");
    emit("update:recovery-email", recoveryEmail.value);
    await goToNextStep();
  } catch (error) {
    if (axios.isAxiosError(error)) {
      const axiosError = error as AxiosError;
      switch (axiosError.response?.status) {
        case 409:
          setRecoveryEmailError("This recovery email is already in use");
          break;
        case 400:
          setRecoveryEmailError("This recovery email is invalid");
          break;
        default:
          handleError(error);
      }
    }
  }
};

const downloadRecoveryCodes = () => {
  const codesText = recoveryCodes.value.join("\n");
  const blob = new Blob([codesText], { type: "text/plain" });
  const url = URL.createObjectURL(blob);

  const a = document.createElement("a");
  a.href = url;
  a.download = "recovery_codes.txt";
  a.click();

  URL.revokeObjectURL(url);
};

const enableMfa = async () => {
  try {
    await authStore.enableMfa({
      code: verificationCode.value,
      secret: secret.value,
      recovery_codes: recoveryCodes.value,
    });
    step.value = 4;
  } catch (error: unknown) {
    if (axios.isAxiosError(error)) {
      const axiosError = error as AxiosError;
      errorAlert.value = true;
      switch (axiosError.response?.status) {
        case 401:
          errorMessage.value = "The verification code sent in your MFA verification is invalid, please try again.";
          break;
        default:
          errorMessage.value = "An error occurred during your MFA verification, try again later.";
          handleError(error);
      }
      return;
    }
    handleError(error);
  }
};

const close = () => {
  showDialog.value = false;
  errorAlert.value = false;
  recoveryEmail.value = "";
  setRecoveryEmailError("");
};

watch(showDialog, async (newValue) => {
  if (newValue && hasRecoveryEmail.value) {
    step.value = 2;
    await generateMfa();
  }
});

defineExpose({
  goToNextStep,
  step,
  showDialog,
});

</script>

<style scoped>
.green-cloud {
  filter: drop-shadow(0px 0px 30px rgba(43, 255, 10, 0.444))
}

.recovery-codes-grid {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  grid-template-rows: repeat(2, 1fr);
  gap: 8px;
}
</style>
