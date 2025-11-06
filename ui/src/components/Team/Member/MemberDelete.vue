<template>
  <v-list-item
    :disabled="!hasAuthorization"
    @click="showDialog = true"
  >
    <div class="d-flex align-center">
      <div class="mr-2">
        <v-icon icon="mdi-delete" />
      </div>
      <v-list-item-title data-test="member-delete-dialog-btn">
        Remove
      </v-list-item-title>
    </div>
  </v-list-item>

  <MessageDialog
    v-model="showDialog"
    title="Are you sure?"
    description="You are about to remove this user from the namespace. If needed, you can re-invite this user to the namespace at any time."
    icon="mdi-account-minus"
    icon-color="error"
    confirm-text="Remove"
    confirm-color="error"
    :confirm-loading="isLoading"
    cancel-text="Close"
    confirm-data-test="member-delete-remove-btn"
    cancel-data-test="member-delete-close-btn"
    data-test="member-delete-card"
    @close="showDialog = false"
    @confirm="remove"
    @cancel="showDialog = false"
  />
</template>

<script setup lang="ts">
import { ref } from "vue";
import handleError from "@/utils/handleError";
import useSnackbar from "@/helpers/snackbar";
import MessageDialog from "@/components/Dialogs/MessageDialog.vue";
import { INamespaceMember } from "@/interfaces/INamespace";
import useAuthStore from "@/store/modules/auth";
import useNamespacesStore from "@/store/modules/namespaces";

const props = defineProps<{
  member: INamespaceMember;
  hasAuthorization: boolean;
}>();

const emit = defineEmits(["update"]);
const showDialog = ref(false);
const isLoading = ref(false);
const authStore = useAuthStore();
const namespacesStore = useNamespacesStore();
const snackbar = useSnackbar();

const update = () => {
  emit("update");
  showDialog.value = false;
};

const remove = async () => {
  isLoading.value = true;
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
  } finally {
    isLoading.value = false;
  }
};
</script>
