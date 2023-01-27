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
        Are you sure?
      </v-card-title>

      <v-card-text class="mt-4 mb-3 pb-1">
        <p
          v-if="hasAuthorization && billingActive && billingInfo != undefined"
          data-test="contentSubscription-p"
        >
          Deleting the namespace will generate an invoice, estimated
          <b> {{ formatCurrency(billingInfo.nextPaymentDue, billingInfo.currency) }} </b> for the time
          of use.
        </p>

        <p data-test="content-text">
          This action cannot be undone. This will permanently delete the
          <b> {{ displayOnlyTenCharacters(name) }} </b>and its related data.
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

export default defineComponent({
  props: {
    nsTenant: {
      type: String,
      required: true,
    },
  },
  setup(props) {
    const store = useStore();
    const router = useRouter();
    const dialog = ref(false);
    const name = ref("");

    const tenant = computed(() => props.nsTenant);
    const billingActive = computed(() => store.getters["billing/active"]);
    const billing = computed(() => store.getters["billing/get"]);
    const billingInfo = computed(
      () => store.getters["billing/getBillInfoData"].info,
    );

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
        } catch (error: any) {
          store.dispatch("snackbar/showSnackbarErrorDefault");
          throw new Error(error);
        }
      }
    };

    onMounted(() => {
      if (hasAuthorization.value && isBillingEnabled()) {
        getSubscriptionInfo();
      }

      name.value = store.getters["namespaces/get"].name;
    });

    const getDueAmount = (data: any) => data.upcoming_invoice.amount_due;

    const remove = async () => {
      try {
        dialog.value = !dialog.value;
        await store.dispatch("namespaces/remove", tenant.value);
        await store.dispatch("auth/logout");
        await store.dispatch("layout/setLayout", "simpleLayout");
        await router.push({ name: "login" });
        store.dispatch(
          "snackbar/showSnackbarSuccessAction",
          INotificationsSuccess.namespaceDelete,
        );
      } catch (error: any) {
        store.dispatch(
          "snackbar/showSnackbarErrorAction",
          INotificationsError.namespaceDelete,
        );
        throw new Error(error);
      }
    };

    return {
      dialog,
      hasAuthorization,
      name,
      tenant,
      billing,
      billingInfo,
      billingActive,
      isBillingEnabled,
      getSubscriptionInfo,
      getDueAmount,
      displayOnlyTenCharacters,
      formatCurrency,
      remove,
    };
  },
});
</script>
