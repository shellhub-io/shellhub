<template>
  <div>
    <v-list-item
      @click="showDialog = true"
      :disabled="!hasAuthorization"
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

    <BaseDialog v-model="showDialog">
      <v-card class="bg-v-theme-surface">
        <v-card-title class="text-h5 pa-3 bg-primary">
          Are you sure?
        </v-card-title>
        <v-divider />

        <v-card-text class="mt-4 mb-0 pb-1">
          <p class="text-body-2 mb-2">
            You are going to close connection for this device.
          </p>

          <p class="text-body-2 mb-2">
            After confirming this action cannot be redone.
          </p>
        </v-card-text>

        <v-card-actions>
          <v-spacer />

          <v-btn variant="text" @click="showDialog = false"> Cancel </v-btn>

          <v-btn color="red darken-1" variant="text" @click="closeSession()">
            Close
          </v-btn>
        </v-card-actions>
      </v-card>
    </BaseDialog>
  </div>
</template>

<script setup lang="ts">
import { ref } from "vue";
import { IDevice } from "@/interfaces/IDevice";
import handleError from "@/utils/handleError";
import useSnackbar from "@/helpers/snackbar";
import BaseDialog from "../BaseDialog.vue";
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
