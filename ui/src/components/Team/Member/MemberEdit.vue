<template>
  <div>
    <v-list-item
      @click="showDialog = true"
      :disabled="!hasAuthorization"
      data-test="member-edit-btn"
    >
      <div class="d-flex align-center">
        <div class="mr-2"><v-icon icon="mdi-pencil" /></div>
        <v-list-item-title data-test="member-edit-title">
          Edit
        </v-list-item-title>
      </div>
    </v-list-item>

    <FormDialog
      v-model="showDialog"
      @close="close"
      @confirm="editMember"
      @cancel="close"
      title="Update member role"
      icon="mdi-account-edit"
      confirm-text="Edit"
      :confirm-loading="isLoading"
      cancel-text="Close"
      confirm-data-test="edit-btn"
      cancel-data-test="close-btn"
      data-test="member-edit-dialog"
    >
      <v-card-text class="pa-6">
        <RoleSelect v-model="newRole" />
      </v-card-text>
    </FormDialog>
  </div>
</template>

<script setup lang="ts">
import { ref } from "vue";
import axios from "axios";
import { BasicRole, INamespaceMember } from "@/interfaces/INamespace";
import handleError from "@/utils/handleError";
import useSnackbar from "@/helpers/snackbar";
import FormDialog from "@/components/FormDialog.vue";
import RoleSelect from "@/components/Team/RoleSelect.vue";
import useAuthStore from "@/store/modules/auth";
import useNamespacesStore from "@/store/modules/namespaces";

const { member, hasAuthorization } = defineProps<{
  member: INamespaceMember;
  hasAuthorization: boolean;
}>();

const emit = defineEmits(["update"]);
const authStore = useAuthStore();
const namespacesStore = useNamespacesStore();
const snackbar = useSnackbar();
const showDialog = ref(false);
const isLoading = ref(false);
const newRole = ref(member.role as BasicRole);

const close = () => { showDialog.value = false; };

const update = () => {
  emit("update");
  close();
};

const handleEditMemberError = (error: unknown) => {
  if (axios.isAxiosError(error)) {
    const status = error.response?.status;
    switch (status) {
      case 400:
        snackbar.showError("The user isn't linked to the namespace.");
        break;
      case 403:
        snackbar.showError("You don't have permission to assign a role to the user.");
        break;
      case 404:
        snackbar.showError("The username doesn't exist.");
        break;
      default:
        snackbar.showError("Failed to update user role.");
    }
  } else snackbar.showError("Failed to update user role.");

  handleError(error);
};

const editMember = async () => {
  isLoading.value = true;
  try {
    await namespacesStore.updateNamespaceMember({
      user_id: member.id,
      tenant_id: authStore.tenantId,
      role: newRole.value,
    });

    snackbar.showSuccess("Successfully updated user role.");
    update();
  } catch (error: unknown) {
    handleEditMemberError(error);
  } finally {
    isLoading.value = false;
  }
};

</script>
