<template>
  <div><slot :open-dialog /></div>

  <MessageDialog
    v-model="showDialog"
    title="Accept Invitation"
    :description="dialogDescription"
    icon="mdi-account-check"
    icon-color="primary"
    confirm-text="Accept"
    cancel-text="Cancel"
    :data-test="dataTest"
    @confirm="handleAccept"
    @cancel="showDialog = false"
    @close="showDialog = false"
  />
</template>

<script setup lang="ts">
import { computed, ref } from "vue";
import MessageDialog from "@/components/Dialogs/MessageDialog.vue";
import useInvitationsStore from "@/store/modules/invitations";
import useAuthStore from "@/store/modules/auth";
import useNamespacesStore from "@/store/modules/namespaces";
import useSnackbar from "@/helpers/snackbar";
import handleError from "@/utils/handleError";

interface Props {
  tenant: string;
  namespaceName?: string;
  role?: string;
  dataTest?: string;
  onSuccess?: () => void | Promise<void>;
}

const props = defineProps<Props>();

const showDialog = ref(false);
const invitationsStore = useInvitationsStore();
const authStore = useAuthStore();
const namespacesStore = useNamespacesStore();
const snackbar = useSnackbar();

const dialogDescription = computed(() =>
  (props.namespaceName && props.role)
    ? `You are about to accept this invitation to join ${props.namespaceName} as ${props.role}.`
    : "Accepting this invitation will allow you to collaborate with the namespace collaborators.",
);
const openDialog = () => { showDialog.value = true; };

const handleAccept = async () => {
  try {
    await invitationsStore.acceptInvitation(props.tenant);

    snackbar.showSuccess("Invitation accepted successfully");
    showDialog.value = false;

    await authStore.enterInvitedNamespace(props.tenant);
    await namespacesStore.fetchNamespaceList();

    if (props.onSuccess) { await props.onSuccess(); }
  } catch (error: unknown) {
    snackbar.showError("Failed to accept invitation");
    handleError(error);
  }
};

defineExpose({ openDialog });
</script>
