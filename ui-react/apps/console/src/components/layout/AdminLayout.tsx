import { Outlet, useLocation } from "react-router-dom";
import AdminSidebar from "./AdminSidebar";
import AdminAppBar from "./AdminAppBar";
import { SidebarMobileDrawer } from "./SidebarShell";
import { useSidebarLayout } from "../../hooks/useSidebarLayout";

export default function AdminLayout() {
  const { pathname } = useLocation();
  const { isOpen, pinned, isDesktop, drawerOpen, handlers } = useSidebarLayout();

  return (
    <div className="flex flex-col h-screen bg-background">
      <div className="flex flex-1 min-h-0 overflow-hidden">
        {isDesktop ? (
          <div
            onMouseEnter={handlers.onMouseEnter}
            onMouseLeave={handlers.onMouseLeave}
            onFocus={handlers.onFocus}
            onBlur={handlers.onBlur}
          >
            <AdminSidebar expanded={isOpen} pinned={pinned} onToggle={handlers.onToggle} />
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
            />
          </SidebarMobileDrawer>
        )}
        <div className="flex-1 flex flex-col min-w-0">
          <AdminAppBar onMenuToggle={isDesktop ? undefined : handlers.openDrawer} />
          <main className="flex-1 overflow-y-auto overflow-x-hidden p-4 sm:p-8 relative">
            <div className="grid-bg scanline absolute inset-0 -z-10" />
            <div
              key={pathname}
              className="page-enter flex-1 flex flex-col"
            >
              <Outlet />
            </div>
          </main>
        </div>
      </div>
    </div>
  );
}
