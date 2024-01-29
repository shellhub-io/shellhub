<template>
  <v-dialog
    v-model="dialog"
    transition="dialog-bottom-transition"
    width="650"
  >
    <v-card class="bg-v-theme-surface" data-test="card-dialog">
      <v-row>
        <v-col align="center" data-test="title">
          <v-card-title class="mt-2" data-test="card-text">
            Verification of access to the authentication device
          </v-card-title>
        </v-col>
      </v-row>
      <v-row>
        <v-col class="pl-3 pr-3 pb-0 pt-0" align="center">
          <v-alert
            variant="text"
            type="warning"
            :icon="false"
            data-test="alert"
            text="Please check the box only if you are confident that
             you still have access to the device used for Multi-Factor Authentication."
          />
          <v-alert
            variant="text"
            type="warning"
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
    </v-card>
  </v-dialog>
</template>

<script setup lang="ts">
import { onMounted, ref } from "vue";
import { useStore } from "../../store";
import handleError from "@/utils/handleError";
import { INotificationsSuccess } from "@/interfaces/INotifications";

const dialog = ref(false);
const checkbox = ref(false);
const store = useStore();
const disableMFA = async () => {
  try {
    await store.dispatch("auth/disableMfa");
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
  dialog.value = true;
});
</script>
