import { lazy, Suspense, useEffect, useRef, useState } from "react";
import {
  ArrowTopRightOnSquareIcon,
  ExclamationTriangleIcon,
  InformationCircleIcon,
} from "@heroicons/react/24/outline";
import { useAuthStore } from "@/stores/authStore";
import { useHasPermission } from "@/hooks/useHasPermission";
import { useNamespace } from "@/hooks/useNamespaces";
import { useOpenBillingPortal, useSubscription } from "@/hooks/useBilling";
import { useInvalidateByIds } from "@/hooks/useInvalidateQueries";
import { formatExpiry } from "@/utils/date";
import type { BillingStatus } from "@/client";
import { cn } from "@shellhub/design-system/cn";
import SettingsField from "@/components/settings/SettingsField";
import { Button } from "@shellhub/design-system/primitives";

const BillingDialog = lazy(() => import("./BillingDialog"));

const STATUS_CHIP: Record<BillingStatus, string> = {
  active: "bg-accent-green/10 text-accent-green border-accent-green/20",
  trialing: "bg-accent-cyan/10 text-accent-cyan border-accent-cyan/20",
  to_cancel_at_end_of_period:
    "bg-accent-yellow/10 text-accent-yellow border-accent-yellow/20",
  past_due: "bg-accent-red/10 text-accent-red border-accent-red/20",
  unpaid: "bg-accent-red/10 text-accent-red border-accent-red/20",
  canceled: "bg-accent-red/10 text-accent-red border-accent-red/20",
  incomplete: "bg-accent-yellow/10 text-accent-yellow border-accent-yellow/20",
  incomplete_expired: "bg-accent-red/10 text-accent-red border-accent-red/20",
  paused: "bg-accent-yellow/10 text-accent-yellow border-accent-yellow/20",
  inactive: "bg-hover-medium text-text-muted border-border",
};

const STATUS_LABEL: Record<BillingStatus, string> = {
  active: "Active",
  trialing: "Trialing",
  to_cancel_at_end_of_period: "Ending soon",
  past_due: "Past due",
  unpaid: "Unpaid",
  canceled: "Canceled",
  incomplete: "Incomplete",
  incomplete_expired: "Expired",
  paused: "Paused",
  inactive: "Inactive",
};

interface BannerConfig {
  tone: "error" | "warning" | "info";
  title: string;
  body: string;
  Icon: typeof ExclamationTriangleIcon;
}

function bannerFor(status: BillingStatus, endAt: number): BannerConfig | null {
  switch (status) {
    case "past_due":
      return {
        tone: "error",
        title: "Payment overdue",
        body: "Your last payment failed. Open the billing portal to update your payment method.",
        Icon: ExclamationTriangleIcon,
      };
    case "unpaid":
      return {
        tone: "error",
        title: "Unpaid invoices",
        body: "You have unpaid invoices. Settle them in the billing portal to restore premium features.",
        Icon: ExclamationTriangleIcon,
      };
    case "canceled":
      return {
        tone: "error",
        title: "Subscription canceled",
        body: "Resubscribe to regain unlimited devices and premium features.",
        Icon: ExclamationTriangleIcon,
      };
    case "to_cancel_at_end_of_period":
      return {
        tone: "warning",
        title: "Subscription ending",
        body: endAt
          ? `Your subscription ends on ${formatExpiry(endAt)}. You can reactivate it from the billing portal.`
          : "Your subscription is scheduled to end. You can reactivate it from the billing portal.",
        Icon: ExclamationTriangleIcon,
      };
    case "incomplete":
      return {
        tone: "warning",
        title: "Subscription incomplete",
        body: "Your initial payment wasn't completed. Open the billing portal to authorize the payment or update your card.",
        Icon: ExclamationTriangleIcon,
      };
    case "incomplete_expired":
      return {
        tone: "error",
        title: "Subscription expired",
        body: "Your initial payment window expired. Subscribe again to get started.",
        Icon: ExclamationTriangleIcon,
      };
    case "paused":
      return {
        tone: "warning",
        title: "Subscription paused",
        body: "Your subscription is paused. Resume it from the billing portal to continue using premium features.",
        Icon: ExclamationTriangleIcon,
      };
    case "inactive":
      return {
        tone: "info",
        title: "No active subscription",
        body: "Subscribe to ShellHub Cloud to register unlimited devices and unlock premium features.",
        Icon: InformationCircleIcon,
      };
    default:
      return null;
  }
}

const BANNER_CLASSES: Record<BannerConfig["tone"], string> = {
  error: "bg-accent-red/[0.06] border-accent-red/10 text-accent-red",
  warning: "bg-accent-yellow/[0.06] border-accent-yellow/10 text-accent-yellow",
  info: "bg-accent-blue/[0.06] border-accent-blue/10 text-accent-blue",
};

function StatusBadge({ status }: { status: BillingStatus }) {
  return (
    <span
      className={cn(
        "inline-flex items-center px-2.5 py-1 text-2xs font-mono font-semibold rounded border",
        STATUS_CHIP[status],
      )}
    >
      {STATUS_LABEL[status]}
    </span>
  );
}

function formatCurrency(amountCents: number, currency: string): string {
  try {
    return new Intl.NumberFormat("en-US", {
      style: "currency",
      currency: currency.toUpperCase(),
    }).format(amountCents / 100);
  } catch {
    return `${(amountCents / 100).toFixed(2)} ${currency.toUpperCase()}`;
  }
}

