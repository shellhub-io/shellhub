<template>
  <v-dialog
    v-if="hasAuthorization"
    v-model="showMessage"
    transition="dialog-bottom-transition"
    width="520"
    data-test="billingWarning-dialog"
  >
    <v-card class="bg-v-theme-surface">
      <v-card-title class="pa-3 bg-primary">
        Update account
      </v-card-title>

      <v-card-text class="mt-4 mb-3 pb-1">
        <p>
          This namespace has maximum number of devices on your free account.
        </p>

        <p>
          If you create the subscription in your account settings, you can
          continue to take advantage of the features available on ShellHub by
          adding more devices.
        </p>
      </v-card-text>

      <v-card-actions>
        <v-spacer />

        <v-btn text data-test="close-btn" @click="close()"> Close </v-btn>

        <v-btn
          to="/settings/billing"
          text
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
