import SessionMenu from "./SessionMenu";
import SidebarShell, { navIcon } from "./SidebarShell";
import NavSectionList from "./NavSectionList";
import { useAdminNav } from "./adminNav";

/**
 * The admin navigation. Its links are instance-wide, so nothing here is filtered by namespace
 * role, only by whether the user is an instance admin at all.
 */
export default function AdminSidebar({
  expanded,
  onClose,
}: {
  expanded: boolean;
  onClose?: () => void;
}) {
  const { sections, disabled } = useAdminNav();

  return (
    <SidebarShell
      expanded={expanded}
      onClose={onClose}
      ariaLabel="Admin navigation"
      logoHref="/admin/dashboard"
      account={<SessionMenu expanded={expanded} />}
    >
      <NavSectionList
        sections={sections.map((section) => ({
          ...section,
          items: section.items.map((link) => ({
            ...link,
            icon: <link.icon className={navIcon} />,
          })),
        }))}
        expanded={expanded}
        disabled={disabled}
        onNavClick={onClose}
      />
    </SidebarShell>
  );
}
