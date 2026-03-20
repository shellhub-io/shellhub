import { useState, useCallback } from "react";
import { Outlet, useLocation } from "react-router-dom";
import AdminSidebar from "./AdminSidebar";
import AdminAppBar from "./AdminAppBar";

export default function AdminLayout() {
  const [expanded, setExpanded] = useState(false);
  const [pinned, setPinned] = useState(false);
  const { pathname } = useLocation();

  const isOpen = expanded || pinned;

  const handleToggle = useCallback(() => {
    setPinned((prev) => !prev);
  }, []);

  return (
    <div className="flex flex-col min-h-screen bg-background">
      <div className="flex flex-1 min-h-0">
        <div
          onMouseEnter={() => setExpanded(true)}
          onMouseLeave={() => setExpanded(false)}
        >
          <AdminSidebar expanded={isOpen} onToggle={handleToggle} />
        </div>
        <div className="flex-1 flex flex-col min-w-0">
          <AdminAppBar />
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
    </div>
  );
}
