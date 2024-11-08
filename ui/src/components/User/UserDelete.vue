<template>
  <v-tooltip location="bottom" class="text-center" :disabled="!hasNamespace">
    <template v-slot:activator="{ props }">
      <div v-bind="props">
        <v-btn
          :disabled="hasNamespace"
          color="red darken-1"
          variant="outlined"
          data-test="delete-dialog-btn"
          @click="open()"
        >
          Delete Account
        </v-btn>
      </div>
    </template>
    <span> All namespaces must be deleted before you can delete your account. </span>
  </v-tooltip>

  <v-dialog v-model="dialog" max-width="540">
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

        <v-btn variant="text" data-test="close-btn" @click="dialog = !dialog">
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
import { ref, computed } from "vue";
import { useRouter } from "vue-router";
import { useStore } from "../../store";
import {
  INotificationsError,
  INotificationsSuccess,
} from "@/interfaces/INotifications";
import handleError from "@/utils/handleError";

const store = useStore();
const router = useRouter();
const dialog = ref(false);

const hasNamespace = computed(() => store.getters["namespaces/getNumberNamespaces"] > 0);

const open = () => {
  dialog.value = true;
};

const deleteAccount = async () => {
  try {
    dialog.value = !dialog.value;

    await store.dispatch("auth/deleteUser");

    store.dispatch(
      "snackbar/showSnackbarSuccessAction",
      INotificationsSuccess.deleteAccount,
    );

    router.push({ name: "Login" });
  } catch (error: unknown) {
    store.dispatch(
      "snackbar/showSnackbarErrorAction",
      INotificationsError.deleteAccount,
    );
    handleError(error);
  }
};
</script>
