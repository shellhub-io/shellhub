<template>
  <div>
    <v-list-item
      :disabled="!hasAuthorization"
      data-test="invitation-edit-btn"
      @click="showDialog = true"
    >
      <div class="d-flex align-center ga-2">
        <v-icon icon="mdi-pencil" />
        <v-list-item-title data-test="invitation-edit-title">Edit Role</v-list-item-title>
      </div>
    </v-list-item>

    <FormDialog
      v-model="showDialog"
      title="Update invitation role"
      icon="mdi-account-edit"
      confirm-text="Update"
      :confirm-loading="isLoading"
      cancel-text="Cancel"
      confirm-data-test="update-btn"
      cancel-data-test="cancel-btn"
      data-test="invitation-edit-dialog"
      @close="close"
      @confirm="editInvitation"
      @cancel="close"
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
import FormDialog from "@/components/Dialogs/FormDialog.vue";
import { IInvitation } from "@/interfaces/IInvitation";
import handleError from "@/utils/handleError";
import useSnackbar from "@/helpers/snackbar";
import RoleSelect from "@/components/Team/RoleSelect.vue";
import useInvitationsStore from "@/store/modules/invitations";

const { invitation, hasAuthorization } = defineProps<{
  invitation: IInvitation;
  hasAuthorization: boolean;
}>();

const emit = defineEmits(["update"]);
const invitationsStore = useInvitationsStore();
const snackbar = useSnackbar();
const showDialog = ref(false);
const isLoading = ref(false);
const newRole = ref(invitation.role);

const close = () => {
  showDialog.value = false;
  newRole.value = invitation.role;
};

const update = () => {
  emit("update");
  close();
};

const handleEditInvitationError = (error: unknown) => {
  if (axios.isAxiosError(error)) {
    const status = error.response?.status;
    switch (status) {
      case 400:
        snackbar.showError("Invalid invitation or role.");
        break;
      case 403:
        snackbar.showError("You don't have permission to edit invitations.");
        break;
      case 404:
        snackbar.showError("Invitation not found.");
        break;
      default:
        snackbar.showError("Failed to update invitation role.");
    }
  } else {
    snackbar.showError("Failed to update invitation role.");
  }

  handleError(error);
};

const editInvitation = async () => {
  isLoading.value = true;
  try {
    await invitationsStore.editInvitation({
      tenant: invitation.namespace.tenant_id,
      user_id: invitation.user.id,
      role: newRole.value,
    });

    snackbar.showSuccess("Successfully updated invitation role.");
    update();
  } catch (error: unknown) {
    handleEditInvitationError(error);
  } finally {
    isLoading.value = false;
  }
};
</script>
