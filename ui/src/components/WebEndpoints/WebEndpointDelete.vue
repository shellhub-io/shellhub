<template>
  <v-btn
    @click="showDialog = true"
    variant="plain"
    class="border rounded bg-v-theme-background"
    density="comfortable"
    size="default"
    icon="mdi-delete"
    :disabled="!canDeleteWebEndpoint"
    data-test="web-endpoint-delete-dialog-btn"
  />

  <MessageDialog
    v-model="showDialog"
    @close="showDialog = false"
    @cancel="showDialog = false"
    @confirm="remove"
    title="Are you sure?"
    description="You are about to remove this Web Endpoint."
    icon="mdi-alert"
    icon-color="error"
    confirm-text="Delete Web Endpoint"
    confirm-color="error"
    cancel-text="Close"
    confirm-data-test="delete-btn"
    cancel-data-test="close-btn"
    data-test="web-endpoint-delete-dialog"
  />
</template>

<script setup lang="ts">
import hasPermission from "@/utils/permission";
import handleError from "@/utils/handleError";
import useSnackbar from "@/helpers/snackbar";
import MessageDialog from "@/components/Dialogs/MessageDialog.vue";
import useWebEndpointsStore from "@/store/modules/web_endpoints";

defineOptions({ inheritAttrs: false });

const props = defineProps<{ address: string }>();
const emit = defineEmits(["update"]);
const showDialog = defineModel({ default: false });

const webEndpointsStore = useWebEndpointsStore();
const snackbar = useSnackbar();

const canDeleteWebEndpoint = hasPermission("webEndpoint:delete");

const update = () => {
  emit("update");
  showDialog.value = false;
};

const remove = async () => {
  try {
    await webEndpointsStore.deleteWebEndpoint(props.address);
    snackbar.showSuccess("Web Endpoint deleted successfully.");
    update();
  } catch (error: unknown) {
    snackbar.showError("Failed to delete Web Endpoint.");
    handleError(error);
  }
};
</script>
