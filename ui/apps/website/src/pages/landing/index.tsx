import { SiteLayout } from "@/components/SiteLayout";
import { Hero } from "./Hero";
import { TrustedBy } from "./TrustedBy";
import { QuickStart } from "./QuickStart";
import { SupportedPlatforms } from "./SupportedPlatforms";
import { HowItWorks } from "./HowItWorks";
import { Architecture } from "./Architecture";
import { FeatureGrid } from "./FeatureGrid";
import { OpenSource } from "./OpenSource";
import { CTA } from "./CTA";

export default function Landing() {
  return (
    <SiteLayout>
      <Hero />
      <TrustedBy />
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
