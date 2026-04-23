import { useCustomer } from "@/hooks/useBilling";
import BillingIcon from "./BillingIcon";

export default function BillingCheckout() {
  const { customer } = useCustomer();
  const defaultPm = customer?.payment_methods.find((pm) => pm.default);

  return (
    <div className="space-y-5">
      <div>
        <h3 className="text-base font-semibold text-text-primary mb-1">
          Review and confirm
        </h3>
        <p className="text-sm text-text-muted">
          Your subscription will renew automatically each month. You can cancel
          or update your payment method any time from the billing portal.
        </p>
      </div>

      {defaultPm ? (
        <div className="bg-card border border-border rounded-xl p-4 flex items-center gap-4">
          <span className="shrink-0 text-text-primary">
            <BillingIcon brand={defaultPm.brand} className="w-8 h-8" />
          </span>
          <div className="min-w-0 flex-1">
            <div className="flex items-center gap-2">
              <p className="text-sm font-mono text-text-primary truncate">
                {defaultPm.brand.charAt(0).toUpperCase() +
                  defaultPm.brand.slice(1)}
                {" • "}
                {defaultPm.number.slice(-4)}
              </p>
              <span className="inline-flex items-center px-2 py-0.5 text-3xs font-mono font-semibold uppercase tracking-label rounded border bg-primary/10 text-primary border-primary/20">
                Default
              </span>
            </div>
            <p className="text-2xs text-text-muted mt-0.5 font-mono">
              Expires {String(defaultPm.exp_month).padStart(2, "0")}/
              {defaultPm.exp_year}
            </p>
          </div>
        </div>
      ) : (
        <div className="bg-card border border-border rounded-xl p-4 text-sm text-text-muted">
          No default payment method. Go back and choose one before confirming.
        </div>
      )}

      <ul className="text-xs text-text-muted space-y-1.5 leading-relaxed list-disc list-inside">
        <li>Charged monthly based on the number of accepted devices.</li>
        <li>Cancel anytime from the Stripe billing portal.</li>
        <li>Invoices and receipts are available in your billing portal.</li>
      </ul>
    </div>
  );
}
