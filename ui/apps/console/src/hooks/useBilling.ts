import { useMutation, useQuery } from "@tanstack/react-query";
import {
  getCustomerOptions,
  getSubscriptionOptions,
  createCustomerMutation,
  createSubscriptionMutation,
  attachPaymentMethodMutation,
  detachPaymentMethodMutation,
  setDefaultPaymentMethodMutation,
} from "../client/@tanstack/react-query.gen";
import { createBillingPortalSession } from "@/client";
import { useInvalidateByIds } from "./useInvalidateQueries";

function useInvalidateBilling() {
  return useInvalidateByIds("getCustomer", "getSubscription", "getNamespace");
}

export function useCustomer(enabled = true) {
  const result = useQuery({
    ...getCustomerOptions(),
    enabled,
  });
  return {
    customer: result.data,
    isLoading: result.isLoading,
    error: result.error,
    refetch: result.refetch,
  };
}

export function useSubscription(enabled = true) {
  const result = useQuery({
    ...getSubscriptionOptions(),
    enabled,
  });
  return {
    subscription: result.data,
    isLoading: result.isLoading,
    error: result.error,
    refetch: result.refetch,
  };
}

export function useCreateCustomer() {
  const invalidate = useInvalidateBilling();
  return useMutation({
    ...createCustomerMutation(),
    onSuccess: invalidate,
  });
}

export function useCreateSubscription() {
  const invalidate = useInvalidateBilling();
  return useMutation({
    ...createSubscriptionMutation(),
    onSuccess: invalidate,
  });
}

export function useAttachPaymentMethod() {
  const invalidate = useInvalidateBilling();
  return useMutation({
    ...attachPaymentMethodMutation(),
    onSuccess: invalidate,
  });
}

export function useDetachPaymentMethod() {
  const invalidate = useInvalidateBilling();
  return useMutation({
    ...detachPaymentMethodMutation(),
    onSuccess: invalidate,
  });
}

export function useSetDefaultPaymentMethod() {
  const invalidate = useInvalidateBilling();
  return useMutation({
    ...setDefaultPaymentMethodMutation(),
    onSuccess: invalidate,
  });
}

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
