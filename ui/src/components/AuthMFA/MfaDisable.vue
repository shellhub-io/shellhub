<template>
  <v-btn
    @click="openDialog"
    color="primary"
    tabindex="0"
    variant="elevated"
    data-test="disable-dialog-btn"
  >Disable MFA</v-btn>
  <v-dialog
    max-width="400px"
    scrollable
    v-model="showDialog"
    @click:outside="close"
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
              @keyup.enter="verificationCode ? mfaValidate() : false"
              label="Verification Code"
              variant="underlined" />
            <v-card-actions class="justify-center pa-0">
              <v-row class="ml-4 mr-4">
                <v-col>
                  <v-btn
                    :disabled="!verificationCode"
                    data-test="verify-btn"
                    color="primary"
                    variant="tonal"
                    block
                    @click="mfaValidate()"
                  >
                    Verify
                  </v-btn>
                </v-col>
              </v-row>
            </v-card-actions>
            <v-row>
              <v-col class="text-subtitle-2 mt-2">
                If you lost your MFA TOPT Provider, and want to use your recovery code,
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
              @keyup.enter="recoveryCode ? mfaValidate() : false"
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
                @click="mfaValidate()"
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
  </v-dialog>
</template>

<script setup lang="ts">
import { computed, ref } from "vue";
import axios, { AxiosError } from "axios";
import { useStore } from "@/store";
import handleError from "@/utils/handleError";
import Logo from "@/assets/logo-inverted.png";
import { INotificationsSuccess } from "@/interfaces/INotifications";

const emits = defineEmits(["success"]);

const store = useStore();
const verificationCode = ref("");
const recoveryCode = ref("");
const el = ref<number>(1);
const showAlert = ref(false);
const alertMessage = ref("");
const showDialog = ref(false);
const userMail = computed(() => localStorage.getItem("email"));

const mfaValidate = async () => {
  try {
    switch (el.value) {
      case 1:
        await store.dispatch("auth/disableMfa", { code: verificationCode.value });
        break;
      case 2:
        await store.dispatch("auth/disableMfa", { recovery_code: recoveryCode.value });
        break;
      default:
        break;
    }
    store.dispatch(
      "snackbar/showSnackbarSuccessAction",
      INotificationsSuccess.cancelMfa,
    );
    emits("success");
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
    await store.dispatch("auth/reqResetMfa", userMail.value);
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

const openDialog = () => {
  showDialog.value = true;
};

defineExpose({ el });
</script>
