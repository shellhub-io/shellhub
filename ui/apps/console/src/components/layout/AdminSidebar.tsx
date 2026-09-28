import SidebarShell from "./SidebarShell";

/**
 * What stays of the sidebar in the admin console: the logo, over which the page frame slides. The
 * admin pages are reached from AdminNavBar across the top of the frame, and the account menu sits
 * beside the tabs, so the sidebar carries neither.
 */
export default function AdminSidebar({
  expanded,
  folded,
}: {
  expanded: boolean;
  folded?: boolean;
}) {
  return (
    <SidebarShell
      expanded={expanded}
      folded={folded}
      covered
      logoHref="/admin/dashboard"
    />
  );
}
