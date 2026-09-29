import { useState } from "react";
import {
  ChevronDownIcon,
  ChevronUpDownIcon,
} from "@heroicons/react/24/outline";
import { Dropdown } from "@shellhub/design-system/primitives";
import { cn } from "@shellhub/design-system/cn";
import { accountDisplayName, useAuthStore } from "@/stores/authStore";
import { getInitials } from "@/utils/string";
import AccountMenuItems from "./AccountMenuItems";
import SupportPaywallDialog from "./SupportPaywallDialog";

/**
 * The signed-in user and the account actions behind it. placement says where it sits: at the foot
 * of the sidebar, with the name when expanded or the avatar alone on the rail, opening upward; or
 * beside the tabs whenever the sidebar cannot show it, opening down, as a compact avatar and name
 * (tabStrip) or the round avatar alone (tabStripAvatar), which the caller picks by window width.
 */
export default function SessionMenu({
  placement,
}: {
  placement: "expanded" | "rail" | "tabStrip" | "tabStripAvatar";
}) {
  const expanded = placement === "expanded";
  const named = placement === "tabStrip";
  const inTabStrip = named || placement === "tabStripAvatar";
  const email = useAuthStore((s) => s.email);
  const [open, setOpen] = useState(false);
  const [paywallOpen, setPaywallOpen] = useState(false);

  const display = useAuthStore(accountDisplayName);

  return (
    <>
      <Dropdown
        mode="content"
        placement={inTabStrip ? "bottom-end" : "top-start"}
        portal
        open={open}
        onOpenChange={setOpen}
      >
        <Dropdown.Trigger>
          <button
            type="button"
            data-testid="session-menu"
            aria-label={`Account menu for ${display}`}
            title={expanded || named ? undefined : display}
            className={cn(
              "flex items-center transition-colors duration-150",
              inTabStrip && !named
                ? "rounded-full p-0.5 hover:bg-hover-subtle"
                : "rounded-lg border border-border bg-card hover:border-border-light",
              named && "gap-2 p-1 pr-2.5 max-w-48",
              !inTabStrip &&
                (expanded
                  ? "w-full gap-2.5 p-2"
                  : "w-full justify-center p-1.5"),
            )}
          >
            <span
              className={cn(
                "rounded-full bg-primary/15 border border-primary/20 flex items-center justify-center text-primary text-2xs font-bold font-mono shrink-0",
                named ? "w-5 h-5" : inTabStrip ? "w-6 h-6" : "w-7 h-7",
              )}
            >
              {getInitials(display)}
            </span>
            {named && (
              <>
                <span className="min-w-0 truncate text-xs font-medium text-text-primary">
                  {display}
                </span>
                <ChevronDownIcon className="w-3.5 h-3.5 text-text-muted shrink-0" />
              </>
            )}
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
          {!expanded && (
            <div className="px-4 pt-3 pb-2.5 border-b border-border leading-tight">
              <p className="text-sm font-medium text-text-primary truncate">
                {display}
              </p>
              {email && email !== display && (
                <p className="text-2xs text-text-muted truncate mt-0.5">
                  {email}
                </p>
              )}
            </div>
          )}
          <AccountMenuItems
            onDone={() => setOpen(false)}
            onHelpNeedsPlan={() => setPaywallOpen(true)}
          />
        </Dropdown.Panel>
      </Dropdown>
      <SupportPaywallDialog
        open={paywallOpen}
        onClose={() => setPaywallOpen(false)}
      />
    </>
  );
}
