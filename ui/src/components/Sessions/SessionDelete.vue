<template>
  <div>
    <v-list-item
      @click="showDialog = true"
      :disabled="!hasAuthorization"
    >
      <div class="d-flex align-center">
        <div class="mr-2">
          <v-icon>mdi-playlist-remove</v-icon>
        </div>

        <v-list-item-title data-test="mdi-information-list-item">
          Delete Session Record
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
            You are going to delete the logs recorded for this session.
          </p>

          <p class="text-body-2 mb-2">
            After confirming this action cannot be redone.
          </p>
        </v-card-text>

        <v-card-actions>
          <v-spacer />

          <v-btn variant="text" @click="showDialog = false">Close</v-btn>

          <v-btn color="red darken-1" variant="text" @click="deleteRecord()">
            Remove
          </v-btn>
        </v-card-actions>
      </v-card>
    </BaseDialog>
  </div>
</template>

<script setup lang="ts">
import { ref } from "vue";
import handleError from "@/utils/handleError";
import useSnackbar from "@/helpers/snackbar";
import BaseDialog from "../BaseDialog.vue";
import useSessionsStore from "@/store/modules/sessions";

const props = defineProps<{
  uid: string;
  hasAuthorization: boolean;
}>();

const emit = defineEmits(["update"]);
const snackbar = useSnackbar();
const showDialog = ref(false);
const sessionsStore = useSessionsStore();

const deleteRecord = async () => {
  try {
    await sessionsStore.deleteSessionLogs(props.uid);
    showDialog.value = false;
    snackbar.showSuccess("Successfully deleted the session logs.");
    emit("update");
  } catch (error: unknown) {
    snackbar.showError("An error occurred while deleting the session logs.");
    handleError(error);
  }
};
</script>
