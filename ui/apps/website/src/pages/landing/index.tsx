import { SiteLayout } from "@/components";
import { Hero } from "./Hero";
import { QuickStart } from "./QuickStart";
import { SupportedPlatforms } from "./SupportedPlatforms";
import { HowItWorks } from "./HowItWorks";
import { Architecture } from "./Architecture";
import { FeatureGrid } from "./FeatureGrid";
import { OpenSource } from "./OpenSource";
import { CTA } from "./CTA";

/**
 * The landing page. Section order is deliberate: the pitch, then how to try it, then how it
 * works, then why to trust it.
 */
export default function Landing() {
  return (
    <SiteLayout>
      <Hero />
      <QuickStart />
      <SupportedPlatforms />
      <HowItWorks />
      <Architecture />
      <FeatureGrid />
      <OpenSource />
      <CTA />
    </SiteLayout>
  );
}
