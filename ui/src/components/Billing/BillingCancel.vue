<template>
  <v-tooltip location="bottom" class="text-center" :disabled="hasAuthorization">
    <template v-slot:activator="{ props }">
      <div v-bind="props">
        <v-btn
          :disabled="!hasAuthorization"
          color="red darken-1"
          variant="outlined"
          data-test="cancel-btn"
          @click="dialog = !dialog"
        >
          Cancel
        </v-btn>
      </div>
    </template>
    <span> You don't have this kind of authorization. </span>
  </v-tooltip>

  <v-dialog v-model="dialog" width="520" transition="dialog-bottom-transition">
    <v-card class="bg-v-theme-surface">
      <v-card-title class="text-h5 pa-3 bg-primary">
        Are you sure?
      </v-card-title>

      <v-card-text class="mt-4 mb-0 pb-1">
        <p class="text-body-2 mb-2">
          Canceling the subscription will generate an invoice, estimated
          <b> {{ formatCurrency(nextPaymentDue, currency) }} </b> for the time of use.
        </p>
      </v-card-text>

      <v-card-actions>
        <v-spacer />

        <v-btn text data-test="close-btn" @click="dialog = !dialog">
          Close
        </v-btn>

        <v-btn text data-test="cancelDialog-btn" @click="cancelSubscription()">
          Cancel
        </v-btn>
      </v-card-actions>
    </v-card>
  </v-dialog>
</template>

<script lang="ts">
import { defineComponent, ref, computed } from "vue";
import { useRouter } from "vue-router";
import { actions, authorizer } from "../../authorizer";
import hasPermission from "../../utils/permission";
import { useStore } from "../../store";
import formatCurrency from "@/utils/currency";
import { INotificationsError, INotificationsSuccess } from "@/interfaces/INotifications";

export default defineComponent({
  props: {
    nextPaymentDue: {
      type: Number,
      required: true,
    },
    currency: {
      type: String,
      required: true,
    },
  },
  setup() {
    const store = useStore();
    const router = useRouter();
    const dialog = ref(false);

    const hasAuthorization = computed(() => {
      const role = store.getters["auth/role"];
      if (role !== "") {
        return hasPermission(
          authorizer.role[role],
          actions.billing.unsubscribe,
        );
      }

      return false;
    });

    const cancelSubscription = async () => {
      try {
        await store.dispatch("billing/cancelSubscription");
        dialog.value = false;
        store.dispatch("snackbar/showSnackbarSuccessAction", INotificationsSuccess.cancelSubscription);

        store.dispatch("devices/setDeviceChooserStatus", store.getters["stats/stats"].registered_devices > 3);
        router.push({ name: "profileSettings" });
      } catch (error: any) {
        store.dispatch("snackbar/showSnackbarErrorAction", INotificationsError.cancelSubscription);
        throw new Error(error);
      }
    };
    return {
      dialog,
      hasAuthorization,
      formatCurrency,
      cancelSubscription,
    };
  },
});
</script>
