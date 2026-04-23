import { useEffect, useRef, useState } from "react";
import { isSdkError } from "@/api/errors";
import {
  Elements,
  CardElement,
  useElements,
  useStripe,
} from "@stripe/react-stripe-js";
import {
  loadStripe,
  type Stripe,
  type StripeCardElementOptions,
  type StripeElementsOptions,
} from "@stripe/stripe-js";
import {
  PlusIcon,
  StarIcon as StarOutlineIcon,
  TrashIcon,
} from "@heroicons/react/24/outline";
import { StarIcon as StarSolidIcon } from "@heroicons/react/24/solid";
import { getConfig } from "@/env";
import { useNamespace } from "@/hooks/useNamespaces";
import { useAuthStore } from "@/stores/authStore";
import {
  useAttachPaymentMethod,
  useCreateCustomer,
  useCustomer,
  useDetachPaymentMethod,
  useSetDefaultPaymentMethod,
} from "@/hooks/useBilling";
import { readNamespaceBilling } from "@/types/billing";
import { stripeErrorMessage } from "@/utils/stripeErrors";
import { LABEL } from "@/utils/styles";
import BillingIcon from "./BillingIcon";

const ELEMENTS_OPTIONS: StripeElementsOptions = {
  appearance: { theme: "night" },
  locale: "en",
  loader: "auto",
  mode: "setup",
  currency: "usd",
  paymentMethodCreation: "manual",
  fonts: [
    {
      cssSrc:
        "https://fonts.googleapis.com/css2?family=IBM+Plex+Sans:wght@400;500;600&display=swap",
    },
  ],
};

const CARD_OPTIONS: StripeCardElementOptions = {
  hidePostalCode: true,
  style: {
    base: {
      color: "#E1E4EA",
      fontFamily: "'IBM Plex Sans', system-ui, sans-serif",
      fontSmoothing: "antialiased",
      fontSize: "14px",
      "::placeholder": { color: "#81879C" },
      iconColor: "#667ACC",
    },
    invalid: {
      color: "#D8737B",
      iconColor: "#D8737B",
    },
  },
};

let cachedStripePromise: Promise<Stripe | null> | null = null;
function getStripePromise(): Promise<Stripe | null> {
  if (!cachedStripePromise) {
    const key = getConfig().stripePublishableKey;
    cachedStripePromise = key ? loadStripe(key) : Promise.resolve(null);
  }
  return cachedStripePromise;
}

interface BillingPaymentProps {
  onHasDefault: () => void;
  onNoPaymentMethods: () => void;
}

export default function BillingPayment(props: BillingPaymentProps) {
  return (
    <Elements stripe={getStripePromise()} options={ELEMENTS_OPTIONS}>
      <BillingPaymentInner {...props} />
    </Elements>
  );
}

