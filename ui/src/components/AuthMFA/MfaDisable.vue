<template>
  <BaseDialog
    v-model="showDialog"
    scrollable
    @close="close"
    data-test="dialog"
  >
    <v-card class="bg-grey-darken-4 bg-v-theme-surface pa-3">
      <v-container>
        <v-window v-model="el">
          <v-card-title class="d-flex justify-center align-center mt-4">
            <v-img
              :src="Logo"
              max-width="220"
              alt="ShellHub logo, a cloud with a shell in your base write ShellHub in the right side"
            />
          </v-card-title>
          <v-window-item :value="1">
            <v-row>
              <v-col align="center">
                <h3 data-test="title">Multi-factor Authentication</h3>
              </v-col>
            </v-row>
            <v-row class="mb-2">
              <v-col align="center">
                <h4 data-test="sub-title">Verify your identity by signing in using the code from your OTP Provider</h4>
              </v-col>
            </v-row>
            <v-slide-y-reverse-transition v-if="showAlert">
              <v-alert
                v-model="showAlert"
                :text="alertMessage"
                type="error"
                closable
                variant="tonal"
                class="mb-4"
                data-test="alert-message"
              />
            </v-slide-y-reverse-transition>
            <v-otp-input
              data-test="verification-code"
              required
              v-model="verificationCode"
              @keyup.enter="verificationCode ? disableMfa() : false"
              label="Verification Code"
              variant="underlined" />
            <v-row>
              <v-col class="text-subtitle-2 mt-2">
                If you lost your MFA TOTP Provider, and want to use your recovery code,
                <v-btn
                  class="pl-0"
                  @click="goToNextStep()"
                  variant="plain"
                  color="primary"
                  density="compact"
                  data-test="use-recovery-code-btn"
                >
                  click here
                </v-btn>
              </v-col>
            </v-row>
            <v-card-actions class="justify-center pa-0">
              <v-row class="ml-4 mr-4 mt-2">
                <v-col>
                  <v-btn
                    :disabled="!verificationCode"
                    data-test="verify-btn"
                    color="primary"
                    variant="tonal"
                    block
                    @click="disableMfa()"
                  >
                    Verify
                  </v-btn>
                </v-col>
              </v-row>
            </v-card-actions>
          </v-window-item>
          <v-window-item :value="2">
            <v-slide-y-reverse-transition v-if="showAlert">
              <v-alert
                v-model="showAlert"
                :text="alertMessage"
                type="error"
                closable
                variant="tonal"
                class="mb-4"
                data-test="alert-message"
              />
            </v-slide-y-reverse-transition>
            <v-row>
              <v-col align="center">
                <h3 data-test="title">Multi-factor Authentication</h3>
              </v-col>
            </v-row>
            <v-row class="mb-2">
              <v-col align="center">
                <h4 data-test="sub-title">If you lost your access to your TOTP provider,
                  please paste one of your recovery codes below</h4>
              </v-col>
            </v-row>
            <v-text-field
              v-model="recoveryCode"
              color="primary"
              required
              @keyup.enter="recoveryCode ? disableMfa() : false"
              label="Recovery Code"
              variant="outlined"
              data-test="recovery-code"
            />
            <v-card-actions class="justify-center pa-0">
              <v-btn
                :disabled="!recoveryCode"
                data-test="recover-btn"
                color="primary"
                variant="tonal"
                block
                @click="disableMfa()"
              >
                Recover Account
              </v-btn>
            </v-card-actions>
            <v-row>
              <v-col class="text-subtitle-2">
                If you lost your recovery codes, we'll send you an e-mail to continue the MFA disable,
                <v-btn
                  class="pl-0"
                  @click="requestMail()"
                  variant="plain"
                  color="primary"
                  density="compact"
                  data-test="send-email-btn"
                >
                  click here
                </v-btn>
              </v-col>
            </v-row>
          </v-window-item>
          <v-window-item :value="3">
            <v-row>
              <v-col align="center">
                <h3 data-test="title">Multi-factor Authentication</h3>
              </v-col>
            </v-row>
            <v-row class="mb-2">
              <v-col align="center">
                <h4 data-test="sub-title">An email has been sent to {{ userMail }}.
                  Please check your inbox and click the link we've provided to disable MFA.</h4>
              </v-col>
            </v-row>
          </v-window-item>
        </v-window>
      </v-container>
      <v-card-actions>
        <v-btn @click="close" data-test="close-btn">Close</v-btn>
      </v-card-actions>
    </v-card>
  </BaseDialog>
</template>

<script setup lang="ts">
import { computed, ref } from "vue";
import axios, { AxiosError } from "axios";
import handleError from "@/utils/handleError";
import Logo from "@/assets/logo-inverted.png";
import useSnackbar from "@/helpers/snackbar";
import BaseDialog from "../BaseDialog.vue";
import useAuthStore from "@/store/modules/auth";

const authStore = useAuthStore();
const snackbar = useSnackbar();
const verificationCode = ref("");
const recoveryCode = ref("");
const el = ref<1 | 2 | 3>(1);
const showAlert = ref(false);
const alertMessage = ref("");
const showDialog = defineModel({ default: false });
const userMail = computed(() => localStorage.getItem("email"));

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
          alertMessage.value = "The verification code sent in your MFA verification is invalid, please try again.";
          break;
        default:
          alertMessage.value = "An error occurred during your MFA verification, try again later.";
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
      alertMessage.value = "An error occurred sending your recovery mail, please try again later.";
    }
    handleError(error);
  }
};

const close = () => {
  recoveryCode.value = "";
  verificationCode.value = "";
  alertMessage.value = "";
  showAlert.value = false;
  showDialog.value = false;
  el.value = 1;
};

defineExpose({ el, showDialog });
</script>
