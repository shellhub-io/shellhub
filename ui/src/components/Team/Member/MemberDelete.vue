<template>
  <v-list-item @click="showDialog = true" :disabled="!hasAuthorization">
    <div class="d-flex align-center">
      <div class="mr-2">
        <v-icon> mdi-delete </v-icon>
      </div>

      <v-list-item-title data-test="member-delete-dialog-btn">
        Remove
      </v-list-item-title>
    </div>
  </v-list-item>

  <BaseDialog v-model="showDialog">
    <v-card class="bg-v-theme-surface" data-test="member-delete-card">
      <v-card-title class="text-h5 pa-5 bg-primary" data-test="member-delete-dialog-title">
        Are you sure?
      </v-card-title>
      <v-divider />

      <v-card-text class="mt-4 mb-0 pb-1" data-test="member-delete-dialog-text">
        <p class="text-body-2 mb-2">
          You are about to remove this user from the namespace.
        </p>

        <p class="text-body-2 mb-2">
          If needed, you can re-invite this user to the namespace at any time.
        </p>
      </v-card-text>

      <v-card-actions>
        <v-spacer />

        <v-btn variant="text" @click="showDialog = false" data-test="member-delete-close-btn"> Close </v-btn>

        <v-btn color="red darken-1" variant="text" @click="remove()" data-test="member-delete-remove-btn">
          Remove
        </v-btn>
      </v-card-actions>
    </v-card>
  </BaseDialog>
</template>

<script setup lang="ts">
import { ref } from "vue";
import handleError from "@/utils/handleError";
import useSnackbar from "@/helpers/snackbar";
import BaseDialog from "@/components/BaseDialog.vue";
import { INamespaceMember } from "@/interfaces/INamespace";
import useAuthStore from "@/store/modules/auth";
import useNamespacesStore from "@/store/modules/namespaces";

const props = defineProps<{
  member: INamespaceMember;
  hasAuthorization: boolean;
}>();

const emit = defineEmits(["update"]);
const showDialog = ref(false);
const authStore = useAuthStore();
const namespacesStore = useNamespacesStore();
const snackbar = useSnackbar();

const update = () => {
  emit("update");
  showDialog.value = false;
};

const remove = async () => {
  try {
    await namespacesStore.removeMemberFromNamespace({
      user_id: props.member.id,
      tenant_id: authStore.tenantId,
    });

    update();
    snackbar.showSuccess("Successfully removed user from namespace.");
  } catch (error: unknown) {
    snackbar.showError("Failed to remove user from namespace.");
    handleError(error);
  }
};
</script>
