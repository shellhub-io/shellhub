import { Outlet, useLocation } from "react-router-dom";
import Sidebar from "./Sidebar";
import AppBar from "./AppBar";
import TerminalManager from "../terminal/TerminalManager";
import ConnectivityBanner from "../common/ConnectivityBanner";
import WelcomeWizardTrigger from "../wizard/WelcomeWizardTrigger";
import { useNamespacesStore } from "../../stores/namespacesStore";

export default function AppLayout() {
  const { pathname } = useLocation();
  const namespaces = useNamespacesStore((s) => s.namespaces);

  return (
    <div className="flex flex-col min-h-screen bg-background">
      <ConnectivityBanner />
      <div className="flex flex-1 min-h-0">
        {namespaces.length > 0 && <Sidebar />}
        <div className="flex-1 flex flex-col min-w-0">
          <AppBar />
          <main className="flex-1 flex flex-col p-8 relative min-h-0">
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
      <TerminalManager />
      <WelcomeWizardTrigger />
    </div>
  );
}
