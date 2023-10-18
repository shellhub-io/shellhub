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
      />
    </v-slide-y-reverse-transition>
    <v-row>
      <v-col align="center">
        <h3 data-test="title">Multi-factor Authentication</h3>
      </v-col>
    </v-row>
    <v-row class="mb-2">
      <v-col align="center">
        <h4 data-test="sub-title">If you lost your access to your mfa TOTP provider, please paste one of your recovery codes below</h4>
      </v-col>
    </v-row>
    <v-text-field
      v-model="recoveryCode"
      color="primary"
      required
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
        @click="loginMfa()"
      >
        Recover Account
      </v-btn>
    </v-card-actions>
  </v-container>
</template>

<script setup lang="ts">
import { ref } from "vue";
import axios, { AxiosError } from "axios";
import { useRouter } from "vue-router";
import { useStore } from "@/store";
import handleError from "@/utils/handleError";

const store = useStore();
const router = useRouter();
const recoveryCode = ref("");
const showAlert = ref(false);
const alertMessage = ref("");

const loginMfa = async () => {
  try {
    await store.dispatch("auth/recoverLoginMfa", { code: recoveryCode.value });
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

defineExpose({
  showAlert,
});
</script>
