import { ReactNode } from "react";
import {
  LockClosedIcon,
  ArrowTopRightOnSquareIcon,
} from "@heroicons/react/24/outline";
import { getConfig } from "@/env";
import { Button } from "@shellhub/design-system/primitives";
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
      <Button
        as="a"
        href="https://www.shellhub.io/pricing"
        target="_blank"
        rel="noopener noreferrer"
        size="lg"
        glow
        iconRight={
          <ArrowTopRightOnSquareIcon className="w-4 h-4" strokeWidth={2} />
        }
      >
        Pricing
      </Button>
    </EmptyState>
  );
}
