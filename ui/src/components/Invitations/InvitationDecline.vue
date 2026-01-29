<template>
  <div><slot :open-dialog /></div>

  <MessageDialog
    v-model="showDialog"
    title="Decline Invitation"
    :description="dialogDescription"
    icon="mdi-account-remove"
    icon-color="error"
    confirm-text="Decline"
    confirm-color="error"
    cancel-text="Cancel"
    :data-test="dataTest"
    @confirm="handleDecline"
    @cancel="showDialog = false"
    @close="showDialog = false"
  />
</template>

<script setup lang="ts">
import { computed, ref } from "vue";
import MessageDialog from "@/components/Dialogs/MessageDialog.vue";
import useInvitationsStore from "@/store/modules/invitations";
import useSnackbar from "@/helpers/snackbar";
import handleError from "@/utils/handleError";

interface Props {
  tenant: string;
  namespaceName?: string;
  dataTest?: string;
  onSuccess?: () => void | Promise<void>;
}

const props = defineProps<Props>();

const showDialog = ref(false);
const invitationsStore = useInvitationsStore();
const snackbar = useSnackbar();

const dialogDescription = computed(() =>
  props.namespaceName
    ? `You are about to decline this invitation to join ${props.namespaceName}. This action cannot be undone.`
    : "You are about to decline this invitation. This action cannot be undone.",
);
const openDialog = () => { showDialog.value = true; };

const handleDecline = async () => {
  try {
    await invitationsStore.declineInvitation(props.tenant);
    snackbar.showSuccess("Invitation declined successfully");
    showDialog.value = false;

    if (props.onSuccess) { await props.onSuccess(); }
  } catch (error: unknown) {
    snackbar.showError("Failed to decline invitation");
    handleError(error);
  }
};
</script>
