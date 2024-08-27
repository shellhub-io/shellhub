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
import { ref, computed } from "vue";
import axios, { AxiosError } from "axios";
import { useRouter, useRoute } from "vue-router";
import { useStore } from "@/store";

const store = useStore();
const router = useRouter();
const route = useRoute();
const modalTitle = ref<string>("You've Been Invited to Join a Namespace");
// eslint-disable-next-line vue/max-len
const modalMessage = ref<string>("Accepting this invitation will allow you to collaborate with the Namespace collaborators. Please choose whether to accept or decline this invitation.");
const modalError = ref<boolean>(false);
const buttonText = ref("Close");
const showDialog = computed(() => store.getters["namespaces/showNamespaceInviteDialog"]);

const close = async () => {
  store.commit("namespaces/setShowNamespaceInvite", false);
  await router.push({ path: "/" }).then(() => {
    window.location.reload();
  });
};

const acceptInvite = async () => {
  try {
    await store.dispatch("namespaces/acceptInvite", {
      tenant: route.query["tenant-id"] as string || route.query.tenantid as string,
      sig: route.query.sig as string,
    });
    // eslint-disable-next-line vue/max-len
    modalMessage.value = "Your invitation has been successfully accepted! You are now a member of the namespace.";
    buttonText.value = "Switch to New Namespace";
    await store.dispatch("namespaces/switchNamespace", {
      tenant_id: route.query["tenant-id"] as string || route.query.tenantid as string,
    });
    await store.dispatch("namespaces/fetch", {
      page: 1,
      perPage: 10,
      filter: "",
    });
    close();
  } catch (error: unknown) {
    modalError.value = true;
    modalTitle.value = "Invite Accept Error";

    if (axios.isAxiosError(error)) {
      const axiosError = error as AxiosError;
      switch (axiosError.response?.status) {
        case 400:
          modalMessage.value = "It seems like there was an issue with the request. Please check the invitation link and try again.";
          break;
        case 403:
          // eslint-disable-next-line vue/max-len
          modalMessage.value = "The token provided appears to be invalid or not associated with your account. Please verify your credentials and try again later.";
          break;
        case 404:
          // eslint-disable-next-line vue/max-len
          modalMessage.value = "We couldn't find the namespace or member associated with this invitation. The invitation might have expired.";
          break;
        case 500:
          modalMessage.value = "Our servers encountered an issue while processing your invitation acceptance. Please try again later.";
          break;
        default:
          modalMessage.value = "An unexpected error occurred. Please try again later.";
          break;
      }
    }
    buttonText.value = "Close";
  }
};

const declineInvite = () => {
  close();
};

defineExpose({ modalTitle, close, acceptInvite, declineInvite, modalMessage, modalError });
</script>
