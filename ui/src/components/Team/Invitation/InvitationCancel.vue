<template>
  <div>
    <v-list-item
      :disabled="!hasAuthorization"
      data-test="invitation-cancel-btn"
      @click="showDialog = true"
    >
      <div class="d-flex align-center ga-2">
        <v-icon icon="mdi-close-circle" />
        <v-list-item-title data-test="invitation-cancel-title">Cancel</v-list-item-title>
      </div>
    </v-list-item>

    <MessageDialog
      v-model="showDialog"
      title="Cancel Invitation"
      :description="`Are you sure you want to cancel the invitation for ${invitation.user.email}?`"
      icon="mdi-alert"
      icon-color="error"
      confirm-text="Cancel Invitation"
      confirm-color="error"
      :confirm-loading="isLoading"
      cancel-text="Close"
      confirm-data-test="cancel-invitation-btn"
      cancel-data-test="close-btn"
      data-test="invitation-cancel-dialog"
      @close="close"
      @confirm="cancelInvitation"
      @cancel="close"
    />
  </div>
</template>

<script setup lang="ts">
import { ref } from "vue";
import axios from "axios";
import MessageDialog from "@/components/Dialogs/MessageDialog.vue";
import { IInvitation } from "@/interfaces/IInvitation";
import handleError from "@/utils/handleError";
import useSnackbar from "@/helpers/snackbar";
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

const close = () => { showDialog.value = false; };

const update = () => {
  emit("update");
  close();
};

const handleCancelInvitationError = (error: unknown) => {
  if (axios.isAxiosError(error)) {
    const status = error.response?.status;
    switch (status) {
      case 400:
        snackbar.showError("Invalid invitation.");
        break;
      case 403:
        snackbar.showError("You don't have permission to cancel invitations.");
        break;
      case 404:
        snackbar.showError("Invitation not found.");
        break;
      default:
        snackbar.showError("Failed to cancel invitation.");
    }
  } else {
    snackbar.showError("Failed to cancel invitation.");
  }

  handleError(error);
};

const cancelInvitation = async () => {
  isLoading.value = true;
  try {
    await invitationsStore.cancelInvitation({
      tenant: invitation.namespace.tenant_id,
      user_id: invitation.user.id,
    });

    snackbar.showSuccess("Successfully cancelled invitation.");
    update();
  } catch (error: unknown) {
    handleCancelInvitationError(error);
  } finally {
    isLoading.value = false;
  }
};
</script>
