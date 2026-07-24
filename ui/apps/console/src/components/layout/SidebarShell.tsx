import type { ReactNode } from "react";
import { NavLink } from "react-router-dom";
import { ChevronLeftIcon } from "@heroicons/react/24/outline";
import { cn } from "@shellhub/design-system/cn";
import {
  IconButton,
  ShellHubCloudIcon,
  ShellHubLogo,
} from "@shellhub/design-system/primitives";

// ---- Shared nav style constants ----

export const navBase =
  "flex items-center gap-3 px-3 py-2 rounded-md text-[13px] font-medium";
const navActive = "bg-primary/10 text-primary border border-primary/20";
const navIdle =
  "text-text-secondary hover:text-text-primary hover:bg-hover-subtle border border-transparent";
export const navDisabled = "text-text-muted/50 cursor-not-allowed";
export const navIcon = "w-[18px] h-[18px]";

// ---- NavItemLink ----

interface NavItemLinkProps {
  item: { to: string; label: string; icon: ReactNode };
  expanded: boolean;
  disabled?: boolean;
  onClick?: () => void;
  badge?: ReactNode;
}

export function NavItemLink({
  item,
  expanded,
  disabled,
  onClick,
  badge,
}: NavItemLinkProps) {
  const align = expanded ? "" : "justify-center";
  const label = expanded ? (
    <span className="flex-1 truncate">{item.label}</span>
  ) : null;

  if (disabled) {
    return (
      <span aria-disabled="true" className={cn(navBase, navDisabled, align)}>
        {item.icon}
        {label}
      </span>
    );
  }

  return (
    <NavLink
      to={item.to}
      title={expanded ? undefined : item.label}
      onClick={onClick}
      className={({ isActive }) =>
        cn(
          navBase,
          "transition-all duration-150",
          isActive ? navActive : navIdle,
          align,
        )
      }
    >
      {item.icon}
      {label}
      {expanded && badge}
    </NavLink>
  );
}

// ---- SidebarMobileDrawer ----

interface SidebarMobileDrawerProps {
  open: boolean;
  onClose: () => void;
  onKeyDown: (e: React.KeyboardEvent) => void;
  children: ReactNode;
}

export function SidebarMobileDrawer({
  open,
  onClose,
  onKeyDown,
  children,
}: SidebarMobileDrawerProps) {
  return (
    // eslint-disable-next-line jsx-a11y/no-noninteractive-element-interactions
    <div
      role="dialog"
      aria-modal={open}
      aria-label="Navigation menu"
      className={cn(
        "fixed inset-0 z-drawer-backdrop",
        !open && "pointer-events-none",
      )}
      onKeyDown={onKeyDown}
      {...(!open && { inert: true })}
    >
      <div
        className={cn(
          "absolute inset-0 bg-black/40 transition-opacity duration-200",
          open ? "opacity-100" : "opacity-0",
        )}
        onClick={onClose}
        aria-hidden="true"
      />
      <div
        className={cn(
          "fixed inset-y-0 left-0 z-drawer w-[220px] transition-transform duration-200 ease-in-out",
          open ? "translate-x-0" : "-translate-x-full",
        )}
      >
        {children}
      </div>
    </div>
  );
}

// ---- SidebarShell ----

interface SidebarShellProps {
  expanded: boolean;
  pinned: boolean;
  onToggle: () => void;
  onClose?: () => void;
  toggleLabel?: string;
  hidden?: boolean;
  ariaLabel: string;
  footerLabel: string;
  logoHref: string;
  /** Optional fixed band rendered between the logo and the scrollable nav. */
  headerSlot?: ReactNode;
  children: ReactNode;
}

export default function SidebarShell({
  expanded,
  pinned,
  onToggle,
  onClose,
  toggleLabel: toggleLabelOverride,
  hidden,
  ariaLabel,
  footerLabel,
  logoHref,
  headerSlot,
  children,
}: SidebarShellProps) {
  const toggleLabel =
    toggleLabelOverride ?? (pinned ? "Unpin sidebar" : "Pin sidebar");
  const toggleTitle =
    toggleLabelOverride ?? (pinned ? "Unpin sidebar" : "Pin sidebar open");

  return (
    <aside
      className={cn(
        "theme-dark bg-surface border-r border-border flex flex-col h-full shrink-0 transition-all duration-200 ease-in-out overflow-hidden",
        hidden ? "w-0 opacity-0" : expanded ? "w-[220px]" : "w-[60px]",
      )}
    >
      {/* Logo */}
      <div className="h-14 flex items-center justify-center border-b border-border px-3">
        <NavLink
          to={logoHref}
          onClick={onClose}
          className="relative flex items-center justify-center"
          aria-label="ShellHub"
        >
          <ShellHubLogo
            aria-hidden
            className={cn(
              "h-8 transition-opacity duration-200",
              expanded ? "opacity-100" : "opacity-0 absolute",
            )}
          />
          <ShellHubCloudIcon
            aria-hidden
            data-testid="sidebar-cloud-icon"
            className={cn(
              "h-6 w-6 transition-opacity duration-200",
              expanded ? "opacity-0 absolute" : "opacity-100",
            )}
          />
        </NavLink>
      </div>

      {/* Optional fixed header band (e.g. command-palette trigger). Skipped
          when the shell is hidden so its control isn't focusable off-screen. */}
      {headerSlot && !hidden && (
        <div className="px-2 py-2.5 border-b border-border">{headerSlot}</div>
      )}

      {/* Navigation (caller provides content) */}
      <nav
        aria-label={ariaLabel}
        className="flex-1 px-2 pt-4 pb-2 overflow-y-auto"
      >
        {children}
      </nav>

      {/* Footer with context-specific sidebar toggle */}
      <div
        className={cn(
          "h-11 px-3 flex items-center justify-between transition-colors duration-200",
          expanded ? "border-t border-border" : "border-t border-transparent",
        )}
      >
        <p
          className={cn(
            "text-2xs font-mono text-text-muted/60 whitespace-nowrap transition-opacity duration-200",
            expanded ? "opacity-100" : "opacity-0",
          )}
        >
          {footerLabel}
        </p>
        <IconButton
          size="sm"
          onClick={onToggle}
          tabIndex={expanded ? 0 : -1}
          aria-label={toggleLabel}
          title={toggleTitle}
          className={cn(
            "duration-200",
            expanded ? "opacity-100" : "opacity-0",
            pinned && "text-primary bg-primary/10",
          )}
        >
          <ChevronLeftIcon
            className={cn(
              "w-3.5 h-3.5 transition-transform duration-200",
              !expanded && "rotate-180",
            )}
            strokeWidth={2}
          />
        </IconButton>
      </div>
    </aside>
  );
}
