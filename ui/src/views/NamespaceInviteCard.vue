<template>
  <v-alert
    v-if="errorAlert"
    type="error"
    variant="tonal"
    data-test="error-alert"
    :text="errorAlert"
  />
  <v-card-title
    class="d-flex justify-center mb-1 text-h6"
    data-test="title"
  >
    Namespace Invitation
  </v-card-title>
  <v-card-text
    class="text-subtitle-1 ml-3 text-justify"
    data-test="message"
  >
    {{ message }}
  </v-card-text>
  <v-card-actions data-test="actions">
    <v-spacer data-test="spacer" />
    <InvitationDecline
      v-slot="{ openDialog }"
      :tenant="tenant"
      data-test="decline-dialog"
      :on-success="handleDeclineSuccess"
    >
      <v-btn
        variant="text"
        color="error"
        data-test="decline-btn"
        :text="isUserValid ? 'Decline' : 'Back to Home Page'"
        @click="isUserValid ? openDialog() : redirectToHome()"
      />
    </InvitationDecline>
    <InvitationAccept
      v-slot="{ openDialog }"
      :tenant="tenant"
      data-test="accept-dialog"
      :on-success="handleAcceptSuccess"
    >
      <v-btn
        variant="text"
        color="primary"
        data-test="accept-btn"
        text="Accept"
        :disabled="!isUserValid"
        @click="openDialog"
      />
    </InvitationAccept>
  </v-card-actions>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from "vue";
import { useRouter, useRoute } from "vue-router";
import InvitationAccept from "@/components/Invitations/InvitationAccept.vue";
import InvitationDecline from "@/components/Invitations/InvitationDecline.vue";

const router = useRouter();
const route = useRoute();
const message = ref("Accepting this invitation will allow you to collaborate with the Namespace collaborators.");
const errorAlert = ref("");
const isUserValid = computed(() => localStorage.getItem("id") === route.query["user-id"]);
const tenant = computed(() => (route.query["tenant-id"] || route.query.tenantid) as string);

const redirectToHome = async () => {
  await router.replace({ query: {} });
  await router.push({ name: "Home" });
};

const handleAcceptSuccess = async () => {
  message.value = "Your invitation has been successfully accepted! You are now a member of the namespace.";
  await redirectToHome();
};

const handleDeclineSuccess = async () => { await redirectToHome(); };

onMounted(() => {
  if (isUserValid.value) return;
  errorAlert.value = "You aren't logged in the account meant for this invitation.";
});
</script>
