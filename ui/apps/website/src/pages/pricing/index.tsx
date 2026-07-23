import { SiteLayout } from "@/components";
import { HeroPricing } from "./HeroPricing";
import { TierCards } from "./TierCards";
import { ComparisonTable } from "./ComparisonTable";
import { PricingFAQ } from "./PricingFAQ";

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
