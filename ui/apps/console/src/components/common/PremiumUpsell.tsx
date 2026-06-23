import { ReactNode } from "react";
import { ArrowTopRightOnSquareIcon } from "@heroicons/react/24/outline";
import ProBadge from "@/components/common/ProBadge";

interface Props {
  /** Sized but uncolored heroicon, e.g. `<UsersIcon className="w-5 h-5" />`. */
  icon: ReactNode;
  title: string;
  description: string;
  href?: string;
  cta?: string;
}

// Inline upsell for a paid feature, placed next to a gated control where the
// full-page FeatureGate would be too heavy.
export default function PremiumUpsell({
  icon,
  title,
  description,
  href = "https://www.shellhub.io/pricing",
  cta = "See plans",
}: Props) {
  return (
    <div className="rounded-xl border border-accent-yellow/25 bg-gradient-to-br from-accent-yellow/[0.08] to-accent-yellow/[0.02] p-4">
      <div className="flex items-start gap-3">
        <span className="flex h-9 w-9 shrink-0 items-center justify-center rounded-lg bg-accent-yellow/15 text-accent-yellow">
          {icon}
        </span>
        <div className="min-w-0 flex-1">
          <div className="flex items-center gap-2">
            <span className="text-sm font-semibold text-text-primary">
              {title}
            </span>
            <ProBadge />
          </div>
          <p className="mt-1 text-xs leading-relaxed text-text-secondary">
            {description}
          </p>
        </div>
      </div>
      <div className="mt-3 flex items-center justify-between gap-3 pl-12">
        <span className="text-2xs text-text-muted">
          Available on Cloud &amp; Enterprise
        </span>
        <a
          href={href}
          target="_blank"
          rel="noopener noreferrer"
          className="inline-flex shrink-0 items-center gap-1.5 rounded-lg bg-accent-yellow/15 px-3 py-1.5 text-2xs font-semibold text-accent-yellow transition-all hover:bg-accent-yellow/25"
        >
          {cta}
          <ArrowTopRightOnSquareIcon className="h-3.5 w-3.5" strokeWidth={2} />
        </a>
      </div>
    </div>
  );
}
