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
import apiClient from "../api/client";
import { toCustomer, toSubscription } from "../types/billing";
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
    customer: toCustomer(result.data),
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
    subscription: toSubscription(result.data),
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

/**
 * `POST /api/billing/portal` is registered dynamically in the cloud module
 * and not present in the OpenAPI spec, so it is not exposed in the generated
 * SDK. Call it via the shared axios client, matching the Vue UI's behavior.
 */
export function useOpenBillingPortal() {
  return useMutation({
    mutationFn: async () => {
      const res = await apiClient.post<{ url: string }>("/api/billing/portal");
      const url = res.data?.url;
      if (!url) throw new Error("Missing billing portal URL");
      window.open(url, "_blank", "noopener,noreferrer");
      return url;
    },
  });
}
