import { ReactNode, lazy, Suspense, useEffect, useRef, useState } from "react";
import { useLocation } from "react-router-dom";
import {
  ArrowTopRightOnSquareIcon,
  CalendarIcon,
  CheckCircleIcon,
  CreditCardIcon,
  ExclamationTriangleIcon,
  InformationCircleIcon,
} from "@heroicons/react/24/outline";
import { useAuthStore } from "@/stores/authStore";
import { useHasPermission } from "@/hooks/useHasPermission";
import { useNamespace } from "@/hooks/useNamespaces";
import { useOpenBillingPortal, useSubscription } from "@/hooks/useBilling";
import { useInvalidateByIds } from "@/hooks/useInvalidateQueries";
import { formatExpiry } from "@/utils/date";
import { readNamespaceBilling, type BillingStatus } from "@/types/billing";

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

function SectionRow({
  icon,
  title,
  description,
  children,
}: {
  icon: ReactNode;
  title: string;
  description: ReactNode;
  children?: ReactNode;
}) {
  return (
    <div className="flex items-center justify-between gap-6 px-5 py-4">
      <div className="flex items-start gap-3 min-w-0 flex-1">
        <span className="w-8 h-8 rounded-lg bg-hover-medium border border-border flex items-center justify-center text-text-muted shrink-0 mt-0.5">
          {icon}
        </span>
        <div className="min-w-0">
          <p className="text-sm font-medium text-text-primary">{title}</p>
          <div className="text-2xs text-text-muted mt-0.5 leading-relaxed">
            {description}
          </div>
        </div>
      </div>
      {children !== undefined && <div className="shrink-0">{children}</div>}
    </div>
  );
}

