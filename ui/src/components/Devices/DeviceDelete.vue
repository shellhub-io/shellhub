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

  <MessageDialog
    v-model="showDialog"
    @close="showDialog = false"
    @confirm="removeDevice"
    @cancel="showDialog = false"
    title="Are you sure?"
    :description="`You are about to remove this ${variant}. After confirming this action cannot be redone.`"
    icon="mdi-alert"
    icon-color="error"
    confirm-text="Remove"
    confirm-color="error"
    cancel-text="Close"
    confirm-data-test="confirm-btn"
    cancel-data-test="close-btn"
    data-test="delete-device-dialog"
  />
</template>

<script setup lang="ts">
import { ref } from "vue";
import { useRouter } from "vue-router";
import handleError from "@/utils/handleError";
import useSnackbar from "@/helpers/snackbar";
import MessageDialog from "../MessageDialog.vue";
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
const devicesStore = useDevicesStore();
const router = useRouter();

const removeDevice = async (): Promise<void> => {
  try {
    await devicesStore.removeDevice(props.uid);

    if (props.redirect) router.push("/devices");

    snackbar.showSuccess("Successfully removed device.");
    emit("update");
  } catch (error: unknown) {
    snackbar.showError("Failed to remove device.");
    handleError(error);
  }

  showDialog.value = false;
};

defineExpose({ removeDevice });
</script>
