import { SparklesIcon } from "@heroicons/react/24/outline";

export default function BillingLetter() {
  return (
    <div className="space-y-5">
      <div className="flex items-center gap-3">
        <span className="w-10 h-10 rounded-lg bg-primary/10 border border-primary/20 flex items-center justify-center text-primary shrink-0">
          <SparklesIcon className="w-5 h-5" />
        </span>
        <div>
          <p className="text-2xs font-mono font-semibold uppercase tracking-label text-primary">
            ShellHub Cloud
          </p>
          <h3 className="text-base font-semibold text-text-primary">
            Premium subscription
          </h3>
        </div>
      </div>

      <div className="space-y-3 text-sm text-text-secondary leading-relaxed">
        <p>
          Your subscription is billed monthly based on the number of devices
          registered in this namespace. You&apos;ll only pay for what you use —
          there are no upfront fees, no commitments, and you can cancel anytime.
        </p>
        <p>
          Subscribing unlocks unlimited devices, session playback, firewall
          rules, and enterprise-grade features tailored for teams operating SSH
          access at scale.
        </p>
        <p className="text-text-muted text-xs">
          Payment is processed securely by Stripe. Your card details are never
          stored on ShellHub servers.
        </p>
      </div>
    </div>
  );
}
