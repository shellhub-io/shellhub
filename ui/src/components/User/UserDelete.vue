<template>
  <MessageDialog
    v-model="showDialog"
    title="Confirm Account Deletion"
    :description="dialogDescription"
    icon="mdi-account-remove"
    icon-color="error"
    cancel-text="Cancel"
    cancel-data-test="close-btn"
    confirm-text="Delete Account"
    confirm-color="error"
    confirm-data-test="delete-user-btn"
    :confirm-disabled="hasNamespaces"
    data-test="user-delete-dialog"
    @cancel="showDialog = false"
    @confirm="deleteAccount"
  >
    <v-alert
      v-if="hasNamespaces"
      type="warning"
      variant="tonal"
      class="mb-4"
      data-test="namespace-warning"
    >
      <strong>Warning:</strong> You cannot delete your account while you have active namespaces.
      Please delete all your owned namespaces before attempting to delete your account.
    </v-alert>
  </MessageDialog>
</template>

<script setup lang="ts">
import { computed } from "vue";
import { useRouter } from "vue-router";
import { AxiosError } from "axios";
import handleError from "@/utils/handleError";
import useSnackbar from "@/helpers/snackbar";
import MessageDialog from "@/components/Dialogs/MessageDialog.vue";
import useAuthStore from "@/store/modules/auth";
import useNamespacesStore from "@/store/modules/namespaces";

const authStore = useAuthStore();
const { hasNamespaces } = useNamespacesStore();
const snackbar = useSnackbar();
const router = useRouter();
const showDialog = defineModel<boolean>({ default: false });

const dialogDescription = computed(() => (
  hasNamespaces
    ? "You cannot delete your account while you have active namespaces."
    : "Are you sure you want to delete your account? This action cannot be undone."
));

const deleteAccount = async () => {
  try {
    await authStore.deleteUser();
    snackbar.showSuccess("Account deleted successfully.");
    await router.push({ name: "Login" });
  } catch (error: unknown) {
    if (error instanceof AxiosError && error.response?.status === 403) {
      snackbar.showError("You cannot delete your account while you have active namespaces.");
      return;
    }
    snackbar.showError("Failed to delete account.");
    handleError(error);
  }
};

defineExpose({ showDialog });
</script>
