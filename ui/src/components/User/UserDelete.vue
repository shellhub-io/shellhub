<template>
  <v-dialog v-model="show" max-width="540">
    <v-card data-test="user-delete-dialog" class="bg-v-theme-surface">
      <v-card-title class="text-h5 pa-4 bg-primary" data-test=title>
        Confirm Account Deletion
      </v-card-title>

      <v-card-text class="mt-4 mb-3 pb-1" data-test="subtitle">
        <div>
          <p class="mb-2">
            Are you sure you want to delete your account? This action cannot be undone.
          </p>
        </div>
      </v-card-text>

      <v-card-actions>
        <v-spacer />

        <v-btn variant="text" data-test="close-btn" @click="show = !show">
          Cancel
        </v-btn>

        <v-btn
          color="red darken-1"
          variant="text"
          data-test="delete-user-btn"
          @click="deleteAccount()"
        >
          Delete Account
        </v-btn>
      </v-card-actions>
    </v-card>
  </v-dialog>
</template>

<script setup lang="ts">
import { useRouter } from "vue-router";
import { useStore } from "@/store";
import handleError from "@/utils/handleError";
import useSnackbar from "@/helpers/snackbar";

const store = useStore();
const snackbar = useSnackbar();
const router = useRouter();
const show = defineModel<boolean>({ default: false });

const deleteAccount = async () => {
  try {
    await store.dispatch("auth/deleteUser");
    snackbar.showSuccess("Account deleted successfully.");
    router.push({ name: "Login" });
  } catch (error: unknown) {
    snackbar.showError("Failed to delete account.");
    handleError(error);
  }
};

defineExpose({ show });
</script>
