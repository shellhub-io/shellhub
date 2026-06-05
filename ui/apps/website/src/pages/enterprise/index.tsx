import { SiteLayout } from "@/components/SiteLayout";
import { HeroEnterprise } from "./HeroEnterprise";
import { AdminPanel } from "./AdminPanel";
import { SecurityFeatures } from "./SecurityFeatures";
import { DeploymentOptions } from "./DeploymentOptions";
import { SupportSection } from "./SupportSection";
import { EnterpriseCTA } from "./EnterpriseCTA";

export default function Enterprise() {
  return (
    <SiteLayout>
      <HeroEnterprise />
      <AdminPanel />
      <SecurityFeatures />
      <DeploymentOptions />
      <SupportSection />
      <EnterpriseCTA />
    </SiteLayout>
  );
}
