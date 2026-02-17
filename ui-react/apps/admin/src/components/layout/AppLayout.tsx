import { Outlet, useLocation } from "react-router-dom";
import Sidebar from "./Sidebar";
import AppBar from "./AppBar";
import TerminalManager from "../terminal/TerminalManager";
import CommandPalette from "../CommandPalette";

export default function AppLayout() {
  const { pathname } = useLocation();

  return (
    <div className="flex min-h-screen bg-background">
      <Sidebar />
      <div className="flex-1 flex flex-col min-w-0">
        <AppBar />
        <main className="flex-1 flex flex-col p-8 relative min-h-0">
          <div className="grid-bg scanline absolute inset-0 -z-10" />
          <div key={pathname} className="page-enter flex-1 flex flex-col min-h-0">
            <Outlet />
          </div>
        </main>
      </div>
      <TerminalManager />
      <CommandPalette />
    </div>
  );
}
