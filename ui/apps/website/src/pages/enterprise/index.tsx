import { SiteLayout } from "@/components";
import { HeroEnterprise } from "./HeroEnterprise";
import { AdminPanel } from "./AdminPanel";
import { SecurityFeatures } from "./SecurityFeatures";
import { DeploymentOptions } from "./DeploymentOptions";
import { SupportSection } from "./SupportSection";
import { EnterpriseCTA } from "./EnterpriseCTA";

/**
 * The enterprise page. Section order is the argument it makes: what the panel does, how it is
 * secured, where it runs, who supports it, then the call to action.
 */
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
