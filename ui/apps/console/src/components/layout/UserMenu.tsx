import { useState } from "react";
import { ChevronDownIcon } from "@heroicons/react/24/outline";
import { Dropdown } from "@shellhub/design-system/primitives";
import { cn } from "@shellhub/design-system/cn";
import { accountDisplayName, useAuthStore } from "@/stores/authStore";
import { getInitials } from "@/utils/string";
import AccountMenuItems from "./AccountMenuItems";

/**
 * The account menu: who is signed in, the way to settings, and sign-out.
 */
export default function UserMenu() {
  const { user, name, email } = useAuthStore();
  const [open, setOpen] = useState(false);

  const display = accountDisplayName({ user, name, email });

  if (!user && !name && !email) return null;

  return (
    <Dropdown mode="content" placement="bottom-end" open={open} onOpenChange={setOpen}>
      <Dropdown.Trigger>
        <button
          type="button"
          aria-label={`Account menu for ${display}`}
          className="flex items-center gap-2 h-8 pl-1 pr-2.5 rounded-lg border border-transparent hover:border-border hover:bg-hover-subtle transition-all duration-150"
        >
          <span className="w-6 h-6 rounded-md bg-primary/15 border border-primary/20 flex items-center justify-center text-primary text-2xs font-bold font-mono">
            {getInitials(display)}
          </span>
          <span className="hidden sm:inline text-xs font-medium text-text-secondary max-w-[120px] truncate">
            {display}
          </span>
          <ChevronDownIcon
            className={cn(
            "w-3 h-3 text-text-muted transition-transform duration-200",
            open && "rotate-180",
          )}
            strokeWidth={2.5}
          />
        </button>
      </Dropdown.Trigger>

      <Dropdown.Panel className="w-56">
        <div className="p-3.5 border-b border-border">
          <div className="flex items-center gap-3">
            <span className="w-9 h-9 rounded-lg bg-primary/15 border border-primary/20 flex items-center justify-center text-primary text-xs font-bold font-mono shrink-0">
              {getInitials(display)}
            </span>
            <div className="min-w-0">
              <p className="text-sm font-semibold text-text-primary truncate">
                {display}
              </p>
              <p className="text-2xs text-text-muted mt-0.5">Logged in</p>
            </div>
          </div>
        </div>

        <AccountMenuItems onDone={() => setOpen(false)} />
      </Dropdown.Panel>
    </Dropdown>
  );
}
