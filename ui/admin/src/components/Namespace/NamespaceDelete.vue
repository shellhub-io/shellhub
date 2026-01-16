<template>
  <MessageDialog
    v-model="showDialog"
    title="Namespace Deletion"
    icon="mdi-delete-alert"
    icon-color="error"
    confirm-color="error"
    confirm-text="Remove"
    :confirm-loading="isLoading"
    cancel-text="Close"
    confirm-data-test="remove-btn"
    cancel-data-test="close-btn"
    @close="showDialog = false"
    @confirm="remove"
    @cancel="showDialog = false"
  >
    <p data-test="content-text">
      This action cannot be undone. This will permanently delete
      <strong>{{ name }}</strong> and its related
      data.
    </p>
  </MessageDialog>
</template>

<script setup lang="ts">
import { ref } from "vue";
import { useRouter, useRoute } from "vue-router";
import handleError from "@/utils/handleError";
import useSnackbar from "@/helpers/snackbar";
import MessageDialog from "@/components/Dialogs/MessageDialog.vue";
import useNamespacesStore from "@admin/store/modules/namespaces";

const props = defineProps<{ tenant: string; name: string }>();
const emit = defineEmits(["update"]);

const namespacesStore = useNamespacesStore();
const snackbar = useSnackbar();
const router = useRouter();
const route = useRoute();

const showDialog = defineModel<boolean>({ required: true });
const isLoading = ref(false);

const remove = async () => {
  isLoading.value = true;
  try {
    await namespacesStore.deleteNamespace(props.tenant);
    snackbar.showSuccess("Namespace deleted successfully.");

    if (route.name === "namespaceDetails") {
      await router.push({ name: "namespaces" });
    } else {
      emit("update");
    }
    showDialog.value = false;
  } catch (error: unknown) {
    snackbar.showError("An error occurred while deleting the namespace.");
    handleError(error);
  } finally {
    isLoading.value = false;
  }
};
</script>