function BillingPaymentInner({
  onHasDefault,
  onNoPaymentMethods,
}: BillingPaymentProps) {
  const stripe = useStripe();
  const elements = useElements();
  const { tenant: tenantId } = useAuthStore();
  const { namespace, refetch: refetchNamespace } = useNamespace(tenantId ?? "");
  const billing = readNamespaceBilling(namespace?.billing);
  // /api/billing/customer returns 400 until the namespace has been bootstrapped
  // with Stripe (POST /customer). Gate the query on that.
  const hasCustomer = !!billing?.customer_id;
  const {
    customer,
    isLoading: customerLoading,
    refetch: refetchCustomer,
  } = useCustomer(hasCustomer);

  const createCustomer = useCreateCustomer();
  const attachPm = useAttachPaymentMethod();
  const detachPm = useDetachPaymentMethod();
  const setDefaultPm = useSetDefaultPaymentMethod();

  const [isAddingCard, setIsAddingCard] = useState(false);
  const [submitting, setSubmitting] = useState(false);
  const [cardComplete, setCardComplete] = useState(false);
  const [error, setError] = useState("");
  const [bootstrapping, setBootstrapping] = useState(true);
  const [retryCount, setRetryCount] = useState(0);

  const bootstrapRef = useRef(false);

  // Strictly sequential bootstrap, mirroring the Vue onMounted flow:
  //   1. refresh namespace to get the current billing.customer_id (if any),
  //   2. if missing, create the Stripe customer (backend writes customer_id),
  //   3. refresh namespace again so `hasCustomer` flips and useCustomer fires.
  // Calling /customer before the namespace has billing data yields HTTP 400.
  useEffect(() => {
    if (bootstrapRef.current) return;
    bootstrapRef.current = true;
    let cancelled = false;

    void (async () => {
      try {
        const fresh = await refetchNamespace();
        if (cancelled) return;
        const current = readNamespaceBilling(fresh.data?.billing);
        if (!current?.customer_id) {
          await createCustomer.mutateAsync({});
          if (cancelled) return;
          await refetchNamespace();
        }
      } catch {
        if (!cancelled)
          setError("Unable to initialize billing. Please try again.");
      } finally {
        if (!cancelled) setBootstrapping(false);
      }
    })();

    return () => {
      cancelled = true;
      bootstrapRef.current = false;
    };
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [retryCount]);

  const paymentMethods = customer?.payment_methods ?? [];
  const hasDefault = paymentMethods.some((pm) => pm.default);

  useEffect(() => {
    if (bootstrapping || customerLoading) return;
    if (hasDefault) onHasDefault();
    else onNoPaymentMethods();
  }, [
    bootstrapping,
    customerLoading,
    hasDefault,
    onHasDefault,
    onNoPaymentMethods,
  ]);

  const handleAddCard = async () => {
    if (!stripe || !elements) return;
    const card = elements.getElement(CardElement);
    if (!card) return;

    setSubmitting(true);
    setError("");
    try {
      const { paymentMethod, error: stripeErr } =
        await stripe.createPaymentMethod({
          type: "card",
          card,
        });
      if (stripeErr) {
        setError(stripeErrorMessage(stripeErr.code, stripeErr.message));
        return;
      }
      if (!paymentMethod) {
        setError("Unable to create payment method. Please try again.");
        return;
      }
      await attachPm.mutateAsync({ body: { id: paymentMethod.id } });
      await refetchCustomer();
      card.clear();
      setIsAddingCard(false);
      setCardComplete(false);
    } catch (err) {
      if (isSdkError(err)) {
        const body = err as unknown as { error?: string };
        const msg = typeof body.error === "string" ? body.error : undefined;
        setError(msg ?? "Failed to attach payment method.");
      } else if (err instanceof Error) {
        setError(err.message);
      } else {
        setError("Failed to attach payment method.");
      }
    } finally {
      setSubmitting(false);
    }
  };

  const handleSetDefault = async (id: string) => {
    setError("");
    try {
      await setDefaultPm.mutateAsync({ body: { id } });
      await refetchCustomer();
    } catch {
      setError("Failed to update default payment method.");
    }
  };

  const handleDetach = async (id: string) => {
    setError("");
    try {
      await detachPm.mutateAsync({ body: { id } });
      await refetchCustomer();
    } catch {
      setError("Failed to remove payment method.");
    }
  };

  if (!bootstrapping && error && !customer) {
    return (
      <div className="flex flex-col items-center justify-center py-10 gap-4">
        <p role="alert" className="text-sm text-accent-red text-center">
          {error}
        </p>
        <button
          type="button"
          onClick={() => {
            bootstrapRef.current = false;
            setError("");
            setBootstrapping(true);
            setRetryCount((c) => c + 1);
          }}
          disabled={createCustomer.isPending}
          className="px-4 py-2 rounded-lg text-xs font-semibold bg-primary text-white hover:bg-primary-600 transition-all disabled:opacity-40 disabled:cursor-not-allowed"
        >
          Retry
        </button>
      </div>
    );
  }

  if (bootstrapping || customerLoading) {
    return (
      <div
        className="flex items-center justify-center py-10"
        role="status"
        aria-live="polite"
      >
        <span className="w-5 h-5 border-2 border-primary/30 border-t-primary rounded-full animate-spin" />
        <span className="sr-only">Loading payment methods…</span>
      </div>
    );
  }

  return (
    <div className="space-y-5">
      <div className="grid grid-cols-2 gap-3">
        <div>
          <label className={LABEL} htmlFor="billing-customer-name">
            Name
          </label>
          <input
            id="billing-customer-name"
            type="text"
            value={customer?.name ?? ""}
            disabled
            readOnly
            className="w-full px-3.5 py-2.5 bg-background border border-border rounded-lg text-sm text-text-muted cursor-not-allowed"
          />
        </div>
        <div>
          <label className={LABEL} htmlFor="billing-customer-email">
            E-mail
          </label>
          <input
            id="billing-customer-email"
            type="email"
            value={customer?.email ?? ""}
            disabled
            readOnly
            className="w-full px-3.5 py-2.5 bg-background border border-border rounded-lg text-sm text-text-muted cursor-not-allowed"
          />
        </div>
      </div>

      <div>
        <h3 className="text-base font-semibold text-text-primary mb-1">
          Payment method
        </h3>
        <p className="text-sm text-text-muted">
          Add a card and mark it as default. This is the card that will be
          charged for your subscription.
        </p>
      </div>

      {paymentMethods.length > 0 && (
        <ul className="space-y-2" aria-label="Saved payment methods">
          {paymentMethods.map((pm) => (
            <li
              key={pm.id}
              className="bg-card border border-border rounded-lg px-4 py-3 flex items-center gap-3"
            >
              <span className="shrink-0 text-text-primary">
                <BillingIcon brand={pm.brand} className="w-7 h-7" />
              </span>
              <div className="min-w-0 flex-1">
                <div className="flex items-center gap-2">
                  <p className="text-sm font-mono text-text-primary truncate">
                    {pm.brand.charAt(0).toUpperCase() + pm.brand.slice(1)}
                    {" • "}
                    {pm.number.slice(-4)}
                  </p>
                  {pm.default && (
                    <span className="inline-flex items-center px-2 py-0.5 text-3xs font-mono font-semibold uppercase tracking-label rounded border bg-primary/10 text-primary border-primary/20">
                      Default
                    </span>
                  )}
                </div>
                <p className="text-2xs text-text-muted mt-0.5 font-mono">
                  Expires {String(pm.exp_month).padStart(2, "0")}/{pm.exp_year}
                </p>
              </div>
              <div className="flex items-center gap-1 shrink-0">
                {pm.default ? (
                  <span
                    className="p-1.5 text-primary"
                    title="Default payment method"
                    aria-label="Default payment method"
                  >
                    <StarSolidIcon className="w-4 h-4" />
                  </span>
                ) : (
                  <button
                    type="button"
                    onClick={() => void handleSetDefault(pm.id)}
                    disabled={setDefaultPm.isPending}
                    className="p-1.5 rounded-md text-text-muted hover:text-text-primary hover:bg-hover-medium transition-colors disabled:opacity-40"
                    aria-label={`Set ${pm.brand} ending ${pm.number.slice(-4)} as default`}
                    title="Set as default"
                  >
                    <StarOutlineIcon className="w-4 h-4" />
                  </button>
                )}
                <button
                  type="button"
                  onClick={() => void handleDetach(pm.id)}
                  disabled={detachPm.isPending || pm.default}
                  className="p-1.5 rounded-md text-text-muted hover:text-accent-red hover:bg-accent-red/10 transition-colors disabled:opacity-40 disabled:hover:text-text-muted disabled:hover:bg-transparent"
                  aria-label={`Remove ${pm.brand} ending ${pm.number.slice(-4)}`}
                  title={
                    pm.default ? "Cannot remove the default card" : "Remove"
                  }
                >
                  <TrashIcon className="w-4 h-4" />
                </button>
              </div>
            </li>
          ))}
        </ul>
      )}

      {!isAddingCard && (
        <button
          type="button"
          onClick={() => setIsAddingCard(true)}
          className="w-full flex items-center justify-center gap-2 px-4 py-3 bg-card border border-dashed border-border hover:border-primary/40 rounded-lg text-sm font-medium text-text-secondary hover:text-text-primary transition-colors"
        >
          <PlusIcon className="w-4 h-4" strokeWidth={2} />
          Add payment method
        </button>
      )}

      {isAddingCard && (
        <div className="bg-card border border-border rounded-lg p-4 space-y-3">
          <div>
            <label className={LABEL} htmlFor="stripe-card-element">
              Card details
            </label>
            <div
              id="stripe-card-element"
              className="w-full px-3.5 py-3 bg-background border border-border rounded-lg focus-within:border-primary/50 focus-within:ring-1 focus-within:ring-primary/20 transition-all"
              aria-label="Credit card details"
            >
              <CardElement
                options={CARD_OPTIONS}
                onChange={(e) => {
                  setCardComplete(e.complete);
                  if (e.error)
                    setError(stripeErrorMessage(e.error.code, e.error.message));
                  else setError("");
                }}
              />
            </div>
          </div>

          <div className="flex justify-end gap-2">
            <button
              type="button"
              onClick={() => {
                setIsAddingCard(false);
                setCardComplete(false);
                setError("");
              }}
              disabled={submitting}
              className="px-4 py-2 text-sm font-medium text-text-secondary hover:text-text-primary rounded-lg hover:bg-hover-subtle transition-colors disabled:opacity-40"
            >
              Cancel
            </button>
            <button
              type="button"
              onClick={() => void handleAddCard()}
              disabled={!stripe || !cardComplete || submitting}
              className="px-5 py-2 bg-primary hover:bg-primary-600 text-white rounded-lg text-sm font-semibold disabled:opacity-dim disabled:cursor-not-allowed transition-all flex items-center gap-2"
            >
              {submitting && (
                <span
                  aria-hidden="true"
                  className="w-4 h-4 border-2 border-white/30 border-t-white rounded-full animate-spin"
                />
              )}
              Save card
            </button>
          </div>
        </div>
      )}

      {error && (
        <p role="alert" className="text-2xs text-accent-red">
          {error}
        </p>
      )}
    </div>
  );
}
