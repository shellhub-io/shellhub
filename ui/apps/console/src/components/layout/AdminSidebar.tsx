import { useTerminalStore } from "@/stores/terminalStore";
import { ADMIN_INSTANCE_PATH } from "@/utils/adminRoute";
import SidebarShell from "./SidebarShell";

/**
 * What stays of the sidebar in the admin console: the logo, over which the page frame slides. The
 * admin pages are reached from AdminNavBar across the top of the frame, and the account menu sits
 * beside the tabs, so the sidebar carries neither.
 */
export default function AdminSidebar({ expanded }: { expanded: boolean }) {
  const minimizeAll = useTerminalStore((s) => s.minimizeAll);

  return (
    <SidebarShell
      expanded={expanded}
      covered
      onNavigate={minimizeAll}
      logoHref={ADMIN_INSTANCE_PATH}
    />
  );
}
