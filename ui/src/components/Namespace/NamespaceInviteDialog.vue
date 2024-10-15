<template>
  <v-dialog
    transition="dialog-bottom-transition"
    width="700"
    @click:outside="close();"
    v-model="showDialog"
  >
    <v-card class="bg-v-theme-surface" data-test="card-dialog">
      <v-card-title class="text-h5 pa-4 bg-primary" data-test="dialog-title">
        {{ modalTitle }}
      </v-card-title>
      <v-container data-test="dialog-content">
        <v-row class="mb-2" data-test="dialog-row">
          <v-col data-test="dialog-col">
            <h4 data-test="dialog-message">{{ modalMessage }}</h4>
          </v-col>
        </v-row>
        <v-card-actions v-if="!modalError" data-test="dialog-actions">
          <v-btn
            variant="text"
            color="error"
            data-test="decline-btn"
            @click="declineInvite()"
          >
            Decline Invitation
          </v-btn>
          <v-spacer data-test="dialog-spacer" />

          <v-btn
            variant="text"
            color="primary"
            data-test="accept-btn"
            @click="acceptInvite()"
          >
            Accept Invitation
          </v-btn>
        </v-card-actions>
        <v-card-actions v-else data-test="error-dialog-actions">
          <v-spacer data-test="error-dialog-spacer" />
          <v-btn
            variant="text"
            color="error"
            data-test="close-btn"
            @click="close()"
          >
            Close
          </v-btn>
        </v-card-actions>
      </v-container>
    </v-card>
  </v-dialog>
</template>

<script setup lang="ts">
import { ref, watch } from "vue";
import axios from "axios";
import { useRouter, useRoute } from "vue-router";
import { useStore } from "@/store";

const store = useStore();
const router = useRouter();
const route = useRoute();
const emit = defineEmits(["close"]);

const modalTitle = ref("You've Been Invited to Join a Namespace");
// eslint-disable-next-line vue/max-len
const modalMessage = ref("Accepting this invitation will allow you to collaborate with the Namespace collaborators. Please choose whether to accept or decline this invitation.");
const modalError = ref(false);
const buttonText = ref("Close");
const showDialog = ref(store.getters["namespaces/showNamespaceInviteDialog"]);

watch(
  () => store.getters["namespaces/showNamespaceInviteDialog"],
  (newValue) => {
    showDialog.value = newValue;
  },
);
const close = async () => {
  store.commit("namespaces/setShowNamespaceInvite", false);
  emit("close");
  await router.replace({ query: {} });
};

const errorToMessage = (status?: number): string => {
  const errorMessages: Record<number, string> = {
    400: "It seems like there was an issue with the request. Please check the invitation link and try again.",
    // eslint-disable-next-line vue/max-len
    403: "The token provided appears to be invalid or not associated with your account. Please verify your credentials and try again later.",
    404: "We couldn't find the namespace or member associated with this invitation. The invitation might have expired.",
    500: "Our servers encountered an issue while processing your invitation acceptance. Please try again later.",
  };

  return status ? errorMessages[status] ?? "An unexpected error occurred. Please try again later."
    : "An unexpected error occurred. Please try again later.";
};

const handleInviteError = (error: unknown) => {
  modalError.value = true;
  modalTitle.value = "Invite Accept Error";

  if (axios.isAxiosError(error)) {
    modalMessage.value = errorToMessage(error.response?.status);
  } else {
    modalMessage.value = "An unexpected error occurred. Please try again later.";
  }

  buttonText.value = "Close";
};

const acceptInvite = async () => {
  try {
    const tenant = (route.query["tenant-id"] || route.query.tenantid) as string;
    const sig = route.query.sig as string;

    await store.dispatch("namespaces/acceptInvite", { tenant, sig });

    modalMessage.value = "Your invitation has been successfully accepted! You are now a member of the namespace.";
    buttonText.value = "Switch to New Namespace";

    await store.dispatch("namespaces/switchNamespace", { tenant_id: tenant });
    await store.dispatch("namespaces/fetch", { page: 1, perPage: 10, filter: "" });

    close();
  } catch (error) {
    handleInviteError(error);
  }
};

const declineInvite = close;

defineExpose({
  modalTitle,
  modalMessage,
  modalError,
  close,
  acceptInvite,
  declineInvite,
});
</script>
