import { MagnifyingGlassIcon } from "@heroicons/react/24/outline";
import { useCommandPaletteStore } from "@/stores/commandPaletteStore";
import { INPUT } from "@/utils/styles";
import { navIcon } from "./SidebarShell";

/**
 * Visible affordance for the command palette, rendered in the Sidebar header.
 * Expanded: an input-style field (reuses the app's `INPUT` style so it matches
 * real inputs). Collapsed: a centered icon button matching nav-item sizing.
 * Clicking opens the palette; the Cmd/Ctrl+K shortcut shares the same store.
 *
 * `onActivate` runs alongside opening — used by the mobile drawer to close
 * itself so the palette isn't stacked on top of an open drawer.
 */
export default function CommandPaletteTrigger({
  expanded,
  onActivate,
}: {
  expanded: boolean;
  onActivate?: () => void;
}) {
  const openPalette = useCommandPaletteStore((s) => s.openPalette);

  const shared = {
    type: "button" as const,
    onClick: () => {
      openPalette();
      onActivate?.();
    },
    "aria-label": "Open command palette",
    "aria-haspopup": "dialog" as const,
    "aria-keyshortcuts": "Meta+K Control+K",
  };

  if (!expanded) {
    return (
      <button
        {...shared}
        title="Quick connect (⌘K)"
        className="w-full flex items-center justify-center px-3 py-2 rounded-md border border-transparent text-text-secondary hover:text-text-primary hover:bg-hover-subtle transition-all duration-150"
      >
        <MagnifyingGlassIcon
          className={navIcon}
          aria-hidden="true"
          strokeWidth={2}
        />
      </button>
    );
  }

  return (
    <button
      {...shared}
      className={`${INPUT} group flex items-center gap-2.5 text-left hover:border-primary/40`}
    >
      <MagnifyingGlassIcon
        className="w-4 h-4 text-text-muted shrink-0"
        aria-hidden="true"
        strokeWidth={2}
      />
      <span className="flex-1 truncate text-text-secondary group-hover:text-text-primary transition-colors">
        Quick connect…
      </span>
      <kbd
        className="shrink-0 px-1.5 py-0.5 text-2xs font-mono text-text-secondary bg-hover-medium border border-border rounded"
        aria-hidden="true"
      >
        ⌘K
      </kbd>
    </button>
  );
}
