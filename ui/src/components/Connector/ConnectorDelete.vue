<template>
  <v-list-item
    v-bind="$attrs"
    :disabled="!hasAuthorization"
    data-test="connector-remove-btn"
    @click="showDialog = true"
  >
    <div class="d-flex align-center">
      <div class="mr-2">
        <v-icon
          data-test="remove-icon"
          icon="mdi-delete"
        />
      </div>
      <v-list-item-title data-test="remove-title">
        Remove
      </v-list-item-title>
    </div>
  </v-list-item>

  <MessageDialog
    v-model="showDialog"
    title="Are you sure?"
    description="You are about to remove this Docker Connector. After confirming this action cannot be undone."
    icon="mdi-alert-circle"
    icon-color="warning"
    confirm-text="Remove"
    confirm-color="error"
    cancel-text="Close"
    confirm-data-test="remove-btn"
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
import useConnectorStore from "@/store/modules/connectors";

const props = defineProps<{
  uid: string;
  hasAuthorization: boolean;
}>();

const emit = defineEmits(["update"]);
const showDialog = ref(false);
const connectorStore = useConnectorStore();
const snackbar = useSnackbar();

const remove = async () => {
  try {
    await connectorStore.deleteConnector(props.uid);
    snackbar.showSuccess("Successfully removed connector.");
    showDialog.value = false;
    emit("update");
  } catch (error: unknown) {
    snackbar.showError("Failed to remove connector.");
    handleError(error);
  }
};

defineExpose({ showDialog });
</script>
