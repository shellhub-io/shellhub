<template>
  <v-dialog
    v-model="dialog"
    transition="dialog-bottom-transition"
    width="650"
    persistent
  >
    <v-card class="bg-v-theme-surface" data-test="card-dialog">
      <v-container>
        <v-row>
          <v-col align="center" data-test="title">
            <v-card-title class="mt-2" data-test="card-text">
              Verification of access to the authentication device
            </v-card-title>
          </v-col>
        </v-row>
        <v-row>
          <v-col>
            <v-slide-y-reverse-transition>
              <v-alert
                v-model="tokenCountdownAlert"
                type="warning"
                :title="invalid.title + (invalid.timeout ? countdownTimer : '')"
                :text="invalid.msg"
                @click:close="!tokenCountdownAlert"
                closable
                variant="tonal"
                data-test="invalid-login-alert"
              />
            </v-slide-y-reverse-transition>
            <v-slide-y-reverse-transition>
              <v-alert
                v-model="isCountdownFinished"
                type="error"
                title="Your recovery code timeout has finished"
                text="To disable your MFA now, please proceed to your profile settings."
                closable
                variant="tonal"
                data-test="invalid-login-alert"
              />
            </v-slide-y-reverse-transition>
          </v-col>
        </v-row>
        <v-row>
          <v-col class="pl-3 pr-3 pb-0 pt-0" align="center">
            <v-alert
              variant="text"
              :icon="false"
              data-test="alert"
              text="Please check the box only if you are confident that
             you still have access to the device used for Multi-Factor Authentication."
            />
            <v-alert
              variant="text"
              :icon="false"
              data-test="alert-second"
              text="Recovery codes prove useful when you must access your account and your authentication device is unavailable.
            Nevertheless, bear in mind that if you lose access to the device, it is advisable to disable Multi-Factor Authentication
            and re-enable it using a currently accessible device."
            />
          </v-col>
        </v-row>
        <v-row>
          <v-col class="ml-4 pt-0" align="center" data-test="checkbox">
            <v-checkbox
              class="pl-4"
              v-model="checkbox"
              data-test="checkbox-recovery"
              label="I have access to my authentication device and don't need to disable MFA"
              @click="checkbox === true"
            />
          </v-col>
        </v-row>
        <v-card-actions>
          <v-btn
            :disabled="isCountdownFinished"
            variant="text"
            color="red"
            data-test="disable-btn"
            @click="disableMFA"
          >
            Disable MFA
          </v-btn>
          <v-spacer />
          <v-btn
            :disabled="!checkbox"
            variant="text"
            data-test="close-btn"
            @click="close"> Close </v-btn>
        </v-card-actions>
      </v-container>
    </v-card>
  </v-dialog>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref, watch } from "vue";
import { useStore } from "../../store";
import handleError from "@/utils/handleError";
import { INotificationsSuccess } from "@/interfaces/INotifications";
import useCountdown from "@/utils/countdownTimeout";

const dialog = ref(false);
const checkbox = ref(false);
const invalid = reactive({ title: "", msg: "", timeout: false });
const tokenCountdownAlert = ref(true);
const isCountdownFinished = ref(false);
const store = useStore();
const disableTimeout = computed(() => store.getters["auth/getDisableTokenTimeout"]);
const recoveryCode = computed(() => store.getters["auth/stateRecoveryCode"]);
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

const disableMFA = async () => {
  try {
    await store.dispatch("auth/disableMfa", { recovery_code: recoveryCode.value });
    store.dispatch(
      "snackbar/showSnackbarSuccessAction",
      INotificationsSuccess.recoveryHelper,
    );
    store.commit("auth/accountRecoveryHelper");
    dialog.value = false;
  } catch (error) {
    store.dispatch(
      "snackbar/setSnackbarErrorDefault",
    );
    handleError(error);
  }
};
const close = async () => {
  store.commit("auth/accountRecoveryHelper");
  dialog.value = false;
};

onMounted(() => {
  startCountdown(disableTimeout.value);
  tokenCountdownAlert.value = true;
  Object.assign(invalid, {
    title: "Your recovery code will expire in ",
    // eslint-disable-next-line vue/max-len
    msg: "If you want to disable your MFA status, do it within the timers limit, if the timer runs out, you will have to disable your MFA in the profile settings.",
    timeout: true,
  });
  dialog.value = true;
});
</script>
