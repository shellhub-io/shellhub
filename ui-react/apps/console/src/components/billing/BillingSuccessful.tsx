import { CheckCircleIcon } from "@heroicons/react/24/solid";

export default function BillingSuccessful() {
  return (
    <div className="flex flex-col items-center text-center py-6 animate-fade-in">
      <span className="w-20 h-20 rounded-full bg-accent-green/10 border border-accent-green/20 flex items-center justify-center text-accent-green mb-5 shadow-[0_0_24px_rgba(130,165,104,0.25)]">
        <CheckCircleIcon className="w-12 h-12" />
      </span>
      <h3 className="text-lg font-semibold text-text-primary mb-2">
        Subscription activated
      </h3>
      <p className="text-sm text-text-secondary max-w-md leading-relaxed">
        Thank you for subscribing. Your premium features are now unlocked for
        this namespace. You can manage billing, invoices, and payment methods at
        any time from the Stripe billing portal.
      </p>
    </div>
  );
}
