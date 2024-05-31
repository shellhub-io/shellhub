<template>
  <v-container>
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
        <h4 data-test="sub-title">Verify your identity by signing in using the code from your OTP Provider</h4>
      </v-col>
    </v-row>
    <v-otp-input
      data-test="verification-code"
      required
      v-model="verificationCode"
      @keyup.enter="verificationCode ? loginMfa() : false"
      label="Verification Code"
      variant="underlined" />
    <v-card-actions class="justify-center pa-0">
      <v-btn
        :disabled="!verificationCode"
        data-test="verify-btn"
        color="primary"
        variant="tonal"
        block
        @click="loginMfa()"
      >
        Verify
      </v-btn>
    </v-card-actions>
    <v-col class="text-caption d-flex align-center justify-center pa-4 mx-auto pt-4 pb-0">
      Did you
      <v-btn
        class="pl-1 text-caption"
        @click="goToRecoveryCodes"
        variant="plain"
        color="primary"
        density="compact"
        data-test="redirect-recover"
      >

        Lost your TOPT password?
      </v-btn>
    </v-col>
  </v-container>
</template>

<script setup lang="ts">
import { computed, ref } from "vue";
import { useRouter } from "vue-router";
import axios, { AxiosError } from "axios";
import { useStore } from "@/store";
import handleError from "@/utils/handleError";

const store = useStore();
const router = useRouter();
const verificationCode = ref("");
const showAlert = ref(false);
const alertMessage = ref("");
const mfaValue = computed(() => store.getters["auth/mfaToken"]);

const loginMfa = async () => {
  try {
    await store.dispatch("auth/validateMfa", { token: mfaValue.value, code: verificationCode.value });
    router.push("/");
  } catch (error) {
    if (axios.isAxiosError(error)) {
      const axiosError = error as AxiosError;
      showAlert.value = true;
      switch (axiosError.response?.status) {
        case 500:
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

const goToRecoveryCodes = () => {
  router.push({ name: "RecoverMfa" });
};

defineExpose({
  showAlert,
});
</script>
