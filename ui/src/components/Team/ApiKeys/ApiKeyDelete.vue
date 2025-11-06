<template>
  <v-list-item
    :disabled="!hasAuthorization"
    @click="showDialog = true"
  >
    <div class="d-flex align-center">
      <div class="d-flex align-center">
        <div
          class="mr-2"
          data-test="delete-icon"
        >
          <v-icon icon="mdi-delete" />
        </div>
        <v-list-item-title data-test="delete-main-btn-title">
          Delete
        </v-list-item-title>
      </div>
    </div>
  </v-list-item>

  <MessageDialog
    v-model="showDialog"
    title="Are you sure?"
    description="You are about to remove this Api Key from the namespace. After confirming this action cannot be undone."
    icon="mdi-key-remove"
    icon-color="error"
    confirm-text="Delete Key"
    confirm-color="error"
    cancel-text="Close"
    confirm-data-test="delete-btn"
    cancel-data-test="close-btn"
    @confirm="remove()"
    @cancel="showDialog = false"
    @close="showDialog = false"
  />
</template>

<script setup lang="ts">
import { ref } from "vue";
import handleError from "@/utils/handleError";
import useSnackbar from "@/helpers/snackbar";
import MessageDialog from "@/components/Dialogs/MessageDialog.vue";
import useApiKeysStore from "@/store/modules/api_keys";

const props = defineProps<{
  keyId: string;
  hasAuthorization: boolean;
}>();

const snackbar = useSnackbar();
const emit = defineEmits(["update"]);
const showDialog = ref(false);
const apiKeyStore = useApiKeysStore();

const update = () => {
  emit("update");
  showDialog.value = false;
};

const remove = async () => {
  try {
    await apiKeyStore.removeApiKey({
      key: props.keyId,
    });
    update();
    snackbar.showSuccess("Api Key deleted successfully.");
  } catch (error: unknown) {
    snackbar.showError("Failed to delete Api Key.");
    handleError(error);
  }
};

defineExpose({ showDialog });
</script>
