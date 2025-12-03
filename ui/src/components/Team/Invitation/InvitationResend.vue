<template>
  <div>
    <v-list-item
      :disabled="!hasAuthorization || isDisabled"
      data-test="invitation-resend-btn"
      @click="showDialog = true"
    >
      <div class="ga-2 d-flex align-center">
        <v-icon icon="mdi-email-fast" />
        <v-list-item-title data-test="invitation-resend-title">Resend</v-list-item-title>
      </div>
    </v-list-item>

    <MessageDialog
      v-model="showDialog"
      title="Resend Invitation"
      :description="`Are you sure you want to resend the invitation to ${invitation.user.email}?`"
      icon="mdi-email-fast"
      icon-color="primary"
      confirm-text="Resend Invitation"
      confirm-color="primary"
      :confirm-loading="isLoading"
      cancel-text="Close"
      confirm-data-test="resend-invitation-btn"
      cancel-data-test="close-btn"
      data-test="invitation-resend-dialog"
      @close="close"
      @confirm="resendInvitation"
      @cancel="close"
    />
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from "vue";
import axios from "axios";
import MessageDialog from "@/components/Dialogs/MessageDialog.vue";
import { IInvitation } from "@/interfaces/IInvitation";
import handleError from "@/utils/handleError";
import useSnackbar from "@/helpers/snackbar";
import useInvitationsStore from "@/store/modules/invitations";
import { isInvitationExpired } from "@/utils/invitations";

const props = defineProps<{
  invitation: IInvitation;
  hasAuthorization: boolean;
}>();

const emit = defineEmits(["update"]);
const invitationsStore = useInvitationsStore();
const snackbar = useSnackbar();
const showDialog = ref(false);
const isLoading = ref(false);

const isExpired = computed(() => isInvitationExpired(props.invitation.expires_at));

const isDisabled = computed(() => {
  if (props.invitation.status === "cancelled") return false;
  if (props.invitation.status === "pending" && isExpired.value) return false;
  return true;
});

const close = () => { showDialog.value = false; };

const update = () => {
  emit("update");
  close();
};

const handleResendInvitationError = (error: unknown) => {
  if (axios.isAxiosError(error)) {
    const status = error.response?.status;
    switch (status) {
      case 400:
        snackbar.showError("Invalid invitation.");
        break;
      case 403:
        snackbar.showError("You don't have permission to send invitations.");
        break;
      case 404:
        snackbar.showError("Invitation not found.");
        break;
      case 409:
        snackbar.showError("This user is already invited or is a member of this namespace.");
        break;
      default:
        snackbar.showError("Failed to resend invitation.");
    }
  } else {
    snackbar.showError("Failed to resend invitation.");
  }

  handleError(error);
};

const resendInvitation = async () => {
  isLoading.value = true;
  try {
    await invitationsStore.sendInvitationEmail({
      tenant_id: props.invitation.namespace.tenant_id,
      email: props.invitation.user.email,
      role: props.invitation.role,
    });

    snackbar.showSuccess("Successfully resent invitation email.");
    update();
  } catch (error: unknown) {
    handleResendInvitationError(error);
  } finally {
    isLoading.value = false;
  }
};
</script>