function StatusBadge({ status }: { status: BillingStatus }) {
  return (
    <span
      className={`inline-flex items-center px-2.5 py-1 text-2xs font-mono font-semibold rounded border ${STATUS_CHIP[status]}`}
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

interface BillingSectionProps {
  sectionId: string;
}

export default function BillingSection({ sectionId }: BillingSectionProps) {
  const location = useLocation();
  const sectionRef = useRef<HTMLDivElement>(null);
  const canSubscribe = useHasPermission("billing:subscribe");
  const { tenant: tenantId } = useAuthStore();
  const { namespace } = useNamespace(tenantId ?? "");
  const billing = readNamespaceBilling(namespace?.billing);
  // GET /subscription returns 400 until the namespace has BOTH a Stripe
  // customer AND an actual subscription (backend: ErrNamespaceSubscriptionUndefined).
  // Gating only on customer_id causes a retry cascade after customer bootstrap
  // — and every billing mutation's invalidate() then awaits those 400 retries,
  // making Save card / Set default hang for ~7s on each click.
  const hasSubscription = !!billing?.customer_id && !!billing?.subscription_id;
  const { subscription, isLoading } = useSubscription(hasSubscription);
  const openPortal = useOpenBillingPortal();
  const invalidate = useInvalidateByIds(
    "getCustomer",
    "getSubscription",
    "getNamespace",
  );
  const [wizardOpen, setWizardOpen] = useState(false);

  useEffect(() => {
    if (location.hash !== `#${sectionId}`) return;
    sectionRef.current?.scrollIntoView({ block: "start", behavior: "smooth" });
  }, [location.hash, sectionId]);

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
  // Only statuses where no Stripe subscription exists — safe to create a new one.
  const canShowSubscribeButton =
    status === "inactive" ||
    status === "canceled" ||
    status === "incomplete_expired";
  const canReopenPortal =
    isActiveLike ||
    status === "unpaid" ||
    status === "paused" ||
    status === "incomplete";

  const dateDescription = (() => {
    if (!endAt) return "—";
    if (status === "to_cancel_at_end_of_period" || status === "canceled") {
      return `Ends on ${formatExpiry(endAt)}`;
    }
    if (isActiveLike) return `Renews on ${formatExpiry(endAt)}`;
    return formatExpiry(endAt);
  })();

  return (
    <div
      id={sectionId}
      ref={sectionRef}
      className="bg-card border border-border rounded-xl overflow-hidden scroll-mt-8"
    >
      <div className="px-5 py-3.5 border-b border-border flex items-center justify-between">
        <h3 className="text-sm font-semibold text-text-primary">Billing</h3>
        {!isLoading && <StatusBadge status={status} />}
      </div>

      <div className="divide-y divide-border">
        {!canSubscribe && (
          <SectionRow
            icon={<InformationCircleIcon className="w-4 h-4" />}
            title="Owner-only"
            description="Only the namespace owner can view and manage billing details."
          />
        )}

        {canSubscribe && banner && (
          <div
            role="status"
            aria-live="polite"
            className={`flex items-start gap-3 px-5 py-3 border-b ${BANNER_CLASSES[banner.tone]}`}
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

        {canSubscribe && (
          <>
            <SectionRow
              icon={<CreditCardIcon className="w-4 h-4" />}
              title="Plan"
              description={
                status === "inactive"
                  ? "Free plan — up to 3 devices."
                  : "ShellHub Cloud Premium — unlimited devices."
              }
            >
              {isActiveLike ? (
                <span className="inline-flex items-center px-2.5 py-1 text-2xs font-mono font-semibold rounded border bg-primary/10 text-primary border-primary/20">
                  Premium
                </span>
              ) : (
                <span className="inline-flex items-center px-2.5 py-1 text-2xs font-mono font-semibold rounded border bg-hover-medium text-text-muted border-border">
                  Free
                </span>
              )}
            </SectionRow>

            {endAt > 0 && (
              <SectionRow
                icon={<CalendarIcon className="w-4 h-4" />}
                title="Billing cycle"
                description={dateDescription}
              />
            )}

            {invoice && isActiveLike && (
              <SectionRow
                icon={<CheckCircleIcon className="w-4 h-4" />}
                title="Upcoming charge"
                description={`Next invoice ${invoice.status === "paid" ? "paid" : "due"}.`}
              >
                <span className="text-sm font-mono text-text-primary tabular-nums">
                  {formatCurrency(invoice.amount, invoice.currency)}
                </span>
              </SectionRow>
            )}

            <SectionRow
              icon={<ArrowTopRightOnSquareIcon className="w-4 h-4" />}
              title="Billing portal"
              description="Manage invoices, payment methods, and download receipts in the Stripe portal."
            >
              <div className="flex items-center gap-2">
                {canShowSubscribeButton && (
                  <button
                    type="button"
                    onClick={() => setWizardOpen(true)}
                    onMouseEnter={preloadStripe}
                    onFocus={preloadStripe}
                    className="px-4 py-2 bg-primary hover:bg-primary-600 text-white rounded-lg text-sm font-semibold transition-all"
                  >
                    Subscribe
                  </button>
                )}
                {canReopenPortal && (
                  <div className="flex flex-col items-end gap-1.5">
                    <button
                      type="button"
                      onClick={() => openPortal.mutate()}
                      disabled={openPortal.isPending}
                      className="inline-flex items-center gap-1.5 px-4 py-2 bg-card hover:bg-hover-medium border border-border hover:border-border-light text-text-primary rounded-lg text-sm font-medium transition-all disabled:opacity-dim disabled:cursor-not-allowed"
                    >
                      {openPortal.isPending ? (
                        <span
                          aria-hidden="true"
                          className="w-4 h-4 border-2 border-text-muted/30 border-t-text-primary rounded-full animate-spin"
                        />
                      ) : (
                        <ArrowTopRightOnSquareIcon
                          className="w-4 h-4"
                          strokeWidth={2}
                        />
                      )}
                      Open portal
                    </button>
                    {openPortal.isError && (
                      <p role="alert" className="text-2xs text-accent-red">
                        Couldn't open the billing portal. Please try again.
                      </p>
                    )}
                  </div>
                )}
              </div>
            </SectionRow>
          </>
        )}
      </div>

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
    </div>
  );
}
