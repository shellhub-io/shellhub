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
  "flex items-center gap-3 h-[38px] px-3 rounded-md text-[13px] font-medium whitespace-nowrap [&>svg]:shrink-0 focus-visible:relative focus-visible:z-raised";
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

/**
 * The width of the sidebar folded to a rail. The layout reserves this much beside the content
 * while the rail floats over it, so the two must not drift apart.
 */
export const railWidth = "w-[60px]";

interface NavItemLinkProps {
  item: { to: string; label: string; icon: ReactNode };
  expanded: boolean;
  disabled?: boolean;
  onClick?: () => void;
  badge?: ReactNode;
}

/**
 * One navigation link. Collapsed, the label fades out but stays in the accessibility tree, so a
 * collapsed sidebar is still navigable by screen reader. The icon keeps one position in both
 * states, which the rail's width centres, so folding moves nothing but the label.
 */
export function NavItemLink({
  item,
  expanded,
  disabled,
  onClick,
  badge,
}: NavItemLinkProps) {
  const label = (
    <span
      className={cn(
        "flex-1 min-w-0 truncate transition-opacity duration-200",
        expanded ? "opacity-100" : "opacity-0",
      )}
    >
      {item.label}
    </span>
  );

  if (disabled) {
    return (
      <span aria-disabled="true" className={cn(navBase, navDisabled)}>
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
        "flex flex-col h-full shrink-0 bg-background transition-all duration-200 ease-in-out overflow-hidden",
        hidden ? "w-0 opacity-0" : expanded ? "w-[220px]" : railWidth,
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
        className="flex-1 -mt-1 px-2 pt-1 pb-2 overflow-y-auto"
      >
        {children}
      </nav>

      <div className="min-h-14 px-2 py-2 flex items-center">
        <div className="min-w-0 flex-1">{account}</div>
      </div>
    </aside>
  );
}
