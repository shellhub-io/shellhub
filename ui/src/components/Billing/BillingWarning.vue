<template>
  <BaseDialog
    v-if="hasAuthorization"
    v-model="showWarningDialog"
    transition="dialog-bottom-transition"
    data-test="billing-warning-dialog"
  >
    <v-card class="bg-v-theme-surface" data-test="card-dialog">
      <v-card-title class="pa-3 bg-primary" data-test="card-title">
        Maximum Device Limit Reached
      </v-card-title>

      <v-card-text class="mt-4 mb-3 pb-1" data-test="card-text">
        <p class="mb-2">
          It seems that your current free account has reached the maximum number of devices allowed in this namespace.
        </p>
        <p class="mb-2">
          With a subscription, you can easily add and manage more devices within your account,
          granting you the flexibility and freedom to scale as needed.
        </p>
      </v-card-text>

      <v-card-actions>
        <v-spacer />

        <v-btn variant="text" data-test="close-btn" @click="close()"> Close </v-btn>

        <v-btn
          to="/settings/billing"
          variant="text"
          color="primary"
          data-test="go-to-billing-btn"
          @click="close()"
        >
          Go to Billing
        </v-btn>
      </v-card-actions>
    </v-card>
  </BaseDialog>
</template>

<script setup lang="ts">
import { computed } from "vue";
import { actions, authorizer } from "@/authorizer";
import hasPermission from "@/utils/permission";
import { useStore } from "@/store";
import BaseDialog from "../BaseDialog.vue";
import useAuthStore from "@/store/modules/auth";
import useBillingStore from "@/store/modules/billing";

const store = useStore();
const authStore = useAuthStore();
const billingStore = useBillingStore();

const hasAuthorization = computed(() => {
  const { role } = authStore;
  return !!role && hasPermission(authorizer.role[role], actions.billing.subscribe);
});

const close = () => {
  if (store.getters["users/statusUpdateAccountDialog"]) {
    store.dispatch("users/setStatusUpdateAccountDialog", false);
  } else if (
    store.getters["users/statusUpdateAccountDialogByDeviceAction"]
  ) {
    store.dispatch(
      "users/setStatusUpdateAccountDialogByDeviceAction",
      false,
    );
  }
};

const showWarningDialog = computed({
  get() {
    return (
      (store.getters["users/statusUpdateAccountDialog"]
        && store.getters["stats/stats"].registered_devices === 3
        && !billingStore.isActive)
        || store.getters["users/statusUpdateAccountDialogByDeviceAction"]
    );
  },
  set() {
    close();
  },
});
</script>
