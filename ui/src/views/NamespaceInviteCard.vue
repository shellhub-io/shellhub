<template>
  <v-alert
    v-if="errorAlert"
    type="error"
    variant="tonal"
    data-test="error-alert"
  >
    {{ errorAlert }}
  </v-alert>
  <v-card-title class="d-flex justify-center mb-1 text-h6" data-test="title">
    {{ title }}
  </v-card-title>
  <div class="text-subtitle-1 ml-3" data-test="message">{{ message }}</div>
  <v-card-actions data-test="actions">
    <v-btn
      variant="text"
      color="error"
      data-test="decline-btn"
      @click="close()"
      :text="isUserValid ? 'Decline Invitation' : 'Back to Home Page'"
    />
    <v-spacer data-test="spacer" />
    <v-btn
      variant="text"
      color="primary"
      data-test="accept-btn"
      :disabled="!isUserValid"
      @click="acceptInvite()"
    >
      Accept Invitation
    </v-btn>
  </v-card-actions>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from "vue";
import axios from "axios";
import { useRouter, useRoute } from "vue-router";
import { useStore } from "@/store";
import useAuthStore from "@/store/modules/auth";

const store = useStore();
const authStore = useAuthStore();
const router = useRouter();
const route = useRoute();

const title = ref("Namespace Invitation");
const message = ref("Accepting this invitation will allow you to collaborate with the Namespace collaborators.");
const errorAlert = ref("");

const isUserValid = computed(() => localStorage.getItem("id") === route.query["user-id"]);

const close = async () => {
  await router.replace({ query: {} });
  await router.push({ name: "Home" });
};

const errorToMessage = (status?: number): string => {
  const errorMessages: Record<number, string> = {
    400: "It seems like there was an issue with the request. Please check the invitation link and try again.",
    404: "We couldn't find the namespace or member associated with this invitation. The invitation might have expired.",
    500: "Our servers encountered an issue while processing your invitation acceptance. Please try again later.",
  };

  return status ? errorMessages[status] ?? "An unexpected error occurred. Please try again later."
    : "An unexpected error occurred. Please try again later.";
};

const handleInviteError = (error: unknown) => {
  title.value = "Invite Accept Error";

  if (axios.isAxiosError(error)) {
    errorAlert.value = errorToMessage(error.response?.status);
  } else {
    message.value = "An unexpected error occurred. Please try again later.";
  }
};

const acceptInvite = async () => {
  try {
    const tenantId = (route.query["tenant-id"] || route.query.tenantid) as string;
    const sig = route.query.sig as string;

    await store.dispatch("namespaces/acceptInvite", { tenantId, sig });

    message.value = "Your invitation has been successfully accepted! You are now a member of the namespace.";

    await authStore.enterInvitedNamespace(tenantId);
    await store.dispatch("namespaces/fetch", { page: 1, perPage: 10, filter: "" });
    await router.push({ name: "Home" });
    close();
  } catch (error) {
    handleInviteError(error);
  }
};

onMounted(() => {
  if (isUserValid.value) {
    return;
  }
  errorAlert.value = "You aren't logged in the account meant for this invitation.";
});

defineExpose({
  title,
  message,
  errorAlert,
  close,
  acceptInvite,
  isUserValid,
  handleInviteError,
});
</script>
