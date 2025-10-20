<template>
  <v-list-item
    v-bind="$attrs"
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

  <MessageDialog
    v-model="showDialog"
    @close="showDialog = false"
    @cancel="showDialog = false"
    @confirm="deleteRecord"
    title="Are you sure?"
    description="You are going to delete the logs recorded for this session. After confirming this action cannot be undone."
    icon="mdi-playlist-remove"
    icon-color="error"
    confirm-text="Remove"
    confirm-color="error"
    cancel-text="Close"
    data-test="delete-session-dialog"
    confirm-data-test="delete-session-btn"
    cancel-data-test="cancel-delete-session-btn"
  />
</template>

<script setup lang="ts">
import { ref } from "vue";
import handleError from "@/utils/handleError";
import useSnackbar from "@/helpers/snackbar";
import MessageDialog from "@/components/Dialogs/MessageDialog.vue";
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
