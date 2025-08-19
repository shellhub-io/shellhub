<template>
  <SettingOwnerInfo v-if="!hasAuthorization" data-test="settings-owner-info-component" />
  <v-container fluid v-else>
    <BillingDialog v-model="dialogCheckout" @reload="reload" />
    <v-card
      variant="flat"
      class="bg-transparent"
      data-test="billing-card"
    >
      <v-card-item>
        <v-list-item
          class="pa-0"
          data-test="billing-header"
        >
          <template v-slot:title>
            <h1 data-test="billing-title">Billing</h1>
          </template>
          <template v-slot:subtitle>
            <span data-test="billing-subtitle">Manage your subscription info</span>
          </template>
          <template v-slot:append>
            <v-btn
              color="primary"
              variant="text"
              class="bg-secondary align-content-lg-center text-none text-uppercase"
              :disabled="billingStatus === ''"
              @click="dialogCheckout = true"
              data-test="subscribe-button"
            >
              Subscribe
            </v-btn>
          </template>
        </v-list-item>
      </v-card-item>
      <v-card-text class="pt-4">
        <v-list
          border
          rounded
          class="bg-background pa-0"
          data-test="billing-details-list"
        >
          <v-card-item
            style="grid-template-columns: max-content 1.5fr 2fr"
            v-if="hasAuthorization"
            data-test="billing-portal-section"
          >
            <template #prepend>
              <v-icon data-test="billing-portal-icon">mdi-account</v-icon>
            </template>
            <template #title>
              <span class="text-subtitle-1" data-test="billing-portal-title">Billing Portal</span>
            </template>
            <div data-test="billing-portal-description">
              Update your ShellHub payment method or download invoices.
            </div>
            <template #append>
              <v-btn
                :disabled="noCustomer.value"
                color="primary"
                class="mt-2 text-none text-uppercase"
                @click="openBillingPortal"
                data-test="billing-portal-button"
              >
                Open Billing Portal
              </v-btn>
            </template>
          </v-card-item>
          <v-divider data-test="billing-divider" />
          <v-card-item
            style="grid-template-columns: max-content 1.5fr 2fr"
            data-test="billing-plan-section"
          >
            <template #prepend>
              <v-icon data-test="billing-plan-icon">mdi-credit-card</v-icon>
            </template>
            <template #title>
              <span class="text-subtitle-1" data-test="billing-plan-title">Plan</span>
            </template>
            <div v-if="!isBillingActive" data-test="billing-plan-description-free">
              You can add up to 3 devices while using the 'Free' plan.
            </div>
            <div v-else data-test="billing-plan-description-premium">
              In this plan, the amount is charged according to the number of devices used.
            </div>
            <template #append>
              <h3 v-if="!isBillingActive" data-test="billing-plan-free">
                Free
              </h3>
              <h3 v-else data-test="billing-plan-premium">
                Premium usage
              </h3>
            </template>
          </v-card-item>
          <v-divider data-test="billing-divider" />
          <div v-if="hasAuthorization && isBillingActive" data-test="billing-active-section">
            <v-card-item
              style="grid-template-columns: max-content 1.5fr 2fr"
              v-if="message"
              data-test="billing-status-section"
            >
              <template #prepend>
                <v-icon data-test="billing-status-icon">mdi-invoice-text-remove</v-icon>
              </template>
              <template #title>
                <span class="text-subtitle-1" data-test="billing-status-title">Billing Status</span>
              </template>
              <template #append>
                <h3 :class="`text-${messageType}`" data-test="billing-status-message">{{ message }}</h3>
              </template>
            </v-card-item>
            <v-divider data-test="billing-divider" />
            <v-card-item
              style="grid-template-columns: max-content 1.5fr 2fr"
              data-test="billing-total-section"
            >
              <template #prepend>
                <v-icon data-test="billing-total-icon">mdi-invoice-text</v-icon>
              </template>
              <template #title>
                <span class="text-subtitle-1" data-test="billing-total-title">Billing estimated total</span>
              </template>
              <template #append>
                <h3 data-test="billing-total-amount">{{ formattedCurrency }}</h3>
              </template>
            </v-card-item>
            <v-divider data-test="billing-divider" />
            <v-card-item
              style="grid-template-columns: max-content 1.5fr 2fr"
              data-test="billing-end-date-section"
            >
              <template #prepend>
                <v-icon data-test="billing-end-date-icon">mdi-invoice-text-clock</v-icon>
              </template>
              <template #title>
                <span class="text-subtitle-1" data-test="billing-end-date-title">Current billing ends at</span>
              </template>
              <template #append>
                <h3 data-test="billing-end-date">{{ formattedDate }}</h3>
              </template>
            </v-card-item>
            <v-divider data-test="billing-divider" />
          </div>
        </v-list>
      </v-card-text>
    </v-card>
  </v-container>
