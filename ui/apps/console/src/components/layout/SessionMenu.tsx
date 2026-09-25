import { useState } from "react";
import { ChevronUpDownIcon } from "@heroicons/react/24/outline";
import { Dropdown } from "@shellhub/design-system/primitives";
import { cn } from "@shellhub/design-system/cn";
import { useAuthStore } from "@/stores/authStore";
import { getInitials } from "@/utils/string";
import AccountMenuItems from "./AccountMenuItems";

/**
 * The signed-in user at the bottom of the sidebar, opening upward into the account actions.
 */
export default function SessionMenu({ expanded }: { expanded: boolean }) {
  const { user, name, email } = useAuthStore();
  const [open, setOpen] = useState(false);

  const display = user || name || email || "Account";

  return (
    <Dropdown
      mode="content"
      placement="top-start"
      portal
      open={open}
      onOpenChange={setOpen}
    >
      <Dropdown.Trigger>
        <button
          type="button"
          data-testid="session-menu"
          aria-label={`Account menu for ${display}`}
          title={expanded ? undefined : display}
          className={cn(
            "w-full flex items-center rounded-lg border border-border bg-card hover:border-border-light transition-colors duration-150",
            expanded ? "gap-2.5 p-2" : "justify-center p-1.5",
          )}
        >
          <span className="w-7 h-7 rounded-full bg-primary/15 border border-primary/20 flex items-center justify-center text-primary text-2xs font-bold font-mono shrink-0">
            {getInitials(display)}
          </span>
          {expanded && (
            <>
              <span className="min-w-0 flex-1 text-left leading-tight">
                <span className="block text-sm font-medium text-text-primary truncate">
                  {display}
                </span>
                {email && email !== display && (
                  <span className="block text-2xs text-text-muted truncate mt-0.5">
                    {email}
                  </span>
                )}
              </span>
              <ChevronUpDownIcon className="w-4 h-4 text-text-muted shrink-0" />
            </>
          )}
        </button>
      </Dropdown.Trigger>

      <Dropdown.Panel aria-label="Account" className="w-56">
        <AccountMenuItems onDone={() => setOpen(false)} />
      </Dropdown.Panel>
    </Dropdown>
  );
}
