<template>
  <v-list-item
    v-bind="$attrs"
    @click="showDialog = true"
    :disabled="!hasAuthorization"
    data-test="device-delete-item"
  >
    <div class="d-flex align-center">
      <div class="mr-2" data-test="remove-icon">
        <v-icon>mdi-delete</v-icon>
      </div>

      <v-list-item-title data-test="remove-title"> Remove </v-list-item-title>
    </div>
  </v-list-item>

  <BaseDialog v-model="showDialog" data-test="delete-device-dialog">
    <v-card class="bg-v-theme-surface" data-test="device-delete-card">
      <v-card-title class="text-h5 pa-5 bg-primary" data-test="dialog-title">
        Are you sure?
      </v-card-title>
      <v-divider />

      <v-card-text class="mt-4 mb-0 pb-1" data-test="dialog-text">
        <p class="text-body-2 mb-2">You are about to remove this {{ variant }}.</p>

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
          @click="removeDevice()"
          data-test="confirm-btn"
        >
          Remove
        </v-btn>
      </v-card-actions>
    </v-card>
  </BaseDialog>
</template>

<script setup lang="ts">
import { ref } from "vue";
import { useRouter } from "vue-router";
import { useStore } from "@/store";
import handleError from "@/utils/handleError";
import useSnackbar from "@/helpers/snackbar";
import BaseDialog from "../BaseDialog.vue";
import useDevicesStore from "@/store/modules/devices";

const props = defineProps<{
  uid: string;
  redirect?: boolean;
  hasAuthorization?: boolean;
  variant: string;
}>();

const emit = defineEmits(["update"]);
const showDialog = ref(false);
const snackbar = useSnackbar();
const store = useStore();
const devicesStore = useDevicesStore();
const router = useRouter();

const emitUpdate = (): void => {
  emit("update");
};

const removeDevice = async (): Promise<void> => {
  try {
    await devicesStore.removeDevice(props.uid);

    if (props.redirect) {
      router.push("/devices");
    } else {
      await store.dispatch("tags/fetch");
    }

    snackbar.showSuccess("Successfully removed device.");
    emitUpdate();
  } catch (error: unknown) {
    snackbar.showError("Failed to remove device.");
    handleError(error);
  }

  showDialog.value = false;
};

defineExpose({ removeDevice });
</script>
