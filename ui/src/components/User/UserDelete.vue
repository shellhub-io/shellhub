<template>
  <MessageDialog
    v-model="showDialog"
    title="Confirm Account Deletion"
    description="Are you sure you want to delete your account? This action cannot be undone."
    icon="mdi-account-remove"
    icon-color="error"
    cancel-text="Cancel"
    cancel-data-test="close-btn"
    confirm-text="Delete Account"
    confirm-color="error"
    confirm-data-test="delete-user-btn"
    @confirm="deleteAccount"
    data-test="user-delete-dialog"
  />
</template>

<script setup lang="ts">
import { useRouter } from "vue-router";
import handleError from "@/utils/handleError";
import useSnackbar from "@/helpers/snackbar";
import MessageDialog from "../MessageDialog.vue";
import useAuthStore from "@/store/modules/auth";

const authStore = useAuthStore();
const snackbar = useSnackbar();
const router = useRouter();
const showDialog = defineModel<boolean>({ default: false });

const deleteAccount = async () => {
  try {
    await authStore.deleteUser();
    snackbar.showSuccess("Account deleted successfully.");
    router.push({ name: "Login" });
  } catch (error: unknown) {
    snackbar.showError("Failed to delete account.");
    handleError(error);
  }
};

defineExpose({ showDialog });
</script>
