<template>
  <MessageDialog
    v-if="canSubscribeToBilling"
    v-model="showWarningDialog"
    transition="dialog-bottom-transition"
    data-test="billing-warning-dialog"
    title="Maximum Device Limit Reached"
    description="It seems that your current free account has reached the maximum number of devices allowed in this namespace.
    With a subscription, you can easily add and manage more devices within your account, granting you the flexibility and
    freedom to scale as needed."
    icon="mdi-alert-circle"
    icon-color="warning"
    confirm-text="Go to Billing"
    confirm-color="primary"
    cancel-text="Close"
    confirm-data-test="go-to-billing-btn"
    cancel-data-test="close-btn"
    @confirm="goToBilling"
    @cancel="close"
    @close="close"
  />
</template>

<script setup lang="ts">
import { useRouter } from "vue-router";
import hasPermission from "@/utils/permission";
import MessageDialog from "../MessageDialog.vue";

const router = useRouter();
const showWarningDialog = defineModel({ default: false });
const canSubscribeToBilling = hasPermission("billing:subscribe");

const close = () => {
  showWarningDialog.value = false;
};

const goToBilling = () => {
  router.push("/settings/billing");
  close();
};
</script>
