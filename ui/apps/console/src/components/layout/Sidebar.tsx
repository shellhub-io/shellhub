import type { ReactNode } from "react";
import { isEnterpriseOrCloud } from "@/env";
import { useTerminalStore } from "@/stores/terminalStore";
import { cn } from "@shellhub/design-system/cn";
import SidebarShell, { NavItemLink } from "./SidebarShell";
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
  onClose,
}: {
  expanded: boolean;
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
      onClose={onClose}
      ariaLabel="Main navigation"
      logoHref="/dashboard"
      account={<SessionMenu expanded={expanded} />}
    >
      {sections.map((section, idx) => (
        <div key={section.items[0]?.to} className={idx > 0 ? "mt-2.5" : ""}>
          {section.title && (
            <div className="relative mb-1.5 h-[14px]">
              <p
                className={cn(
                  "px-3 text-2xs leading-[14px] font-mono font-semibold uppercase tracking-label text-text-muted/60 whitespace-nowrap overflow-hidden transition-opacity duration-200",
                  expanded ? "opacity-100" : "opacity-0",
                )}
              >
                {section.title}
              </p>
              <span
                aria-hidden="true"
                className={cn(
                  "absolute left-3 right-3 top-[3px] h-px bg-border transition-opacity duration-200",
                  expanded ? "opacity-0" : "opacity-100",
                )}
              />
            </div>
          )}
          <div className="space-y-0.5">
            {section.items.map((item) => (
              <NavItemLink
                key={item.to}
                item={item}
                expanded={expanded}
                onClick={handleNavClick}
                badge={pickBadge(item)}
              />
            ))}
          </div>
        </div>
      ))}
    </SidebarShell>
  );
}
