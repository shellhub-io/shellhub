<template>
  <BaseDialog v-model="showDialog">
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

        <v-btn variant="text" data-test="close-btn" @click="showDialog = !showDialog">
          Close
        </v-btn>

        <v-btn
          color="red darken-1"
          variant="text"
          data-test="leave-btn"
          :disabled="!hasAuthorization"
          @click="leave()"
        >
          Leave
        </v-btn>
      </v-card-actions>
    </v-card>
  </BaseDialog>
</template>

<script setup lang="ts">
import { computed } from "vue";
import { useRouter } from "vue-router";
import hasPermission from "@/utils/permission";
import { actions, authorizer } from "@/authorizer";
import handleError from "@/utils/handleError";
import useSnackbar from "@/helpers/snackbar";
import BaseDialog from "../BaseDialog.vue";
import useAuthStore from "@/store/modules/auth";
import useNamespacesStore from "@/store/modules/namespaces";

const namespacesStore = useNamespacesStore();
const authStore = useAuthStore();
const router = useRouter();
const snackbar = useSnackbar();
const showDialog = defineModel({ default: false });
const tenant = computed(() => localStorage.getItem("tenant") as string);

const hasAuthorization = computed(() => {
  const { role } = authStore;
  return !!role && hasPermission(authorizer.role[role], actions.namespace.leave);
});

const leave = async () => {
  try {
    await namespacesStore.leaveNamespace(tenant.value);
    showDialog.value = false;
    snackbar.showSuccess("You have left the namespace.");
    router.go(0);
  } catch (error: unknown) {
    snackbar.showError("Failed to leave the namespace.");
    handleError(error);
  }
};

defineExpose({ showDialog });
</script>
