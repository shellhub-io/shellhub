import { NavLink, Outlet } from "react-router-dom";
import {
  ComputerDesktopIcon,
  PlusIcon,
  ServerStackIcon,
} from "@heroicons/react/24/outline";
import { cn } from "@shellhub/design-system/cn";
import PageHeader from "@/components/common/PageHeader";
import { useNavSectionTitle } from "@/components/layout/navSections";

const TABS = [
  {
    to: "/devices/add",
    icon: ComputerDesktopIcon,
    title: "Interactive",
    sub: "Install it, then accept in your browser.",
  },
  {
    to: "/devices/add/fleet",
    icon: ServerStackIcon,
    title: "Fleet",
    sub: "Provision many, unattended, with a provisioning key.",
  },
];

/**
 * Where devices come in from: interactively, with someone at the machine to accept it, or as a
 * fleet provisioned with one of the namespace's provisioning keys. Each tab has its own URL, and
 * Interactive is the default.
 */
export default function AddDevice() {
  const overline = useNavSectionTitle("/devices");

  return (
    <div className="max-w-3xl">
      <PageHeader
        icon={<PlusIcon className="w-6 h-6" />}
        overline={overline}
        title="Add Device"
        description="Install the ShellHub agent to connect a device to your namespace"
      />

      <nav
        aria-label="How to add"
        className="flex gap-1 border-b border-border mb-8"
      >
        {TABS.map((tab) => (
          <NavLink
            key={tab.to}
            to={tab.to}
            end
            className={({ isActive }) =>
              cn(
                "group flex items-start gap-2.5 px-4 py-3 -mb-px border-b-2 transition-colors max-w-[20rem]",
                isActive
                  ? "border-primary text-primary"
                  : "border-transparent text-text-muted hover:text-text-secondary",
              )
            }
          >
            <tab.icon className="w-4 h-4 shrink-0 mt-0.5" strokeWidth={1.8} />
            <span className="flex flex-col gap-0.5 min-w-0">
              <span className="text-sm font-semibold leading-tight">
                {tab.title}
              </span>
              <span className="text-2xs leading-snug text-text-muted hidden sm:block">
                {tab.sub}
              </span>
            </span>
          </NavLink>
        ))}
      </nav>

      <Outlet />
    </div>
  );
}
