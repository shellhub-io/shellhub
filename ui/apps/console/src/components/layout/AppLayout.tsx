import { Outlet, useLocation } from "react-router-dom";
import Sidebar from "./Sidebar";
import AppBar from "./AppBar";
import TerminalManager from "../terminal/TerminalManager";
import ConnectivityBanner from "../common/ConnectivityBanner";
import LicenseBanner from "../common/LicenseBanner";
import DeviceLimitBanner from "@/components/common/DeviceLimitBanner";
import WelcomeWizardTrigger from "../wizard/WelcomeWizardTrigger";
import AnnouncementModalTrigger from "../announcements/AnnouncementModalTrigger";
import DeviceChooserTrigger from "../billing/DeviceChooserTrigger";
import { SidebarMobileDrawer } from "./SidebarShell";
import ChatwootProvider from "./ChatwootProvider";
import SkipToContentLink from "./SkipToContentLink";
import CommandPalette from "@/components/commandPalette/CommandPalette";
import { useNamespaces } from "@/hooks/useNamespaces";
import { useTerminalStore } from "@/stores/terminalStore";
import { useSidebarLayout } from "@/hooks/useSidebarLayout";
import VaultAutoLockBanner from "@/components/vault/VaultAutoLockBanner";
import { cn } from "@shellhub/design-system/cn";
import { getConfig } from "@/env";

export default function AppLayout() {
  const { enterprise, cloud } = getConfig();
  const isEnterprise = enterprise && !cloud;
  const { pathname } = useLocation();
  const { namespaces } = useNamespaces();
  const hasVisibleTerminal = useTerminalStore((s) =>
    s.sessions.some((t) => t.state !== "minimized"),
  );
  const { isOpen, pinned, isDesktop, drawerOpen, handlers } =
    useSidebarLayout();

  const showSidebar = namespaces.length > 0;
  const sidebarOffset = showSidebar && isDesktop ? (isOpen ? 220 : 60) : 0;

  return (
    <ChatwootProvider>
      <div
        className={cn("flex flex-col h-screen bg-background", hasVisibleTerminal && "overflow-hidden")}
      >
        <SkipToContentLink />
        <ConnectivityBanner />
        {isEnterprise && (
          <>
            <LicenseBanner />
            <DeviceLimitBanner />
          </>
        )}
        <div className="flex flex-1 min-h-0">
          {showSidebar && isDesktop && (
            <div
              onMouseEnter={handlers.onMouseEnter}
              onMouseLeave={handlers.onMouseLeave}
              onFocus={handlers.onFocus}
              onBlur={handlers.onBlur}
            >
              <Sidebar
                expanded={isOpen}
                pinned={pinned}
                onToggle={handlers.onToggle}
              />
            </div>
          )}
          {showSidebar && !isDesktop && (
            <SidebarMobileDrawer
              open={drawerOpen}
              onClose={handlers.closeDrawer}
              onKeyDown={handlers.onDrawerKeyDown}
            >
              <Sidebar
                expanded
                pinned={false}
                onToggle={handlers.closeDrawer}
                onClose={handlers.closeDrawer}
                toggleLabel="Close sidebar"
              />
            </SidebarMobileDrawer>
          )}
          <div className="flex flex-col size-full">
            <AppBar
              onMenuToggle={
                showSidebar && !isDesktop ? handlers.toggleDrawer : undefined
              }
            />
            <div className="relative size-full">
              <div className="grid-bg scanline absolute inset-0 -z-10" />
              <main
                id="main-content"
                tabIndex={-1}
                key={pathname}
                className="page-enter absolute inset-0 p-8 pb-4 overflow-y-auto"
              >
                <Outlet />
              </main>
              <div className="content-seam pointer-events-none absolute inset-0 z-10" />
            </div>
          </div>
        </div>
        <TerminalManager sidebarOffset={sidebarOffset} />
        <CommandPalette />
        <WelcomeWizardTrigger />
        <AnnouncementModalTrigger />
        <DeviceChooserTrigger />
        <VaultAutoLockBanner />
      </div>
    </ChatwootProvider>
  );
}
