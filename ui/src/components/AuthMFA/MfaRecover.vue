<template>
  <v-container>
    <v-slide-y-reverse-transition>
      <v-alert
        v-model="showAlert"
        :text="alertMessage"
        type="error"
        closable
        variant="tonal"
        class="mb-4"
        data-test="alert-message"
        role="alert"
        aria-live="assertive"
      />
    </v-slide-y-reverse-transition>
    <v-row>
      <v-col align="center">
        <h3 data-test="title">
          Multi-factor Authentication
        </h3>
      </v-col>
    </v-row>
    <v-row class="mb-2">
      <v-col align="center">
        <h4 data-test="sub-title">
          If you lost your access to your MFA TOTP provider, please paste one of your recovery codes below
        </h4>
      </v-col>
    </v-row>
    <v-text-field
      v-model="recoveryCode"
      color="primary"
      required
      label="Recovery Code"
      autocomplete="one-time-code"
      variant="outlined"
      data-test="recovery-code"
      @keyup.enter="recoveryCode ? recoverMfa() : false"
    />
    <v-card-actions class="justify-center pa-0">
      <v-btn
        :disabled="!recoveryCode"
        data-test="recover-btn"
        color="primary"
        variant="tonal"
        block
        @click="recoverMfa()"
      >
        Recover Account
      </v-btn>
    </v-card-actions>
    <v-row>
      <v-col class="text-caption pa-4 mx-auto pt-4 pb-0">
        If you lost your recovery codes, we'll send you an e-mail to disable this account MFA,
        <v-btn
          class="text-caption pl-0 pb-1"
          variant="plain"
          color="primary"
          density="compact"
          @click="requestMail()"
        >
          click here
        </v-btn>
      </v-col>
    </v-row>
  </v-container>
</template>

<script setup lang="ts">
import { ref } from "vue";
import axios, { AxiosError } from "axios";
import { useRouter } from "vue-router";
import handleError from "@/utils/handleError";
import useAuthStore from "@/store/modules/auth";

const authStore = useAuthStore();
const router = useRouter();
const recoveryCode = ref("");
const showAlert = ref(false);
const alertMessage = ref("");

const recoverMfa = async () => {
  try {
    await authStore.recoverMfa(recoveryCode.value);
    await router.push("/");
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

const requestMail = async () => {
  try {
    await authStore.requestMfaReset();
    await router.push("/recover-mfa/mail-sucessful");
  } catch (error) {
    if (!(axios.isAxiosError(error) && error.response?.status === 403)) handleError(error);

    showAlert.value = true;
    alertMessage.value = "An error occurred sending your recovery mail, please try again later.";
  }
};
</script>
