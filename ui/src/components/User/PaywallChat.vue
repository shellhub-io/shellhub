<template>
  <MessageDialog
    v-model="showDialog"
    title="Upgrade to have access to chat support!"
    icon="mdi-chat-question"
    icon-color="success"
    cancel-text="Close"
    cancel-data-test="close-btn"
    confirm-text="Upgrade"
    confirm-data-test="upgrade-btn"
    @confirm="redirectToPricing"
    @cancel="showDialog = false"
    data-test="paywall-chat-dialog"
  >
    <p class="text-center mb-4" data-test="upgrade-description-1">
      Get real-time assistance from our team with priority responses.
      Skip the hassle of documentation — upgrade now and unlock direct chat support.
    </p>
    <p class="text-center text-grey" data-test="upgrade-description-2">
      However, you can still use
      <a
        href="https://docs.shellhub.io/"
        target="_blank"
        rel="noopener noreferrer"
        data-test="link-anchor"
      >
        our Documentation
      </a>
      to find answers to your questions and troubleshoot issues on your own.
    </p>
  </MessageDialog>
</template>

<script setup lang="ts">
import MessageDialog from "../MessageDialog.vue";

const showDialog = defineModel({ default: false });

const redirectToPricing = () => {
  window.open("https://www.shellhub.io/pricing", "_blank", "noopener,noreferrer");
};

defineExpose({ showDialog });
</script>
