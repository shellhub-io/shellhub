import BillingSection from "@/components/billing/BillingSection";
import SettingsSection from "@/components/settings/SettingsSection";

/**
 * The Billing section of a cloud namespace's settings. The rows and the dialogs they open live
 * with the rest of billing in components/billing; this only titles them as a section.
 */
export default function BillingSettings() {
  return (
    <SettingsSection
      title="Billing"
      description="The plan this namespace is on and how it is paid for."
    >
      <BillingSection />
    </SettingsSection>
  );
}
