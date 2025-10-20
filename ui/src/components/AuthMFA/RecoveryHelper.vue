<template>
  <WindowDialog
    v-model="showDialog"
    @close="close"
    transition="dialog-bottom-transition"
    :persistent="!checkbox"
    title="MFA Recovery Verification"
    description="Verify access to your authentication device"
    icon="mdi-shield-refresh-outline"
    icon-color="warning"
    :show-close-button="false"
  >
    <div class="pa-8">
      <v-slide-y-reverse-transition>
        <v-alert
          v-model="tokenCountdownAlert"
          type="warning"
          :title="invalid.title + (invalid.timeout ? countdownTimer : '')"
          :text="invalid.msg"
          @click:close="!tokenCountdownAlert"
          variant="tonal"
          data-test="invalid-login-alert"
          class="mb-4"
        />
      </v-slide-y-reverse-transition>

      <v-slide-y-reverse-transition>
        <v-alert
          v-model="isCountdownFinished"
          type="error"
          title="Your recovery code timeout has finished"
          text="To disable your MFA now, please proceed to your profile settings."
          variant="tonal"
          data-test="invalid-login-alert"
          class="mb-4"
        />
      </v-slide-y-reverse-transition>
      <p class="mb-4 text-justify">
        Recovery codes prove useful when you must access your account and your authentication device is unavailable.
        Nevertheless, bear in mind that if you lose access to the device, it is advisable to disable Multi-Factor Authentication
        and re-enable it using a currently accessible device.
      </p>
      <p class="mb-4 text-justify">
        Please check the box only if you are confident that you still have access to the device used for Multi-Factor Authentication.
      </p>

      <v-checkbox
        v-model="checkbox"
        data-test="checkbox-recovery"
        label="I have access to my authentication device and don't need to disable MFA"
      />
    </div>

    <template #footer>
      <v-spacer />
      <v-card-actions>
        <v-btn
          :disabled="!checkbox"
          variant="text"
          data-test="close-btn"
          @click="close"
        >
          Close
        </v-btn>
        <v-btn
          :disabled="isCountdownFinished"
          variant="text"
          color="error"
          data-test="disable-btn"
          @click="disableMFA"
        >
          Disable MFA
        </v-btn>
      </v-card-actions>
    </template>
  </WindowDialog>
</template>

<script setup lang="ts">
import { onMounted, reactive, ref, watch } from "vue";
import handleError from "@/utils/handleError";
import useCountdown from "@/utils/countdownTimeout";
import useSnackbar from "@/helpers/snackbar";
import WindowDialog from "@/components/Dialogs/WindowDialog.vue";
import useAuthStore from "@/store/modules/auth";

const showDialog = defineModel<boolean>({ required: true });
const authStore = useAuthStore();
const checkbox = ref(false);
const invalid = reactive({ title: "", msg: "", timeout: false });
const tokenCountdownAlert = ref(true);
const isCountdownFinished = ref(false);
const snackbar = useSnackbar();
const { disableTimeout, recoveryCode } = authStore;
const { startCountdown, countdown } = useCountdown();

const countdownTimer = ref("");

watch(countdown, (newValue) => {
  countdownTimer.value = newValue;
  if (countdownTimer.value === "0 seconds") {
    tokenCountdownAlert.value = false;
    isCountdownFinished.value = true;
    checkbox.value = true;
  }
});

const close = async () => {
  authStore.isRecoveringMfa = false;
  showDialog.value = false;
};

const disableMFA = async () => {
  try {
    await authStore.disableMfa({ recovery_code: recoveryCode });
    snackbar.showSuccess("MFA disabled successfully.");
    close();
  } catch (error) {
    snackbar.showError("An error occurred while disabling MFA.");
    handleError(error);
  }
};

onMounted(() => {
  startCountdown(disableTimeout);
  tokenCountdownAlert.value = true;
  Object.assign(invalid, {
    title: "Your recovery code will expire in ",
    // eslint-disable-next-line vue/max-len
    msg: "If you want to disable your MFA status, do it within the timers limit, if the timer runs out, you will have to disable your MFA in the profile settings.",
    timeout: true,
  });
  showDialog.value = true;
});
</script>
