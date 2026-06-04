import { Outlet, useLocation } from "react-router-dom";
import AdminSidebar from "./AdminSidebar";
import AdminAppBar from "./AdminAppBar";
import { SidebarMobileDrawer } from "./SidebarShell";
import { useSidebarLayout } from "@/hooks/useSidebarLayout";

export default function AdminLayout() {
  const { pathname } = useLocation();
  const { isOpen, pinned, isDesktop, drawerOpen, handlers } =
    useSidebarLayout();

  return (
    <div className="flex flex-col h-screen bg-background">
      <div className="flex flex-1 min-h-0">
        {isDesktop ? (
          <div
            onMouseEnter={handlers.onMouseEnter}
            onMouseLeave={handlers.onMouseLeave}
            onFocus={handlers.onFocus}
            onBlur={handlers.onBlur}
          >
            <AdminSidebar
              expanded={isOpen}
              pinned={pinned}
              onToggle={handlers.onToggle}
            />
          </div>
        ) : (
          <SidebarMobileDrawer
            open={drawerOpen}
            onClose={handlers.closeDrawer}
            onKeyDown={handlers.onDrawerKeyDown}
          >
            <AdminSidebar
              expanded
              pinned={false}
              onToggle={handlers.closeDrawer}
              onClose={handlers.closeDrawer}
              toggleLabel="Close sidebar"
            />
          </SidebarMobileDrawer>
        )}
        <div className="flex flex-col size-full">
          <AdminAppBar
            onMenuToggle={isDesktop ? undefined : handlers.toggleDrawer}
          />
          <div className="relative size-full">
            <div className="grid-bg scanline absolute inset-0 -z-10" />
            <main
              key={pathname}
              className="page-enter absolute inset-0 p-8 pb-4 overflow-y-auto"
            >
              <Outlet />
            </main>
          </div>
        </div>
      </div>
    </div>
  );
}
