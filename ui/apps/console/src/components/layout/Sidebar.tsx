import type { ReactNode } from "react";
import { isEnterpriseOrCloud } from "@/env";
import { useTerminalStore } from "@/stores/terminalStore";
import SidebarShell from "./SidebarShell";
import NavSectionList from "./NavSectionList";
import { useNavSections, type NavItem } from "./navSections";
import SessionMenu from "./SessionMenu";

function ProBadge() {
  return (
    <span className="text-2xs font-mono font-semibold text-accent-yellow/80 bg-accent-yellow/10 px-1.5 py-0.5 rounded">
      Pro
    </span>
  );
}

function BetaBadge() {
  return (
    <span className="text-2xs font-mono font-semibold text-accent-cyan/90 bg-accent-cyan/10 px-1.5 py-0.5 rounded">
      Beta
    </span>
  );
}

function pickBadge(item: NavItem): ReactNode | undefined {
  if (item.premium && !isEnterpriseOrCloud()) return <ProBadge />;
  if (item.beta) return <BetaBadge />;
  return undefined;
}

/**
 * The main navigation. Links the role cannot use are shown disabled rather than hidden, so the
 * shape of the product does not change with permissions.
 */
export default function Sidebar({
  expanded,
  covered,
  onClose,
}: {
  expanded: boolean;
  covered?: boolean;
  onClose?: () => void;
}) {
  const minimizeAll = useTerminalStore((s) => s.minimizeAll);

  const sections = useNavSections();

  const handleNavClick = () => {
    minimizeAll();
    onClose?.();
  };

  return (
    <SidebarShell
      expanded={expanded}
      covered={covered}
      onClose={onClose}
      ariaLabel="Main navigation"
      logoHref="/dashboard"
      account={<SessionMenu placement={expanded ? "expanded" : "rail"} />}
    >
      <NavSectionList
        sections={sections}
        expanded={expanded}
        onNavClick={handleNavClick}
        badge={pickBadge}
      />
    </SidebarShell>
  );
}
