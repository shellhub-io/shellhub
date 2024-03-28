<template>
  <v-list-item
    v-bind="$attrs"
    @click="showDialog = true"
    :disabled="notHasAuthorization"
    data-test="device-delete-item"
  >
    <div class="d-flex align-center">
      <div class="mr-2" data-test="remove-icon">
        <v-icon>mdi-delete</v-icon>
      </div>

      <v-list-item-title data-test="remove-title"> Remove </v-list-item-title>
    </div>
  </v-list-item>

  <v-dialog max-width="450" v-model="showDialog" data-test="delete-dialog">
    <v-card class="bg-v-theme-surface" data-test="device-delete-card">
      <v-card-title class="text-h5 pa-5 bg-primary" data-test="dialog-title">
        Are you sure?
      </v-card-title>
      <v-divider />

      <v-card-text class="mt-4 mb-0 pb-1" data-test="dialog-text">
        <p class="text-body-2 mb-2">You are about to remove this device.</p>

        <p class="text-body-2 mb-2">
          After confirming this action cannot be redone.
        </p>
      </v-card-text>

      <v-card-actions>
        <v-spacer />

        <v-btn variant="text" @click="showDialog = false" data-test="close-btn">
          Close
        </v-btn>

        <v-btn
          color="red darken-1"
          variant="text"
          @click="removeDevice(props.uid, props.redirect)"
          data-test="confirm-btn"
        >
          Remove
        </v-btn>
      </v-card-actions>
    </v-card>
  </v-dialog>
</template>

<script setup lang="ts">
import { ref } from "vue";
import { useRouter } from "vue-router";
import {
  INotificationsError,
  INotificationsSuccess,
} from "@/interfaces/INotifications";
import { useStore } from "../../store";
import handleError from "@/utils/handleError";

const props = defineProps({
  uid: {
    type: String,
    required: true,
  },
  redirect: {
    type: Boolean,
    default: false,
  },
  notHasAuthorization: {
    type: Boolean,
    default: false,
  },
});
const emit = defineEmits(["update"]);
const showDialog = ref(false);
const store = useStore();
const router = useRouter();

const showSuccessNotification = (message: string): void => {
  store.dispatch("snackbar/showSnackbarSuccessAction", message);
};

const showErrorNotification = (message: string): void => {
  store.dispatch("snackbar/showSnackbarErrorAction", message);
};

const emitUpdate = (): void => {
  emit("update");
};

const closeDialog = (): void => {
  showDialog.value = false;
};

const removeDevice = async (uid: string, redirect: boolean): Promise<void> => {
  try {
    await store.dispatch("devices/remove", uid);

    if (redirect) {
      router.push("/devices");
    } else {
      await store.dispatch("tags/fetch");
    }

    showSuccessNotification(INotificationsSuccess.deviceDelete);
    emitUpdate();
  } catch (error: unknown) {
    showErrorNotification(INotificationsError.deviceDelete);
    handleError(error);
  } finally {
    closeDialog();
  }
};

defineExpose({ removeDevice });
</script>
