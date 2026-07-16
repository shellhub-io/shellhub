import { useEffect, useRef, useState, type ReactNode } from "react";
import {
  EllipsisVerticalIcon,
  NoSymbolIcon,
  PauseIcon,
  PencilIcon,
  PlayIcon,
} from "@heroicons/react/24/outline";
import { IconButton } from "@shellhub/design-system/primitives";
import { type InstallKey } from "@/client";
import { useClickOutside } from "@/hooks/useClickOutside";
import { useEscapeKey } from "@/hooks/useEscapeKey";
import RestrictedAction from "@/components/common/RestrictedAction";
import { type Action } from "@/utils/permission";

function MenuItem({
  action,
  icon,
  label,
  danger,
  onSelect,
}: {
  action: Action;
  icon: ReactNode;
  label: string;
  danger?: boolean;
  onSelect: () => void;
}) {
  return (
    <RestrictedAction action={action}>
      <button
        type="button"
        role="menuitem"
        tabIndex={-1}
        onClick={onSelect}
        className={`flex w-full items-center gap-2.5 px-3 py-2 text-left text-xs transition-colors ${
          danger
            ? "text-accent-red hover:bg-accent-red/10"
            : "text-text-secondary hover:bg-hover-subtle hover:text-text-primary"
        }`}
      >
        <span className="shrink-0">{icon}</span>
        {label}
      </button>
    </RestrictedAction>
  );
}

/**
 * Trailing overflow menu for a install key row. Built on the codebase's
 * click-outside popover pattern (no shared dropdown primitive exists): an
 * ellipsis trigger toggles an absolutely-positioned menu, closed on Escape or
 * an outside click.
 */
export default function InstallKeyActionsMenu({
  installKey,
  onEdit,
  onToggleDisabled,
  onRevoke,
}: {
  installKey: InstallKey;
  onEdit: (key: InstallKey) => void;
  onToggleDisabled: (key: InstallKey) => void;
  onRevoke: (key: InstallKey) => void;
}) {
  const [open, setOpen] = useState(false);
  const wrapperRef = useRef<HTMLDivElement>(null);
  const menuRef = useRef<HTMLDivElement>(null);

  useClickOutside(wrapperRef, () => setOpen(false));
  useEscapeKey(() => setOpen(false), open);

  // Move focus into the menu when it opens so it can be driven from the keyboard.
  useEffect(() => {
    if (open) {
      menuRef.current
        ?.querySelector<HTMLButtonElement>('[role="menuitem"]')
        ?.focus();
    }
  }, [open]);

  // Arrow keys (and Home/End) roam the rendered menu items, wrapping around. Queried live because
  // items are conditionally shown (permissions, the system key hides Revoke).
  const onMenuKeyDown = (event: React.KeyboardEvent) => {
    const items = Array.from(
      menuRef.current?.querySelectorAll<HTMLButtonElement>(
        '[role="menuitem"]',
      ) ?? [],
    );
    if (items.length === 0) return;

    const current = items.indexOf(document.activeElement as HTMLButtonElement);
    let next: number;
    switch (event.key) {
      case "ArrowDown":
        next = (current + 1) % items.length;
        break;
      case "ArrowUp":
        next = (current - 1 + items.length) % items.length;
        break;
      case "Home":
        next = 0;
        break;
      case "End":
        next = items.length - 1;
        break;
      default:
        return;
    }

    event.preventDefault();
    items[next]?.focus();
  };

  const run = (fn: () => void) => () => {
    setOpen(false);
    fn();
  };

  // A revoked key is terminal — no actions. The legacy/system key is auto-managed but still editable
  // (its mode governs keyless enrollments), so it keeps Edit; Revoke stays hidden — it's permanent.
  if (installKey.revoked) {
    return null;
  }

  return (
    <div ref={wrapperRef} className="relative shrink-0">
      <IconButton
        variant="ghost"
        aria-label="Install Key actions"
        aria-haspopup="menu"
        aria-expanded={open}
        onClick={() => setOpen((v) => !v)}
      >
        <EllipsisVerticalIcon className="w-4 h-4" />
      </IconButton>

      {open && (
        <div
          ref={menuRef}
          role="menu"
          tabIndex={-1}
          onKeyDown={onMenuKeyDown}
          className="absolute right-0 top-full mt-1 z-50 w-40 py-1 bg-surface border border-border rounded-lg shadow-2xl animate-fade-in"
        >
          <MenuItem
            action="installKey:edit"
            icon={<PencilIcon className="w-4 h-4" />}
            label="Edit"
            onSelect={run(() => onEdit(installKey))}
          />
          {/* Disable is available for the legacy key too: disabling it turns off keyless enrollment
              (devices without a key are rejected). Revoke stays hidden — the legacy key is permanent. */}
          <MenuItem
            action="installKey:disable"
            icon={
              installKey.disabled ? (
                <PlayIcon className="w-4 h-4" />
              ) : (
                <PauseIcon className="w-4 h-4" />
              )
            }
            label={installKey.disabled ? "Enable" : "Disable"}
            onSelect={run(() => onToggleDisabled(installKey))}
          />
          {!installKey.system && (
            <MenuItem
              action="installKey:revoke"
              icon={<NoSymbolIcon className="w-4 h-4" />}
              label="Revoke"
              danger
              onSelect={run(() => onRevoke(installKey))}
            />
          )}
        </div>
      )}
    </div>
  );
}
