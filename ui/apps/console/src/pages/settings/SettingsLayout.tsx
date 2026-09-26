import {
  Cog6ToothIcon,
  CommandLineIcon,
  CreditCardIcon,
  TicketIcon,
} from "@heroicons/react/24/outline";
import { Navigate, useLocation } from "react-router-dom";
import SectionedLayout from "@/components/settings/SectionedLayout";
import { useNavSectionTitle } from "@/components/layout/navSections";
import { isCloud } from "@/env";

/**
 * The namespace settings area, one section per URL under /settings. A link to the billing anchor
 * of the single page that came before lands on the Billing section.
 */
export default function SettingsLayout() {
  const sectionTitle = useNavSectionTitle("/settings");
  const { pathname, hash } = useLocation();
  const sections = [
    { to: "general", label: "General", icon: Cog6ToothIcon },
    { to: "ssh", label: "SSH", icon: CommandLineIcon },
    { to: "provisioning-keys", label: "Provisioning keys", icon: TicketIcon },
    ...(isCloud()
      ? [{ to: "billing", label: "Billing", icon: CreditCardIcon }]
      : []),
  ];

  if (hash === "#billing" && pathname === "/settings" && isCloud()) {
    return <Navigate to="billing" replace />;
  }

  return (
    <SectionedLayout
      base="/settings"
      icon={<Cog6ToothIcon className="w-6 h-6" />}
      overline={sectionTitle}
      title="Settings"
      description="How this namespace is named, reached over SSH, provisioned and billed"
      sections={sections}
    />
  );
}
