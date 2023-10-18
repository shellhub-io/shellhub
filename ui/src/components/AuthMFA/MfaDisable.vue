<template>
  <v-btn
    @click="dialog = true"
    color="primary"
    tabindex="0"
    variant="elevated"
    data-test="disable-dialog-btn"
  >Disable MFA</v-btn>

  <v-dialog v-model="dialog" width="650" transition="dialog-bottom-transition" data-test="dialog">
    <v-card class="bg-v-theme-surface">
      <v-card-title class="text-h5 pa-5 bg-primary" data-test="title">
        Are you sure?
      </v-card-title>
      <v-divider />
      <v-card-text class="mt-4 mb-0 pb-1" data-test="dialog-text">
        <p class="mb-2">
          Are you sure you want to disable multi-factor authentication (MFA)?
          This will reduce your account's security by removing an additional layer of protection that helps prevent unauthorized access.
        </p>
      </v-card-text>

      <v-card-actions>
        <v-spacer />

        <v-btn variant="text" @click="() => dialog = false" data-test="close-btn"> Close </v-btn>

        <v-btn variant="text" color="red" @click="disableMFA()" data-test="disable-btn">
          Disable
        </v-btn>
      </v-card-actions>
    </v-card>
  </v-dialog>
</template>

<script setup lang="ts">
import { ref } from "vue";
import { useStore } from "@/store";
import handleError from "@/utils/handleError";
import { INotificationsSuccess } from "@/interfaces/INotifications";

const store = useStore();
const dialog = ref(false);
const disableMFA = async () => {
  try {
    await store.dispatch("auth/disableMfa");
    dialog.value = false;
    store.dispatch(
      "snackbar/showSnackbarSuccessAction",
      INotificationsSuccess.cancelMfa,
    );
  } catch (error) {
    store.dispatch(
      "snackbar/setSnackbarErrorDefault",
    );
    handleError(error);
  }
};

</script>
