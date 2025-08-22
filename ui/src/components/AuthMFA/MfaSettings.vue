<template>
  <BaseDialog
    v-model="showDialog"
    scrollable
    transition="dialog-bottom-transition"
    data-test="dialog"
    @click:outside="close()"
  >
    <v-card class="bg-v-theme-surface content" data-test="card-first-page">
      <v-container>
        <v-window v-model="el">
          <v-window-item :value="1">
            <v-row>
              <v-col align="center" data-test="title">
                <v-card-title class="mt-2" data-test="card-text">
                  Add a recovery mail to proceed into the MFA Process
                </v-card-title>
              </v-col>-
            </v-row>
            <v-row class="mb-2">
              <v-col align="center">
                <h4>
                  In case you lose access to all your MFA credentials,
                  we'll need a recovery email to verify your identity
                  and reset your account access.
                </h4>
                <p class="mt-2">
                  To ensure you can recover your account if you lose access to your MFA credentials, please associate a recovery email.
                </p>
              </v-col>
            </v-row>
            <v-row>
              <v-col align="center">
                <v-text-field
                  width="400"
                  v-model="recoveryEmail"
                  label="Recovery Email"
                  :error-messages="recoveryEmailError"
                  required
                  variant="underlined"
                  data-test="recovery-email-text"
                />
              </v-col>
            </v-row>
            <v-card-actions>
              <v-btn variant="text" data-test="close-btn" @click="close">
                Close
              </v-btn>
              <v-spacer />
              <v-btn
                :disabled="!recoveryEmail || !!recoveryEmailError"
                variant="text"
                color="primary"
                data-test="disable-btn"
                @click="updateRecoveryEmail"
              >
                Save Recovery Email
              </v-btn>
            </v-card-actions>
          </v-window-item>
          <v-window-item :value="2">
            <v-row>
              <v-col align="center" class="pt-2" data-test="title-first-page">
                <h2>Your Recovery Codes</h2>
              </v-col>
            </v-row>
            <v-row>
              <v-col>
                <v-col class="pa-0">
                  <v-alert
                    variant="text"
                    type="warning"
                    :icon="false"
                    data-test="alert-first-page"
                    text="Please tick the box below when you're confident you've saved your recovery codes.
                     Without them, you won't be able to get back into your account if you lose your MFA device.
                     Keep in mind that the codes will change if you come back to this page."
                  />
                </v-col>
              </v-col>
            </v-row>
            <v-card class="mb-2">
              <v-row>
                <v-col
                  v-for="(code, index) in recoveryCodes"
                  :key="index"
                  :cols="4"
                  align="center"
                  class="pa-4 ma-0 pl-0 pr-0"
                  data-test="recovery-codes"
                >
                  <h4>{{ code }}</h4>
                </v-col>
              </v-row>

            </v-card>

            <v-row>
              <v-col>
                <v-btn
                  @click="downloadRecoveryCodes()"
                  color="primary"
                  tabindex="0"
                  variant="elevated"
                  prepend-icon="mdi-download-box-outline"
                  class="mr-2"
                  data-test="download-recovery-codes-btn"
                >Download</v-btn>
                <CopyWarning :copied-item="'Recovery codes'">
                  <template #default="{ copyText }">
                    <v-btn
                      @click="copyText(recoveryCodes.join('\n'))"
                      color="primary"
                      tabindex="0"
                      variant="elevated"
                      prepend-icon="mdi-content-copy"
                      data-test="copy-recovery-codes-btn"
                    >Copy</v-btn>
                  </template>
                </CopyWarning>
              </v-col>
            </v-row>

            <v-row>
              <v-col class="pt-0">
                <v-checkbox
                  v-model="checkbox"
                  data-test="checkbox-recovery"
                  label="I have saved my recovery codes and I want to continue the MFA Setup"
                  @click="checkbox === true"
                />
              </v-col>
            </v-row>
            <v-card-actions>
              <v-btn variant="text" data-test="close-btn" @click="close">
                Close
              </v-btn>

              <v-spacer />

              <v-btn variant="text" :disabled="!checkbox" color="primary" data-test="next-btn" @click="goToNextStep()">
                Next Step
              </v-btn>
            </v-card-actions>
          </v-window-item>
          <v-window-item :value="3">
            <v-card class="bg-v-theme-surface content" data-test="card-second-page">
              <v-row>
                <v-col align="center" data-test="title-second-page">
                  <h2>Set up multi-factor authentication</h2>
                </v-col>
              </v-row>
              <v-row>
                <v-col align="center" data-test="qr-code">
                  <qrcode-vue :value="mfaQRCode" :size="250" level="L" render-as="svg" :margin="2" />
                </v-col>
              </v-row>
              <v-row>
                <v-col data-test="secret">
                  <v-text-field
                    readonly
                    label="Secret"
                    variant="outlined"
                    v-model="secret"
                    hint="Alternatively, you can use this secret to enable in your MFA App if you cannot read the QR code."
                    persistent-hint
                  />
                </v-col>
              </v-row>
              <v-row>
                <v-col data-test="step-one">
                  <p><strong>Step 1:</strong> To configure your multi-factor authentication,
                    either scan the QR code above or manually enter the Secret Key provided
                    into your preferred TOTP (Time-Based One-Time Password) provider.</p>
                </v-col>
              </v-row>
              <v-row>
                <v-col data-test="step-two">
                  <p><strong>Step 2:</strong> Enter the 6-digit code from your TOTP provider after signing in.</p>
                </v-col>
              </v-row>
              <v-alert
                v-if="errorAlert"
                type="error"
                data-test="error-alert">{{ errorMessage }} </v-alert>
              <v-row>
                <v-col>
                  <v-otp-input
                    data-test="verification-code"
                    required
                    v-model="verificationCode"
                    @keyup.enter="verificationCode ? enableMfa() : false"
                    label="Verification Code"
                    variant="underlined" />
                </v-col>
              </v-row>
              <v-card-actions>
                <v-btn variant="text" color="primary" data-test="back-btn" @click="el--">
                  Back
                </v-btn>

                <v-spacer />
                <v-btn
                  variant="text"
                  :disabled="!verificationCode"
                  color="primary"
                  data-test="verify-btn"
                  @click="enableMfa()">
                  Verify
                </v-btn>
              </v-card-actions>
            </v-card>
          </v-window-item>
          <v-window-item :value="4">
            <v-row>
              <v-col align="center" data-test="congratulation-text">
                <h2>Congratulations! You've successfully verified your code.</h2>
              </v-col>
            </v-row>
            <v-row>
              <v-col align="center">
                <v-icon
                  end
                  icon="mdi-cloud-lock-outline"
                  color="green"
                  size="100"
                  class="green-cloud"
                  data-test="green-cloud-icon" />
              </v-col>
            </v-row>
            <v-row>
              <v-col align="start" class="ml-5 pb-0" data-test="title-bp">
                <h4>Your account is now more secure with:</h4>
              </v-col>
            </v-row>
            <v-row>
              <v-col class="ml-5 pt-0" data-test="congratulation-bullet-point">
                <ul>
                  <li>Two-step verification adding an extra layer of protection.</li>
                  <li>Reduced risk of unauthorized access even if your password is compromised.</li>
                  <li>Enhanced security against phishing attacks and identity theft.</li>
                </ul>
              </v-col>
            </v-row>
            <v-card-actions>
              <v-spacer />
              <v-btn variant="text" data-test="close-btn" @keyup.enter="close()" @click="close()">
                Close
              </v-btn>
            </v-card-actions>
          </v-window-item>
        </v-window>
      </v-container>
    </v-card>
  </BaseDialog>
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
import BaseDialog from "../BaseDialog.vue";
import useAuthStore from "@/store/modules/auth";
import useUsersStore from "@/store/modules/users";
import { IUserPatch } from "@/interfaces/IUser";

const authStore = useAuthStore();
const usersStore = useUsersStore();
const snackbar = useSnackbar();
const el = ref<number>(1);
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
  el.value++;
  if (el.value === 2) await generateMfa();
};

const updateRecoveryEmail = async () => {
  try {
    await usersStore.patchData({
      recovery_email: recoveryEmail.value,
    } as IUserPatch);
    authStore.recoveryEmail = recoveryEmail.value;
    snackbar.showSuccess("Recovery email updated successfully.");
    emit("update:recovery-email", recoveryEmail.value);
    goToNextStep();
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
    el.value = 4;
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
};

watch(showDialog, async (newValue) => {
  if (newValue && hasRecoveryEmail.value) {
    el.value = 2;
    await generateMfa();
  }
});

defineExpose({
  goToNextStep,
  el,
  showDialog,
});

</script>

<style scoped>
.green-cloud {
  filter: drop-shadow(0px 0px 30px rgba(43, 255, 10, 0.444))
}
</style>
