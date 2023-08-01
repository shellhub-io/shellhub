<template>
  <v-dialog
    v-if="hasAuthorization"
    v-model="showMessage"
    transition="dialog-bottom-transition"
    width="650"
    data-test="billingWarning-dialog"
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
          data-test="goToBilling-btn"
          @click="close()"
        >
          Go to Billing
        </v-btn>
      </v-card-actions>
    </v-card>
  </v-dialog>
</template>

<script lang="ts">
import { defineComponent, computed } from "vue";
import { actions, authorizer } from "../../authorizer";
import hasPermission from "../../utils/permission";
import { useStore } from "../../store";

export default defineComponent({
  setup() {
    const store = useStore();

    const hasAuthorization = computed(() => {
      const role = store.getters["auth/role"];
      if (role !== "") {
        return hasPermission(
          authorizer.role[role],
          actions.billing.subscribe,
        );
      }

      return false;
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

    const showMessage = computed({
      get() {
        return (
          (store.getters["users/statusUpdateAccountDialog"]
            && store.getters["stats/stats"].registered_devices === 3
            && !store.getters["billing/active"])
          || store.getters["users/statusUpdateAccountDialogByDeviceAction"]
        );
      },
      set() {
        close();
      },
    });

    return {
      hasAuthorization,
      showMessage,
      close,
    };
  },
});
</script>
