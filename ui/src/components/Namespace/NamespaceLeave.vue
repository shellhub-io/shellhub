<template>
  <v-dialog v-model="dialog" max-width="540">
    <v-card data-test="namespace-leave-dialog" class="bg-v-theme-surface">
      <v-card-title class="text-h5 pa-4 bg-primary" data-test="title">
        Namespace Leave
      </v-card-title>

      <v-card-text class="mt-4 mb-3 pb-1" data-test="subtitle">
        <div>
          <p class="mb-2">
            Are you sure you want to leave this namespace? If you leave, you will need an invitation to rejoin.
          </p>
        </div>
      </v-card-text>

      <v-card-actions>
        <v-spacer />

        <v-btn variant="text" data-test="close-btn" @click="dialog = !dialog">
          Close
        </v-btn>

        <v-btn
          color="red darken-1"
          variant="text"
          data-test="leave-btn"
          :disabled="hasAuthorization"
          @click="leave()"
        >
          Leave
        </v-btn>
      </v-card-actions>
    </v-card>
  </v-dialog>
</template>

<script setup lang="ts">
import { computed } from "vue";
import { useStore } from "../../store";
import hasPermission from "../../utils/permission";
import { actions, authorizer } from "../../authorizer";
import {
  INotificationsError,
  INotificationsSuccess,
} from "../../interfaces/INotifications";
import handleError from "@/utils/handleError";

const store = useStore();
const dialog = defineModel({ default: false });
const tenant = computed(() => localStorage.getItem("tenant"));

const hasAuthorization = computed(() => {
  const role = store.getters["auth/role"];
  if (role !== "") {
    return !hasPermission(
      authorizer.role[role],
      actions.namespace.leave,
    );
  }
  return false;
});

const leave = async () => {
  try {
    dialog.value = !dialog.value;

    await store.dispatch("namespaces/leave", tenant.value).then(() => {
      window.location.reload();
    });

    await store.dispatch(
      "snackbar/showSnackbarSuccessAction",
      INotificationsSuccess.namespaceLeave,
    );
  } catch (error: unknown) {
    store.dispatch(
      "snackbar/showSnackbarErrorAction",
      INotificationsError.namespaceLeave,
    );
    handleError(error);
  }
};

defineExpose({ dialog });
</script>
