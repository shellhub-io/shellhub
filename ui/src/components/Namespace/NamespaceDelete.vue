<template>
  <v-tooltip location="bottom" class="text-center" :disabled="hasAuthorization">
    <template v-slot:activator="{ props }">
      <div v-bind="props">
        <v-btn
          :disabled="!hasAuthorization"
          color="red darken-1"
          variant="outlined"
          data-test="delete-btn"
          @click="dialog = !dialog"
        >
          Delete namespace
        </v-btn>
      </div>
    </template>
    <span> You don't have this kind of authorization. </span>
  </v-tooltip>

  <v-dialog v-model="dialog" max-width="540">
    <v-card data-test="namespaceDelete-dialog" class="bg-v-theme-surface">
      <v-card-title class="text-h5 pa-4 bg-primary">
        Namespace Deletion Restriction
      </v-card-title>

      <v-card-text class="mt-4 mb-3 pb-1">
        <div
          v-if="hasAuthorization && billingActive"
          data-test="contentSubscription-p"
        >
          <p class="mb-2">
            To ensure the integrity of your namespace,
            we have implemented a restriction that prevents its deletion while you have an active subscription or an unpaid invoice.
          </p>
          <p class="mb-2">
            Kindly note that in order to proceed with the deletion of your namespace,
            please ensure that there are no active subscriptions associated with it, and all outstanding invoices are settled.
          </p>
        </div>

        <p data-test="content-text" v-else>
          This action cannot be undone. This will permanently delete the
          <b> {{ displayOnlyTenCharacters(name) }} </b> and its related data.
        </p>
      </v-card-text>

      <v-card-actions>
        <v-spacer />

        <v-btn variant="text" data-test="close-btn" @click="dialog = !dialog">
          Close
        </v-btn>

        <v-btn
          color="red darken-1"
          variant="text"
          data-test="remove-btn"
          @click="remove()"
          :disabled="billingActive"
        >
          Remove
        </v-btn>
      </v-card-actions>
    </v-card>
  </v-dialog>
</template>

<script lang="ts">
import { defineComponent, ref, computed, onMounted } from "vue";
import { useRouter } from "vue-router";
import axios, { AxiosError } from "axios";
import { useStore } from "../../store";
import hasPermission from "../../utils/permission";
import { actions, authorizer } from "../../authorizer";
import { envVariables } from "../../envVariables";
import { displayOnlyTenCharacters } from "../../utils/string";
import formatCurrency from "@/utils/currency";
import {
  INotificationsError,
  INotificationsSuccess,
} from "../../interfaces/INotifications";
import handleError from "@/utils/handleError";

export default defineComponent({
  props: {
    nsTenant: {
      type: String,
      required: true,
    },
  },
  emits: ["billing-in-debt"],
  setup(props, ctx) {
    const store = useStore();
    const router = useRouter();
    const dialog = ref(false);
    const name = ref("");
    const tenant = computed(() => props.nsTenant);
    const billingActive = computed(() => store.getters["billing/active"]);
    const billing = computed(() => store.getters["billing/get"]);

    const hasAuthorization = computed(() => {
      const role = store.getters["auth/role"];
      if (role !== "") {
        return hasPermission(
          authorizer.role[role],
          actions.namespace.remove,
        );
      }
      return false;
    });

    const isBillingEnabled = () => envVariables.billingEnable;

    const getSubscriptionInfo = async () => {
      if (billingActive.value) {
        try {
          await store.dispatch("billing/getSubscription");
        } catch (error: unknown) {
          store.dispatch("snackbar/showSnackbarErrorDefault");
          handleError(error);
        }
      }
    };

    onMounted(() => {
      if (hasAuthorization.value && isBillingEnabled()) {
        getSubscriptionInfo();
      }

      name.value = store.getters["namespaces/get"].name;
    });

    const remove = async () => {
      try {
        dialog.value = !dialog.value;
        await store.dispatch("namespaces/remove", tenant.value);
        await store.dispatch("auth/logout");
        await router.push({ name: "login" });
        store.dispatch(
          "snackbar/showSnackbarSuccessAction",
          INotificationsSuccess.namespaceDelete,
        );
      } catch (error: unknown) {
        if (axios.isAxiosError(error)) {
          const axiosError = error as AxiosError;
          switch (axiosError.response?.status) {
            case 402:
              ctx.emit("billing-in-debt");
              break;
            default:
              break;
          }
        }
        store.dispatch(
          "snackbar/showSnackbarErrorAction",
          INotificationsError.namespaceDelete,
        );
        handleError(error);
      }
    };

    return {
      dialog,
      hasAuthorization,
      name,
      tenant,
      billing,
      billingActive,
      isBillingEnabled,
      getSubscriptionInfo,
      displayOnlyTenCharacters,
      formatCurrency,
      remove,
    };
  },
});
</script>
