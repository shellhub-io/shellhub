import { Outlet, useLocation } from "react-router-dom";
import Sidebar from "./Sidebar";
import AppBar from "./AppBar";
import TerminalManager from "../terminal/TerminalManager";
import ConnectivityBanner from "../common/ConnectivityBanner";
import WelcomeWizardTrigger from "../wizard/WelcomeWizardTrigger";
import { SidebarMobileDrawer } from "./SidebarShell";
import { useNamespaces } from "@/hooks/useNamespaces";
import { useTerminalStore } from "@/stores/terminalStore";
import { useSidebarLayout } from "@/hooks/useSidebarLayout";

export default function AppLayout() {
  const { pathname } = useLocation();
  const { namespaces } = useNamespaces();
  const hasVisibleTerminal = useTerminalStore((s) =>
    s.sessions.some((t) => t.state !== "minimized"),
  );
  const { isOpen, isDesktop, drawerOpen, handlers } = useSidebarLayout();

  const showSidebar = namespaces.length > 0;
  const sidebarOffset =
    showSidebar && isDesktop ? (isOpen ? 220 : 60) : 0;

  return (
    <div
      className={`flex flex-col h-screen bg-background ${hasVisibleTerminal ? "overflow-hidden" : ""}`}
    >
      <ConnectivityBanner />
      <div className="flex flex-1 min-h-0">
        {showSidebar && isDesktop && (
          <Sidebar expanded={isOpen} />
        )}
        {showSidebar && !isDesktop && (
          <SidebarMobileDrawer
            open={drawerOpen}
            onClose={handlers.closeDrawer}
            onKeyDown={handlers.onDrawerKeyDown}
          >
            <Sidebar
              expanded
              onToggle={handlers.closeDrawer}
              onClose={handlers.closeDrawer}
            />
          </SidebarMobileDrawer>
        )}
        <div className="flex-1 flex flex-col min-w-0">
          <AppBar onMenuToggle={showSidebar && !isDesktop ? handlers.toggleDrawer : undefined} />
          <main className="flex-1 flex flex-col p-8 relative min-h-0 overflow-y-auto">
            <div className="grid-bg scanline absolute inset-0 -z-10" />
            <div
              key={pathname}
              className="page-enter flex-1 flex flex-col min-h-0"
            >
              <Outlet />
            </div>
          </main>
        </div>
      </div>
      <TerminalManager sidebarOffset={sidebarOffset} />
      <WelcomeWizardTrigger />
    </div>
  );
}
