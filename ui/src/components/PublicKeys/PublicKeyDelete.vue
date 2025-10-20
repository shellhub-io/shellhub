<template>
  <v-list-item v-bind="$attrs" @click="showDialog = true" :disabled="!hasAuthorization" data-test="public-key-remove-btn">
    <div class="d-flex align-center">
      <div class="mr-2">
        <v-icon data-test="remove-icon"> mdi-delete </v-icon>
      </div>

      <v-list-item-title data-test="remove-title"> Remove </v-list-item-title>
    </div>
  </v-list-item>

  <MessageDialog
    v-model="showDialog"
    @close="showDialog = false"
    @confirm="remove"
    @cancel="showDialog = false"
    title="Are you sure?"
    description="You are about to delete this public key"
    icon="mdi-alert"
    icon-color="error"
    confirm-text="Delete"
    confirm-color="error"
    cancel-text="Close"
    confirm-data-test="confirm-btn"
    cancel-data-test="close-btn"
    data-test="delete-public-key-dialog"
  />
</template>

<script setup lang="ts">
import { ref } from "vue";
import handleError from "@/utils/handleError";
import useSnackbar from "@/helpers/snackbar";
import MessageDialog from "@/components/Dialogs/MessageDialog.vue";
import usePublicKeysStore from "@/store/modules/public_keys";

const { fingerprint, hasAuthorization } = defineProps<{
  fingerprint: string;
  hasAuthorization: boolean;
}>();
const emit = defineEmits(["update"]);

const publicKeysStore = usePublicKeysStore();
const showDialog = ref(false);
const snackbar = useSnackbar();

const remove = async () => {
  try {
    await publicKeysStore.deletePublicKey(fingerprint);
    snackbar.showSuccess("The public key was removed successfully");
    emit("update");
  } catch (error: unknown) {
    snackbar.showError("Failed to remove the public key.");
    handleError(error);
  } finally {
    showDialog.value = false;
  }
};
</script>
