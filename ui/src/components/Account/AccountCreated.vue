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

    <div class="mt-2 mb-0 text-center" data-test="accountCreated-message">
      Thank you for registering an account on ShellHub.
      <p v-if="isNormalMessage" data-test="accountCreated-normal-message">
        An email was sent with a confirmation link. You need to click on the link to activate your account.
      </p>
      <p v-else data-test="accountCreated-sig-message">
        You will be redirected in 5 seconds, if you weren't redirected, please click the button below.
      </p>
    </div>

    <p v-if="isNormalMessage" class="text-caption mt-2 mb-0" data-test="accountCreated-email-info">
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
import { computed, PropType, watch } from "vue";
import { useRouter, useRoute } from "vue-router";
import { INotificationsSuccess } from "../../interfaces/INotifications";
import { useStore } from "../../store";
import handleError from "@/utils/handleError";

const props = defineProps({
  messageKind: {
    type: String as PropType<"sig" | "normal">,
    required: true,
  },
  show: {
    type: Boolean,
    required: true,
  },
  username: {
    type: String,
    required: true,
  },
});
const emit = defineEmits(["show"]);

const store = useStore();
const router = useRouter();
const route = useRoute();

const token = computed(() => store.getters["users/getSignToken"]);

const showMessage = computed({
  get: () => props.show,
  set: (value: boolean) => emit("show", value),
});

const isNormalMessage = computed(() => props.messageKind === "normal");

const buttonText = computed(() => (isNormalMessage.value ? "RESEND EMAIL" : "REDIRECT"));
const buttonDataTest = computed(() => (isNormalMessage.value ? "resendEmail-btn" : "redirect-btn"));

const resendEmail = async () => {
  try {
    await store.dispatch("users/resendEmail", props.username);
    store.dispatch("snackbar/showSnackbarSuccessAction", INotificationsSuccess.resendEmail);
  } catch (error) {
    store.dispatch("snackbar/showSnackbarErrorDefault");
    handleError(error);
  }
};

const redirect = async () => {
  try {
    store.commit("namespaces/setShowNamespaceInvite", true);
    await router.push({ name: "Home", query: route.query });
  } catch (error) {
    handleError(error);
  }
};

const handleAction = async () => {
  if (isNormalMessage.value) {
    resendEmail();
  } else {
    redirect();
  }
};

watch(showMessage, (newValue) => {
  if (newValue && props.messageKind === "sig") {
    store.commit("auth/setToken", token.value);
    setTimeout(redirect, 5000);
  }
});

defineExpose({ resendEmail, redirect, showMessage });
</script>
