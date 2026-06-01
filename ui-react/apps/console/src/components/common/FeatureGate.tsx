import { ReactNode } from "react";
import {
  LockClosedIcon,
  ArrowTopRightOnSquareIcon,
} from "@heroicons/react/24/outline";
import { getConfig } from "@/env";
import EmptyState, {
  type EmptyStateFeature,
} from "@/components/common/EmptyState";

interface FeatureGateProps {
  children: ReactNode;
  feature: string;
  description: string;
  highlights?: EmptyStateFeature[];
}

export default function FeatureGate({
  children,
  feature,
  description,
  highlights,
}: FeatureGateProps) {
  if (getConfig().cloud || getConfig().enterprise) {
    return <>{children}</>;
  }

  return (
    <EmptyState
      accent="yellow"
      icon={<LockClosedIcon className="w-8 h-8" />}
      overline="Premium Feature"
      title={feature}
      description={description}
      features={highlights}
      footnote="Available on ShellHub Cloud and Enterprise editions."
    >
      <a
        href="https://www.shellhub.io/pricing"
        target="_blank"
        rel="noopener noreferrer"
        className="inline-flex items-center gap-2 px-6 py-3 bg-primary hover:bg-primary-600 text-white rounded-lg text-sm font-semibold transition-all duration-200 shadow-lg shadow-primary/20 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-primary/50 focus-visible:ring-offset-2 focus-visible:ring-offset-background"
      >
        Pricing
        <ArrowTopRightOnSquareIcon className="w-4 h-4" strokeWidth={2} />
      </a>
    </EmptyState>
  );
}
