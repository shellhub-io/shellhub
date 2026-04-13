import type { ReactNode } from "react";
import { NavLink } from "react-router-dom";
import { ChevronLeftIcon } from "@heroicons/react/24/outline";

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
      <span
        aria-disabled="true"
        className={`${navBase} ${navDisabled} ${align}`}
      >
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
        `${navBase} transition-all duration-150 ${isActive ? navActive : navIdle} ${align}`
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
    <div
      role="dialog"
      aria-modal={open}
      aria-label="Navigation menu"
      className={`fixed inset-0 z-50 ${open ? "" : "pointer-events-none"}`}
      onKeyDown={onKeyDown}
      {...(!open && { inert: "" })}
    >
      <div
        className={`absolute inset-0 bg-black/40 transition-opacity duration-200 ${
          open ? "opacity-100" : "opacity-0"
        }`}
        onClick={onClose}
        aria-hidden="true"
      />
      <div
        className={`fixed inset-y-0 left-0 z-50 w-[220px] transition-transform duration-200 ease-in-out ${
          open ? "translate-x-0" : "-translate-x-full"
        }`}
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
  hidden?: boolean;
  ariaLabel: string;
  footerLabel: string;
  logoHref: string;
  children: ReactNode;
}

export default function SidebarShell({
  expanded,
  pinned,
  onToggle,
  onClose,
  hidden,
  ariaLabel,
  footerLabel,
  logoHref,
  children,
}: SidebarShellProps) {
  return (
    <aside
      className={`bg-surface border-r border-border flex flex-col min-h-screen shrink-0 transition-all duration-200 ease-in-out overflow-hidden ${
        hidden ? "w-0 opacity-0" : expanded ? "w-[220px]" : "w-[60px]"
      }`}
    >
      {/* Logo */}
      <div className="h-14 flex items-center justify-center border-b border-border px-3">
        <NavLink
          to={logoHref}
          onClick={onClose}
          className="relative flex items-center justify-center"
        >
          <img
            src="/logo.svg"
            alt="ShellHub"
            className={`h-8 transition-opacity duration-200 ${expanded ? "opacity-100" : "opacity-0 absolute"}`}
          />
          <img
            src="/cloud-icon.svg"
            alt="ShellHub"
            className={`h-6 w-6 transition-opacity duration-200 ${expanded ? "opacity-0 absolute" : "opacity-100"}`}
          />
        </NavLink>
      </div>

      {/* Navigation (caller provides content) */}
      <nav
        aria-label={ariaLabel}
        className="flex-1 px-2 pt-4 pb-2 overflow-y-auto"
      >
        {children}
      </nav>

      {/* Footer with pin toggle */}
      <div
        className={`h-11 px-3 flex items-center justify-between transition-colors duration-200 ${expanded ? "border-t border-border" : "border-t border-transparent"}`}
      >
        <p
          className={`text-2xs font-mono text-text-muted/60 whitespace-nowrap transition-opacity duration-200 ${expanded ? "opacity-100" : "opacity-0"}`}
        >
          {footerLabel}
        </p>
        <button
          type="button"
          onClick={onToggle}
          tabIndex={expanded ? 0 : -1}
          aria-label={pinned ? "Unpin sidebar" : "Pin sidebar"}
          title={pinned ? "Unpin sidebar" : "Pin sidebar open"}
          className={`p-1 rounded-md transition-all duration-200 ${
            expanded ? "opacity-100" : "opacity-0"
          } ${
            pinned
              ? "text-primary bg-primary/10"
              : "text-text-muted hover:text-text-primary hover:bg-hover-subtle"
          }`}
        >
          <ChevronLeftIcon
            className={`w-3.5 h-3.5 transition-transform duration-200 ${
              expanded ? "" : "rotate-180"
            }`}
            strokeWidth={2}
          />
        </button>
      </div>
    </aside>
  );
}
