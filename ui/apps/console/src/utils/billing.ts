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

const BLOCKED_STATUSES = new Set<BillingStatus>([
  "past_due",
  "unpaid",
  "incomplete",
  "incomplete_expired",
  "paused",
]);

/**
 * Whether the subscription exists but needs payment or repair. Mirrors the backend's
 * blocked-subscription states: `evaluate.go` (inactive statuses) and `report.go`
 * (`past_due`, which is active but still returns a subscription 402 on accept).
 */
export function isSubscriptionBlocked(
  billing: NamespaceBilling | undefined | null,
): boolean {
  const status = billing?.subscription?.status;
  return status !== undefined && BLOCKED_STATUSES.has(status);
}
