import type { ReactNode } from "react";
import { NavLink } from "react-router-dom";
import { cn } from "@shellhub/design-system/cn";
import { useTerminalFullscreen } from "@/stores/terminalStore";
import LogoMark from "./LogoMark";

/**
 * The classes every sidebar link shares, so an active, idle and disabled link differ only in
 * colour and cannot drift apart in spacing or type.
 */
export const navBase =
  "flex items-center gap-3 px-3 py-2 rounded-md text-[13px] font-medium";
const navActive = "bg-primary/10 text-primary border border-primary/20";
const navIdle =
  "text-text-secondary hover:text-text-primary hover:bg-hover-subtle border border-transparent";
/**
 * Added to a link the current role may not follow. It dims and blocks the pointer, but the link
 * stays in place — see Sidebar.
 */
export const navDisabled = "text-text-muted/50 cursor-not-allowed";
/**
 * The icon size every sidebar link uses, so the labels line up whatever icon is beside them.
 */
export const navIcon = "w-[18px] h-[18px]";

interface NavItemLinkProps {
  item: { to: string; label: string; icon: ReactNode };
  expanded: boolean;
  disabled?: boolean;
  onClick?: () => void;
  badge?: ReactNode;
}

/**
 * One navigation link. Collapsed, the label is hidden but stays in the accessibility tree, so a
 * collapsed sidebar is still navigable by screen reader.
 */
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

interface SidebarMobileDrawerProps {
  open: boolean;
  onClose: () => void;
  onKeyDown: (e: React.KeyboardEvent) => void;
  children: ReactNode;
}

/**
 * The sidebar as a drawer, for viewports too narrow to hold it open. Traps focus while open, so
 * the page behind cannot be tabbed into.
 */
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
          "fixed inset-y-0 z-drawer w-[220px] transition-transform duration-200 ease-in-out",
          "left-0 border-r border-border",
          open ? "translate-x-0" : "-translate-x-full",
        )}
      >
        {children}
      </div>
    </div>
  );
}

interface SidebarShellProps {
  expanded: boolean;
  onClose?: () => void;
  ariaLabel: string;
  logoHref: string;
  account: ReactNode;
  children: ReactNode;
}

/**
 * The frame both sidebars are built in, on the page background with the logo on top and the
 * account menu at the foot, so the app and admin navigations differ only in their links. It
 * folds away while a terminal is fullscreen, when the tab strip shows the logo instead.
 */
export default function SidebarShell({
  expanded,
  onClose,
  ariaLabel,
  logoHref,
  account,
  children,
}: SidebarShellProps) {
  const hidden = useTerminalFullscreen();

  return (
    <aside
      className={cn(
        "theme-dark flex flex-col h-full shrink-0 bg-background transition-all duration-200 ease-in-out overflow-hidden",
        hidden ? "w-0 opacity-0" : expanded ? "w-[220px]" : "w-[60px]",
      )}
    >
      <div
        data-tauri-drag-region
        className="flex items-end h-12 pl-[12.7px] pb-[7px]"
      >
        <NavLink to={logoHref} onClick={onClose} aria-label="ShellHub">
          <LogoMark full={expanded} />
        </NavLink>
      </div>

      <nav
        aria-label={ariaLabel}
        className="flex-1 px-2 pt-4 pb-2 overflow-y-auto"
      >
        {children}
      </nav>

      <div className="min-h-14 px-2 py-2 flex items-center">
        <div className="min-w-0 flex-1">{account}</div>
      </div>
    </aside>
  );
}
