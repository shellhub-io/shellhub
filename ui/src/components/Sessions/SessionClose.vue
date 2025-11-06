<template>
  <v-list-item
    v-bind="$attrs"
    :disabled="!hasAuthorization"
    @click="showDialog = true"
  >
    <div class="d-flex align-center">
      <div class="mr-2">
        <v-icon> mdi-close-circle </v-icon>
      </div>

      <v-list-item-title data-test="mdi-information-list-item">
        Close Session
      </v-list-item-title>
    </div>
  </v-list-item>

  <MessageDialog
    v-model="showDialog"
    title="Are you sure?"
    description="You are going to close connection for this device. After confirming this action cannot be undone."
    icon="mdi-close-circle"
    icon-color="error"
    confirm-text="Close"
    confirm-color="error"
    cancel-text="Cancel"
    data-test="close-session-dialog"
    confirm-data-test="close-session-btn"
    cancel-data-test="cancel-close-session-btn"
    @close="showDialog = false"
    @cancel="showDialog = false"
    @confirm="closeSession"
  />
</template>

<script setup lang="ts">
import { ref } from "vue";
import { IDevice } from "@/interfaces/IDevice";
import handleError from "@/utils/handleError";
import useSnackbar from "@/helpers/snackbar";
import MessageDialog from "@/components/Dialogs/MessageDialog.vue";
import useSessionsStore from "@/store/modules/sessions";

const props = defineProps<{
  uid: string;
  device: IDevice;
  hasAuthorization: boolean;
}>();

const emit = defineEmits(["update"]);
const showDialog = ref(false);
const sessionsStore = useSessionsStore();
const snackbar = useSnackbar();

const closeSession = async () => {
  try {
    await sessionsStore.closeSession({
      uid: props.uid,
      device_uid: props.device.uid,
    });
    showDialog.value = false;
    snackbar.showSuccess("Session closed successfully.");
    emit("update");
  } catch (error: unknown) {
    snackbar.showError("Failed to close session.");
    handleError(error);
  }
};
</script>
