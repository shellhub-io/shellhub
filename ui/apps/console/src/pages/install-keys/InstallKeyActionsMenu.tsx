import {
  useCallback,
  useEffect,
  useRef,
  useState,
  type ReactNode,
} from "react";
import { createPortal } from "react-dom";
import {
  EllipsisVerticalIcon,
  NoSymbolIcon,
  PauseIcon,
  PencilIcon,
  PlayIcon,
} from "@heroicons/react/24/outline";
import { IconButton } from "@shellhub/design-system/primitives";
import { type InstallKey } from "@/client";
import { useEscapeKey } from "@/hooks/useEscapeKey";
import RestrictedAction from "@/components/common/RestrictedAction";
import { type Action } from "@/utils/permission";
import { isPairingKey, isSystemKey } from "./helpers";

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
 * Trailing overflow menu for an install key row. An ellipsis trigger toggles a menu closed on Escape
 * or an outside click. The menu renders in a portal (see the position effect) so it escapes the
 * table's horizontal-scroll clip; there's no shared dropdown primitive to lean on.
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
  const triggerRef = useRef<HTMLDivElement>(null);
  const menuRef = useRef<HTMLDivElement>(null);
  // Right-aligned position, in viewport coordinates. The menu renders in a portal on document.body so
  // it escapes the table's `overflow-x-auto` clip — otherwise opening it extends the scroll width and
  // shifts the whole table (and the menu is clipped out of view).
  const [pos, setPos] = useState({ top: 0, right: 0 });

  const updatePosition = useCallback(() => {
    if (!triggerRef.current) return;
    const rect = triggerRef.current.getBoundingClientRect();
    setPos({ top: rect.bottom + 4, right: window.innerWidth - rect.right });
  }, []);

  useEscapeKey(() => setOpen(false), open);

  useEffect(() => {
    if (!open) return undefined;
    updatePosition();
    // Move focus into the menu when it opens so it can be driven from the keyboard.
    menuRef.current
      ?.querySelector<HTMLButtonElement>('[role="menuitem"]')
      ?.focus();

    const reposition = () => updatePosition();
    window.addEventListener("scroll", reposition, true);
    window.addEventListener("resize", reposition);
    const onDown = (e: MouseEvent) => {
      if (
        !menuRef.current?.contains(e.target as Node) &&
        !triggerRef.current?.contains(e.target as Node)
      )
        setOpen(false);
    };
    document.addEventListener("mousedown", onDown);
    return () => {
      window.removeEventListener("scroll", reposition, true);
      window.removeEventListener("resize", reposition);
      document.removeEventListener("mousedown", onDown);
    };
  }, [open, updatePosition]);

  // The pairing key force-accepts and has no editable fields, so it has no actions at all.
  if (isPairingKey(installKey)) return null;

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

  // A revoked key is terminal, so it has no actions.
  if (installKey.revoked) {
    return null;
  }

  return (
    <div ref={triggerRef} className="shrink-0">
      <IconButton
        variant="ghost"
        aria-label="Install Key actions"
        aria-haspopup="menu"
        aria-expanded={open}
        onClick={() => setOpen((v) => !v)}
      >
        <EllipsisVerticalIcon className="w-4 h-4" />
      </IconButton>

      {open &&
        createPortal(
          <div
            ref={menuRef}
            role="menu"
            tabIndex={-1}
            onKeyDown={onMenuKeyDown}
            style={{ top: pos.top, right: pos.right }}
            className="fixed z-dropdown w-40 py-1 bg-surface border border-border rounded-lg shadow-2xl animate-fade-in"
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
            {!isSystemKey(installKey) && (
              <MenuItem
                action="installKey:revoke"
                icon={<NoSymbolIcon className="w-4 h-4" />}
                label="Revoke"
                danger
                onSelect={run(() => onRevoke(installKey))}
              />
            )}
          </div>,
          document.body,
        )}
    </div>
  );
}
