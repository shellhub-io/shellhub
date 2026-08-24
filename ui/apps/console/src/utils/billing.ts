import type { BillingStatus, NamespaceBilling } from "@/client";

const ACTIVE_STATUSES = new Set<BillingStatus>([
  "active",
  "trialing",
  "past_due",
  "to_cancel_at_end_of_period",
]);

/**
 * Whether the namespace has a subscription that still grants service. A namespace with a
 * customer but no subscription never completed checkout, and stays on the free tier.
 */
export function hasActiveSubscription(
  billing: NamespaceBilling | undefined | null,
): boolean {
  const status = billing?.subscription?.status;
  return status !== undefined && ACTIVE_STATUSES.has(status);
}
