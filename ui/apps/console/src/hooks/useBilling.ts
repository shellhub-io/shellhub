import { useMutation } from "@tanstack/react-query";
import {
  createCustomerMutation,
  createSubscriptionMutation,
  attachPaymentMethodMutation,
  detachPaymentMethodMutation,
  setDefaultPaymentMethodMutation,
  createBillingPortalSession,
} from "../client";
import { useInvalidateByIds } from "./useInvalidateQueries";

function useInvalidateBilling() {
  return useInvalidateByIds("getCustomer", "getSubscription", "getNamespace");
}

/**
 * Creates the billing customer, refreshing every billing query on success.
 */
export function useCreateCustomer() {
  const invalidate = useInvalidateBilling();
  return useMutation({
    ...createCustomerMutation(),
    onSuccess: invalidate,
  });
}

/**
 * Starts a subscription, refreshing every billing query on success.
 */
export function useCreateSubscription() {
  const invalidate = useInvalidateBilling();
  return useMutation({
    ...createSubscriptionMutation(),
    onSuccess: invalidate,
  });
}

/**
 * Attaches a payment method, refreshing every billing query on success.
 */
export function useAttachPaymentMethod() {
  const invalidate = useInvalidateBilling();
  return useMutation({
    ...attachPaymentMethodMutation(),
    onSuccess: invalidate,
  });
}

/**
 * Detaches a payment method, refreshing every billing query on success.
 */
export function useDetachPaymentMethod() {
  const invalidate = useInvalidateBilling();
  return useMutation({
    ...detachPaymentMethodMutation(),
    onSuccess: invalidate,
  });
}

/**
 * Makes a payment method the default one, refreshing every billing query on success.
 */
export function useSetDefaultPaymentMethod() {
  const invalidate = useInvalidateBilling();
  return useMutation({
    ...setDefaultPaymentMethodMutation(),
    onSuccess: invalidate,
  });
}

/**
 * Opens the provider's billing portal. The URL is single-use and short-lived, so it is minted on
 * demand rather than fetched with the page and held.
 */
export function useOpenBillingPortal() {
  return useMutation({
    mutationFn: async () => {
      const { data } = await createBillingPortalSession({ throwOnError: true });
      if (!data.url) throw new Error("Missing billing portal URL");
      window.open(data.url, "_blank", "noopener,noreferrer");
      return data.url;
    },
  });
}