/**
 * The billing settings of a cloud namespace: its status, plan, cycle and upcoming charge, and the
 * way to subscribe or into the provider's portal. Only whoever may subscribe sees the details.
 */
export default function BillingSection() {
  const canSubscribe = useHasPermission("billing:subscribe");
  const { tenant: tenantId } = useAuthStore();
  const { namespace } = useNamespace(tenantId ?? "");
  const billing = namespace?.billing;
  const hasSubscription = !!billing?.customer_id && !!billing?.subscription?.id;
  const { subscription, isLoading } = useSubscription(hasSubscription);
  const openPortal = useOpenBillingPortal();
  const invalidate = useInvalidateByIds(
    "getCustomer",
    "getSubscription",
    "getNamespace",
  );
  const [wizardOpen, setWizardOpen] = useState(false);

  const invalidateRef = useRef(invalidate);

  useEffect(() => {
    const onPageShow = (e: PageTransitionEvent) => {
      if (e.persisted) void invalidateRef.current();
    };
    window.addEventListener("pageshow", onPageShow);
    return () => window.removeEventListener("pageshow", onPageShow);
  }, []);

  const preloadStripe = () => {
    void import("./BillingDialog");
  };

  const status: BillingStatus = subscription?.status ?? "inactive";
  const endAt = subscription?.end_at ?? 0;
  const invoice = subscription?.invoices?.[0];
  const banner = bannerFor(status, endAt);
  const isActiveLike =
    status === "active" ||
    status === "trialing" ||
    status === "past_due" ||
    status === "to_cancel_at_end_of_period";
  const canShowSubscribeButton =
    status === "inactive" ||
    status === "canceled" ||
    status === "incomplete_expired";
  const canReopenPortal =
    isActiveLike ||
    status === "unpaid" ||
    status === "paused" ||
    status === "incomplete" ||
    status === "canceled" ||
    status === "incomplete_expired";

  const dateDescription = (() => {
    if (!endAt) return "—";
    if (status === "to_cancel_at_end_of_period" || status === "canceled") {
      return `Ends on ${formatExpiry(endAt)}`;
    }
    if (isActiveLike) return `Renews on ${formatExpiry(endAt)}`;
    return formatExpiry(endAt);
  })();

  return (
    <>
      {!canSubscribe ? (
        <SettingsField
          title="Owner only"
          description="Only the namespace owner can see and manage billing."
        />
      ) : (
        <>
          {banner && (
            <div
              role="status"
              aria-live="polite"
              className={cn(
                "flex items-start gap-3 px-4 py-3 rounded-xl border",
                BANNER_CLASSES[banner.tone],
              )}
            >
              <banner.Icon
                aria-hidden="true"
                className="w-4 h-4 mt-0.5 shrink-0"
              />
              <div>
                <p className="text-sm font-semibold">{banner.title}</p>
                <p className="text-2xs opacity-90 mt-0.5">{banner.body}</p>
              </div>
            </div>
          )}

          <SettingsField
            title="Plan"
            description={
              status === "inactive"
                ? "Free plan, up to 3 devices."
                : "ShellHub Cloud Premium, unlimited devices."
            }
          >
            <div className="flex items-center gap-2">
              {!isLoading && <StatusBadge status={status} />}
              {isActiveLike ? (
                <span className="inline-flex items-center px-2.5 py-1 text-2xs font-mono font-semibold rounded border bg-primary/10 text-primary border-primary/20">
                  Premium
                </span>
              ) : (
                <span className="inline-flex items-center px-2.5 py-1 text-2xs font-mono font-semibold rounded border bg-hover-medium text-text-muted border-border">
                  Free
                </span>
              )}
            </div>
          </SettingsField>

          {endAt > 0 && (
            <SettingsField title="Billing cycle" description={dateDescription} />
          )}

          {invoice && isActiveLike && (
            <SettingsField
              title="Upcoming charge"
              description={`Next invoice ${invoice.status === "paid" ? "paid" : "due"}.`}
            >
              <span className="text-sm font-mono text-text-primary tabular-nums">
                {formatCurrency(invoice.amount, invoice.currency)}
              </span>
            </SettingsField>
          )}

          <SettingsField
            title="Billing portal"
            description="Invoices, payment methods and receipts, in the Stripe portal."
          >
            <div className="flex items-center gap-2">
              {canShowSubscribeButton && (
                <Button
                  size="sm"
                  onClick={() => setWizardOpen(true)}
                  onMouseEnter={preloadStripe}
                  onFocus={preloadStripe}
                >
                  Subscribe
                </Button>
              )}
              {canReopenPortal && (
                <div className="flex flex-col items-end gap-1.5">
                  <Button
                    size="sm"
                    variant="secondary"
                    loading={openPortal.isPending}
                    icon={
                      <ArrowTopRightOnSquareIcon
                        className="w-4 h-4"
                        strokeWidth={2}
                      />
                    }
                    onClick={() => openPortal.mutate()}
                  >
                    Open portal
                  </Button>
                  {openPortal.isError && (
                    <p role="alert" className="text-2xs text-accent-red">
                      Couldn't open the billing portal. Try again.
                    </p>
                  )}
                </div>
              )}
            </div>
          </SettingsField>
        </>
      )}

      <Suspense fallback={null}>
        {wizardOpen && (
          <BillingDialog
            open={wizardOpen}
            onClose={() => setWizardOpen(false)}
            onSuccess={() => {
              void invalidate();
            }}
          />
        )}
      </Suspense>
    </>
  );
}
