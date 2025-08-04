import { defineStore } from "pinia";
import { computed, ref } from "vue";
import * as billingApi from "../api/billing";
import { IBilling } from "@/interfaces/IBilling";
import { envVariables } from "@/envVariables";

const useBillingStore = defineStore("billing", () => {
  const billing = ref<IBilling>({} as IBilling);
  const isActive = computed(() => billing.value.active ?? false);
  const status = computed(() => billing.value.status ?? "inactive");
  const invoices = computed(() => billing.value.invoices ?? []);

  const getSubscriptionInfo = async (): Promise<void> => {
    try {
      if (envVariables.isCloud) {
        const res = await billingApi.getSubscriptionInfo();
        billing.value = res.data as IBilling;
      }
    } catch (error) {
      billing.value.active = false;
    }
  };

  const openBillingPortal = async (): Promise<void> => {
    const res = await billingApi.getBillingPortalUrl();
    window.open(res.data.url, "_blank");
  };

  return {
    billing,
    isActive,
    status,
    invoices,
    getSubscriptionInfo,
    openBillingPortal,
  };
});

export default useBillingStore;
