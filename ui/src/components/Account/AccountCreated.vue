<template>
  <v-card
    v-if="showMessage"
    color="transparent"
    class="elevation-0 bg-v-theme-surface mt-2 pt-0"
    data-test="accountCreated-card"
  >
    <v-card-title class="text-h5 pt-0 text-center">
      Account Creation Successful
    </v-card-title>

    <div
      class="mt-2 mb-0 text-center"
      data-test="accountCreated-message"
    >
      Thank you for registering an account on ShellHub.
      <p
        v-if="isNormalMessage"
        data-test="accountCreated-normal-message"
      >
        An email was sent with a confirmation link. You need to click on the link to activate your account.
      </p>
      <p
        v-else
        data-test="accountCreated-sig-message"
      >
        You will be redirected in 5 seconds, if you weren't redirected, please click the button below.
      </p>
    </div>

    <p
      v-if="isNormalMessage"
      class="text-caption mt-2 mb-0"
      data-test="accountCreated-email-info"
    >
      If you haven't received the email, click on the button.
    </p>

    <v-card-actions class="justify-center">
      <v-btn
        type="submit"
        color="primary"
        :data-test="buttonDataTest"
        @click="handleAction"
      >
        {{ buttonText }}
      </v-btn>
    </v-card-actions>
  </v-card>
</template>

<script setup lang="ts">
import { computed, watch } from "vue";
import { useRouter, useRoute } from "vue-router";
import handleError from "@/utils/handleError";
import useSnackbar from "@/helpers/snackbar";
import useAuthStore from "@/store/modules/auth";
import useUsersStore from "@/store/modules/users";

const props = defineProps<{
  messageKind: "sig" | "normal";
  show: boolean;
  username: string;
}>();

const emit = defineEmits(["show"]);
const authStore = useAuthStore();
const usersStore = useUsersStore();
const router = useRouter();
const route = useRoute();
const snackbar = useSnackbar();

const token = computed(() => usersStore.signUpToken);

const showMessage = computed({
  get: () => props.show,
  set: (value: boolean) => emit("show", value),
});

const isNormalMessage = computed(() => props.messageKind === "normal");

const buttonText = computed(() => (isNormalMessage.value ? "RESEND EMAIL" : "REDIRECT"));
const buttonDataTest = computed(() => (isNormalMessage.value ? "resendEmail-btn" : "redirect-btn"));

const resendEmail = async () => {
  try {
    await usersStore.resendEmail(props.username);
    snackbar.showSuccess("Email successfully sent.");
  } catch (error) {
    snackbar.showError("Failed to send email.");
    handleError(error);
  }
};

const redirect = async () => {
  try {
    // eslint-disable-next-line @typescript-eslint/no-unused-vars
    const { redirect, ...cleanedQuery } = route.query;

    // Perform the redirect to accept-invite with the cleaned query parameters
    await router.push({ name: "AcceptInvite", query: cleanedQuery });
  } catch (error) {
    handleError(error);
  }
};

const handleAction = async () => {
  if (isNormalMessage.value) await resendEmail();
  else await redirect();
};

watch(showMessage, (newValue) => {
  if (newValue && props.messageKind === "sig") {
    authStore.token = token.value as string;
    setTimeout(() => { void redirect(); }, 5000);
  }
});

defineExpose({ resendEmail, redirect, showMessage });
</script>
