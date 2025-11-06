<template>
  <v-btn
    variant="plain"
    class="border rounded bg-v-theme-background"
    density="comfortable"
    size="default"
    icon="mdi-delete"
    :disabled="!canDeleteWebEndpoint"
    data-test="web-endpoint-delete-dialog-btn"
    @click="showDialog = true"
  />

  <MessageDialog
    v-model="showDialog"
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
    @close="showDialog = false"
    @cancel="showDialog = false"
    @confirm="remove"
  />
</template>

<script setup lang="ts">
import { ref } from "vue";
import hasPermission from "@/utils/permission";
import handleError from "@/utils/handleError";
import useSnackbar from "@/helpers/snackbar";
import MessageDialog from "@/components/Dialogs/MessageDialog.vue";
import useWebEndpointsStore from "@/store/modules/web_endpoints";

defineOptions({ inheritAttrs: false });

const props = defineProps<{ address: string }>();
const emit = defineEmits(["update"]);
const showDialog = ref(false);

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

defineExpose({ showDialog });
</script>
