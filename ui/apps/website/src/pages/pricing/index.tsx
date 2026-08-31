import { SiteLayout } from "@/components";
import { HeroPricing } from "./HeroPricing";
import { TierCards } from "./TierCards";
import { ComparisonTable } from "./ComparisonTable";
import { PricingFAQ } from "./PricingFAQ";

/**
 * The pricing page: the plans, then the full comparison, then the questions the plans raise.
 */
export default function Pricing() {
  return (
    <SiteLayout>
      <HeroPricing />
      <TierCards />
      <ComparisonTable />
      <PricingFAQ />
    </SiteLayout>
  );
}
