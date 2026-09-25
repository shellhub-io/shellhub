import { useEffect } from "react";
import { Outlet, useLocation } from "react-router-dom";
import { ChevronDownIcon } from "@heroicons/react/24/outline";
import Sidebar from "./Sidebar";
import AdminSidebar from "./AdminSidebar";
import { useSyncWorkspaceTab } from "@/hooks/useSyncWorkspaceTab";
import TerminalManager from "../terminal/TerminalManager";
import TabStrip from "./TabStrip";
import WindowControls from "./WindowControls";
import ConnectivityBanner from "../common/ConnectivityBanner";
import LicenseBanner from "../common/LicenseBanner";
import DeviceLimitBanner from "@/components/common/DeviceLimitBanner";
import WelcomeWizardTrigger from "../wizard/WelcomeWizardTrigger";
import AnnouncementModalTrigger from "../announcements/AnnouncementModalTrigger";
import DeviceChooserTrigger from "../billing/DeviceChooserTrigger";
import { SidebarMobileDrawer, railWidth } from "./SidebarShell";
import ChatwootProvider from "./ChatwootProvider";
import SkipToContentLink from "./SkipToContentLink";
import CommandPalette from "@/components/commandPalette/CommandPalette";
import CreateNamespaceHost from "./CreateNamespaceHost";
import { useNamespaces } from "@/hooks/useNamespaces";
import {
  useTerminalFullscreen,
  useTerminalStore,
} from "@/stores/terminalStore";
import { useSidebarLayout } from "@/hooks/useSidebarLayout";
import { useTerminalThemeStore } from "@/stores/terminalThemeStore";
import { useScrollEdges } from "@/hooks/useScrollEdges";
import VaultAutoLockBanner from "@/components/vault/VaultAutoLockBanner";
import { cn } from "@shellhub/design-system/cn";
import LogoMark from "./LogoMark";
import { isEnterprise } from "@/env";
import { isAdminPath } from "@/utils/adminRoute";

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
  const windowInView = useTerminalStore(
    (s) =>
      s.sessions.some((t) => t.state !== "minimized") ||
      s.recordings.some((r) => r.shown),
  );
  const terminalFullscreen = useTerminalFullscreen();
  const { isOpen, pinned, isDesktop, drawerOpen, handlers } =
    useSidebarLayout();
  const {
    ref: scrollRef,
    moreAbove,
    moreBelow,
  } = useScrollEdges<HTMLElement>();

  const isAdminRoute = isAdminPath(pathname);
  const showSidebar = isAdminRoute || namespaces.length > 0;
  const NavSidebar = isAdminRoute ? AdminSidebar : Sidebar;
  useSyncWorkspaceTab(pathname, isAdminRoute);

  return (
    <ChatwootProvider>
      <div
        className={cn(
          "flex flex-col h-screen bg-background",
          windowInView && "overflow-hidden",
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
          {showSidebar && isDesktop && (
            <div
              className={cn(
                "relative shrink-0",
                !pinned && !terminalFullscreen && railWidth,
              )}
            >
              <div
                onMouseEnter={handlers.onMouseEnter}
                onMouseLeave={handlers.onMouseLeave}
                onFocus={handlers.onFocus}
                onBlur={handlers.onBlur}
                className={cn(
                  "h-full",
                  !pinned &&
                    "absolute inset-y-0 left-0 z-appbar border-r border-transparent transition-[box-shadow,border-color] duration-200",
                  !pinned &&
                    isOpen &&
                    "border-border shadow-[16px_0_40px_-12px_rgba(0,0,0,0.7)]",
                )}
              >
                <NavSidebar expanded={isOpen} />
              </div>
            </div>
          )}
          {showSidebar && !isDesktop && (
            <SidebarMobileDrawer
              open={drawerOpen}
              onClose={handlers.closeDrawer}
              onKeyDown={handlers.onDrawerKeyDown}
            >
              <NavSidebar expanded onClose={handlers.closeDrawer} />
            </SidebarMobileDrawer>
          )}
          <div
            className={cn(
              "flex flex-col flex-1 min-w-0 pr-2 pb-2",
              !(showSidebar && isDesktop && !terminalFullscreen) && "pl-2",
            )}
          >
            <TabStrip
              leading={
                (!isDesktop || terminalFullscreen) && (
                  <LogoMark
                    full={isDesktop}
                    className="mr-3 mb-[7px] shrink-0"
                  />
                )
              }
              trailing={
                <WindowControls
                  sidebarPinned={pinned}
                  sidebarMode={isDesktop ? "pin" : "drawer"}
                  onToggleSidebar={
                    showSidebar
                      ? isDesktop
                        ? handlers.onToggle
                        : handlers.toggleDrawer
                      : undefined
                  }
                />
              }
            />
            <div
              className={cn(
                "relative flex-1 min-h-0 overflow-hidden rounded-[10px] border border-border bg-surface",
                "peer-data-[first-tab-active=true]:rounded-tl-none",
              )}
            >
              <div className="grid-bg scanline absolute inset-0 z-bg" />
              <main
                id="main-content"
                ref={scrollRef}
                tabIndex={-1}
                key={pathname}
                className="page-enter absolute inset-0 p-8 pb-4 overflow-y-auto outline-none"
              >
                <Outlet />
              </main>
              <div
                aria-hidden="true"
                className={cn(
                  "pointer-events-none absolute inset-x-0 top-0 h-8 bg-gradient-to-b from-surface to-transparent transition-opacity duration-200",
                  moreAbove ? "opacity-100" : "opacity-0",
                )}
              />
              <div
                aria-hidden="true"
                className={cn(
                  "pointer-events-none absolute inset-x-0 bottom-0 h-10 flex items-end justify-center pb-1 bg-gradient-to-t from-surface to-transparent transition-opacity duration-200",
                  moreBelow ? "opacity-100" : "opacity-0",
                )}
              >
                <ChevronDownIcon
                  className="w-4 h-4 text-text-muted/70"
                  strokeWidth={2}
                />
              </div>
              <TerminalManager />
            </div>
          </div>
        </div>
        <CommandPalette />
        <CreateNamespaceHost />
        <WelcomeWizardTrigger />
        <AnnouncementModalTrigger />
        <DeviceChooserTrigger />
        <VaultAutoLockBanner />
      </div>
    </ChatwootProvider>
  );
}
