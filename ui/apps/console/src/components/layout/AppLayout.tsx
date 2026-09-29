import { useEffect } from "react";
import { Outlet, useLocation } from "react-router-dom";
import Sidebar from "./Sidebar";
import AdminSidebar from "./AdminSidebar";
import AdminNavBar, { belowAdminNavBar } from "./AdminNavBar";
import SessionMenu from "./SessionMenu";
import { useSyncWorkspaceTab } from "@/hooks/useSyncWorkspaceTab";
import TerminalManager from "../terminal/TerminalManager";
import TabStrip from "./TabStrip";
import WindowControls from "./WindowControls";
import ConnectivityBanner from "../common/ConnectivityBanner";
import LicenseBanner from "../common/LicenseBanner";
import DeviceLimitBanner from "@/components/common/DeviceLimitBanner";
import WelcomeWizardTrigger from "../wizard/WelcomeWizardTrigger";
import DeviceChooserTrigger from "../billing/DeviceChooserTrigger";
import {
  SIDEBAR_EXPANDED_PX,
  SIDEBAR_RAIL_PX,
  SidebarMobileDrawer,
} from "./SidebarShell";
import ChatwootProvider from "./ChatwootProvider";
import SkipToContentLink from "./SkipToContentLink";
import CommandPalette from "@/components/commandPalette/CommandPalette";
import CreateNamespaceHost from "./CreateNamespaceHost";
import { useNamespaces } from "@/hooks/useNamespaces";
import { useWindowShown } from "@/stores/terminalStore";
import { useSidebarLayout } from "@/hooks/useSidebarLayout";
import { useTerminalThemeStore } from "@/stores/terminalThemeStore";
import { useScrollEdges } from "@/hooks/useScrollEdges";
import ScrollFades from "@/components/common/ScrollFades";
import VaultAutoLockBanner from "@/components/vault/VaultAutoLockBanner";
import { cn } from "@shellhub/design-system/cn";
import LogoMark from "./LogoMark";
import { isEnterprise } from "@/env";
import { isAdminPath } from "@/utils/adminRoute";
import { isAccountPath } from "@/utils/accountRoute";
import { isPreferencesPath } from "@/utils/preferencesRoute";

/**
 * The shell of the signed-in app: the sidebar and a tab strip sit on the chrome, and the
 * routed page and the terminal sessions share one framed panel below the tabs. The terminal lives
 * here rather than on a page, so a session survives navigating away from the device it belongs to.
 */
export default function AppLayout() {
  const { pathname } = useLocation();
  const loadTerminalThemes = useTerminalThemeStore((s) => s.loadThemes);
  useEffect(() => {
    void loadTerminalThemes();
  }, [loadTerminalThemes]);
  const { namespaces } = useNamespaces();
  const windowShown = useWindowShown();
  const { isOpen, pinned, isDesktop, isWide, drawerOpen, handlers } =
    useSidebarLayout();
  const {
    ref: scrollRef,
    moreAbove,
    moreBelow,
  } = useScrollEdges<HTMLElement>();

  const isAdminRoute = isAdminPath(pathname);
  const showSidebar = isAdminRoute || namespaces.length > 0;
  const drawerNav = showSidebar && !isDesktop && !isAdminRoute;
  const desktopSidebar = showSidebar && isDesktop;
  useSyncWorkspaceTab(pathname, isAdminRoute);
  const frameOverNav =
    desktopSidebar &&
    (windowShown ||
      isAdminRoute ||
      isAccountPath(pathname) ||
      isPreferencesPath(pathname));
  const sessionMenuInTabStrip = frameOverNav || !desktopSidebar;

  return (
    <ChatwootProvider>
      <div
        className={cn(
          "flex flex-col h-screen bg-background",
          windowShown && "overflow-hidden",
        )}
      >
        <SkipToContentLink />
        <ConnectivityBanner />
        {isEnterprise() && (
          <>
            <LicenseBanner />
            <DeviceLimitBanner />
          </>
        )}
        <div className="bg-background flex flex-1 min-h-0">
          {desktopSidebar && (
            <div
              style={{
                width: !pinned ? SIDEBAR_RAIL_PX : undefined,
              }}
              className="relative shrink-0"
            >
              <div
                onMouseEnter={handlers.onMouseEnter}
                onMouseLeave={handlers.onMouseLeave}
                onFocus={handlers.onFocus}
                onBlur={handlers.onBlur}
                className={cn(
                  "h-full",
                  !pinned &&
                    "absolute inset-y-0 left-0 border-r border-transparent transition-[box-shadow,border-color] duration-200",
                  !pinned && !frameOverNav && "z-appbar",
                  !pinned &&
                    isOpen &&
                    !frameOverNav &&
                    "border-border shadow-[16px_0_40px_-12px_rgba(0,0,0,0.7)]",
                )}
              >
                {isAdminRoute ? (
                  <AdminSidebar
                    expanded={isOpen && pinned}
                  />
                ) : (
                  <Sidebar
                    expanded={isOpen && (pinned || !frameOverNav)}
                    covered={frameOverNav}
                  />
                )}
              </div>
            </div>
          )}
          {drawerNav && (
            <SidebarMobileDrawer
              open={drawerOpen}
              onClose={handlers.closeDrawer}
              onKeyDown={handlers.onDrawerKeyDown}
            >
              <Sidebar
                expanded
                withAccount={false}
                onClose={handlers.closeDrawer}
              />
            </SidebarMobileDrawer>
          )}
          <div
            className={cn(
              "flex flex-col flex-1 min-w-0 pr-2 pb-2",
              !desktopSidebar && "pl-2",
            )}
          >
            <TabStrip
              leading={
                !desktopSidebar && (
                  <LogoMark
                    full={isDesktop}
                    className="mr-3 mb-[5px] shrink-0"
                  />
                )
              }
              trailing={
                <>
                  {sessionMenuInTabStrip && (
                    <SessionMenu
                      placement={isDesktop ? "tabStrip" : "tabStripAvatar"}
                    />
                  )}
                  <WindowControls
                    onOpenNavigation={
                      drawerNav ? handlers.toggleDrawer : undefined
                    }
                  />
                </>
              }
            />
            <div
              style={{
                marginLeft: frameOverNav
                  ? `calc(-${pinned ? SIDEBAR_EXPANDED_PX : SIDEBAR_RAIL_PX}px + 0.5rem)`
                  : undefined,
              }}
              className={cn(
                "relative flex-1 min-h-0 overflow-hidden rounded-[10px] border border-border bg-surface",
                "transition-[margin-left] duration-200 ease-in-out",
                !frameOverNav &&
                  "peer-data-[first-tab-active=true]:rounded-tl-none",
              )}
            >
              <div className="grid-bg scanline absolute inset-0 z-bg" />
              {isAdminRoute && <AdminNavBar compact={!isWide} />}
              <main
                id="main-content"
                ref={scrollRef}
                tabIndex={-1}
                key={pathname}
                className={cn(
                  "page-enter absolute inset-0 p-8 pb-4 overflow-y-auto outline-none",
                  isAdminRoute && belowAdminNavBar,
                )}
              >
                <Outlet />
              </main>
              <ScrollFades moreAbove={moreAbove} moreBelow={moreBelow} />
              <TerminalManager />
            </div>
          </div>
        </div>
        <CommandPalette />
        <CreateNamespaceHost />
        <WelcomeWizardTrigger />
        <DeviceChooserTrigger />
        <VaultAutoLockBanner />
      </div>
    </ChatwootProvider>
  );
}
