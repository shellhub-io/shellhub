<template>
  <v-tooltip location="bottom" class="text-center" :disabled="hasAuthorization">
    <template v-slot:activator="{ props }">
      <div v-bind="props">
        <v-btn
          :disabled="!hasAuthorization"
          color="red darken-1"
          variant="outlined"
          data-test="leave-dialog-btn"
          @click="open()"
        >
          Leave Namespace
        </v-btn>
      </div>
    </template>
    <span> You don't have this kind of authorization. </span>
  </v-tooltip>

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
          @click="leave()"
        >
          Leave
        </v-btn>
      </v-card-actions>
    </v-card>
  </v-dialog>
</template>

<script setup lang="ts">
import { ref, computed } from "vue";
import { useStore } from "../../store";
import hasPermission from "../../utils/permission";
import { actions, authorizer } from "../../authorizer";
import {
  INotificationsError,
  INotificationsSuccess,
} from "../../interfaces/INotifications";
import handleError from "@/utils/handleError";

const store = useStore();
const dialog = ref(false);
const tenant = computed(() => localStorage.getItem("tenant"));

const open = () => {
  dialog.value = true;
};

const hasAuthorization = computed(() => {
  const role = store.getters["auth/role"];
  if (role !== "") {
    return hasPermission(
      authorizer.role[role],
      actions.namespace.leave,
    );
  }
  return false;
});

const leave = async () => {
  try {
    dialog.value = !dialog.value;

    await store.dispatch("namespaces/leave", tenant.value);

    window.location.reload();

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
</script>