</template>

<script setup lang="ts">
import {
  ref,
  computed,
  onMounted,
  reactive,
} from "vue";
import { storeToRefs } from "pinia";
import { useEventListener } from "@vueuse/core";
import hasPermission from "@/utils/permission";
import { actions, authorizer } from "@/authorizer";
import BillingDialog from "../Billing/BillingDialog.vue";
import SettingOwnerInfo from "./SettingOwnerInfo.vue";
import formatCurrency from "@/utils/currency";
import { formatUnixToDate } from "@/utils/date";
import handleError from "@/utils/handleError";
import useAuthStore from "@/store/modules/auth";
import useBillingStore from "@/store/modules/billing";
import useNamespacesStore from "@/store/modules/namespaces";

const authStore = useAuthStore();
const billingStore = useBillingStore();
const namespacesStore = useNamespacesStore();
const { billing: billingInfo, isActive: isBillingActive, status: billingStatus } = storeToRefs(billingStore);
const namespace = computed(() => namespacesStore.currentNamespace);
const el = ref<number>(1);
const dialogCheckout = ref(false);
const noCustomer = reactive({ value: false });
const message = ref("");
const messageType = ref();
const formattedDate = ref();
const formattedCurrency = ref();

const hasAuthorization = computed(() => {
  const { role } = authStore;
  return !!role && hasPermission(authorizer.role[role], actions.billing.subscribe);
});

useEventListener("pageshow", (event) => {
  const historyPage = event.persisted
  || (typeof window.performance !== "undefined"
  && (window.performance.getEntries()[0] as PerformanceNavigationTiming).type === "back_forward");
  if (historyPage) {
    window.location.reload();
  }
});

const handleErrors = async () => {
  switch (billingStatus.value) {
    case "to_cancel_at_end_of_period":
      message.value = `Your subscription will be canceled at ${formattedDate.value
      }, if you want to renew your subscription to premium, please, access our billing portal.`;
      messageType.value = "warning";
      break;
    case "past_due":
      message.value = "Your subscription payment method has failed. Please, access the billing portal to keep your subscription";
      messageType.value = "warning";
      break;
    case "unpaid":
      // eslint-disable-next-line vue/max-len
      message.value = "You have unpaid invoices which made your subscription to be canceled. Please, solve this issue opening the billing portal.";
      messageType.value = "error";
      break;
    case "canceled":
      message.value = "Your subscription was canceled. To continue to use ShellHub premium benefits, please, subscribe to a new one.";
      messageType.value = "error";
      break;
    default:
      break;
  }
};

const getSubscriptionInfo = async () => {
  if (hasAuthorization.value) {
    try {
      await billingStore.getSubscriptionInfo();
      const invoice = billingStore.invoices[0];
      formattedDate.value = formatUnixToDate(billingInfo.value.end_at);
      formattedCurrency.value = formatCurrency(invoice?.amount, invoice?.currency.toUpperCase());
    } catch (error: unknown) {
      handleError(error);
    }
  }
};

onMounted(async () => {
  const tenant = computed(() => localStorage.getItem("tenant") as string);
  await namespacesStore.fetchNamespace(tenant.value);
  if (!namespace.value.billing || !namespace.value.billing.customer_id) {
    noCustomer.value = true;
  }
  await getSubscriptionInfo();
  await handleErrors();
});

const openBillingPortal = async () => {
  try {
    await billingStore.openBillingPortal();
  } catch (error: unknown) {
    handleError(error);
  }
};

const reload = () => {
  window.location.reload();
};

defineExpose({
  dialogCheckout,
  el,
});
</script>

<style scoped>
p {
  text-align: justify;
  overflow: auto;
}

.hover-text {
  cursor: pointer;
}

.hover-text:hover {
  text-decoration: underline;
}
</style>
