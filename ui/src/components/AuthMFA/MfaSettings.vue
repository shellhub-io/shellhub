<template>
  <v-btn
    @click="setupMfa()"
    color="primary"
    tabindex="0"
    variant="elevated"
    data-test="enable-dialog-btn"
    class="mb-6"
  >Enable MFA</v-btn>
  <v-row justify="center">
    <v-dialog v-model="dialog" width="auto" scrollable transition="dialog-bottom-transition" data-test="dialog">
      <v-card class="bg-v-theme-surface content" width="650" data-test="card-first-page">
        <v-container>
          <v-window v-model="el">
            <v-window-item :value="1">
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
                  <v-btn
                    @click="copyRecoveryCodes()"
                    color="primary"
                    tabindex="0"
                    variant="elevated"
                    prepend-icon="mdi-content-copy"
                    data-test="copy-recovery-codes-btn"
                  >Copy</v-btn>
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
                <v-btn variant="text" data-test="close-btn" @click="dialog = !dialog">
                  Close
                </v-btn>

                <v-spacer />

                <v-btn variant="text" :disabled="!checkbox" color="primary" data-test="next-btn" @click="goToNextStep()">
                  Next Step
                </v-btn>
              </v-card-actions>
            </v-window-item>
            <v-window-item :value="2">
              <v-card class="bg-v-theme-surface content" data-test="card-second-page">
                <v-row>
                  <v-col align="center" data-test="title-second-page">
                    <h2>Set up multi-factor authentication</h2>
                  </v-col>
                </v-row>
                <v-row>
                  <v-col align="center" data-test="qr-code">
                    <qrcode-vue :value="value" :size="250" level="L" render-as="svg" :margin="2" />
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
                  <v-col data-test="secret">
                    <p>Secret: <strong>{{ secret }}</strong></p>
                  </v-col>
                </v-row>
                <v-row>
                  <v-col data-test="step-two">
                    <p><strong>Step 2:</strong> Enter the 6-digit code from your TOTP provider after signing in.</p>
                  </v-col>
                </v-row>
                <v-row v-if="errorAlert">
                  <v-col>
                    <v-alert
                      type="error"
                      :text="errorMessage"
                      data-test="error-alert" />
                  </v-col>
                </v-row>
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
                  <v-btn variant="text" :disabled="!verificationCode" color="primary" data-test="verify-btn" @click="enableMfa()">
                    Verify
                  </v-btn>
                </v-card-actions>
              </v-card>
            </v-window-item>
          </v-window>
        </v-container>
      </v-card>
    </v-dialog>
  </v-row>
</template>

<script setup lang="ts">
import { ref, computed } from "vue";
// eslint-disable-next-line import/no-extraneous-dependencies
import QrcodeVue from "qrcode.vue";
import axios, { AxiosError } from "axios";
import { useClipboard } from "@vueuse/core";
import { useStore } from "@/store";
import handleError from "@/utils/handleError";
import { INotificationsCopy } from "@/interfaces/INotifications";

const store = useStore();
const el = ref<number>(1);
const dialog = ref(false);
const value = computed(() => store.getters["auth/link_mfa"]);
const secret = computed(() => store.getters["auth/secret"]);
const recoveryCodes = computed(() => store.getters["auth/recoveryCodes"]);
const verificationCode = ref("");
const checkbox = ref(false);
const errorAlert = ref(false);
const errorMessage = ref("");

const setupMfa = async () => {
  try {
    await store.dispatch("auth/generateMfa").then(() => {
      el.value = 1;
      dialog.value = true;
      checkbox.value = false;
    });
  } catch (error) {
    handleError(error);
  }
};

const copyRecoveryCodes = () => {
  const codesText = recoveryCodes.value.join("\n");
  const { copy } = useClipboard();
  copy(codesText);

  store.dispatch(
    "snackbar/showSnackbarCopy",
    INotificationsCopy.recoveryCodes,
  );
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

const goToNextStep = () => {
  el.value++;
};

const enableMfa = async () => {
  try {
    await store.dispatch("auth/enableMfa", {
      token_mfa: verificationCode.value,
      secret: secret.value,
      codes: recoveryCodes.value,
    });
  } catch (error) {
    if (axios.isAxiosError(error)) {
      const axiosError = error as AxiosError;
      errorAlert.value = true;
      switch (axiosError.response?.status) {
        case 500:
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

defineExpose({
  goToNextStep,
  el,
});
</script>
