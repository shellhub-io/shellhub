import { icons, LISTBOX_ID } from "./items";

interface PaletteHeaderProps {
  query: string;
  commandMode: boolean;
  hasResults: boolean;
  activeOptionId: string | undefined;
  onQueryChange: (value: string) => void;
  onKeyDown: (e: React.KeyboardEvent) => void;
}

/** Mode-aware header: the search/command glyph (plus a "Commands" badge in
 *  command mode) and the combobox input. */
export default function PaletteHeader({
  query,
  commandMode,
  hasResults,
  activeOptionId,
  onQueryChange,
  onKeyDown,
}: PaletteHeaderProps) {
  return (
    <div className="flex items-center gap-3 px-4 border-b border-border shrink-0">
      <span
        className={`shrink-0 ${commandMode ? "text-primary" : "text-text-muted"}`}
        aria-hidden="true"
      >
        {commandMode ? icons.command : icons.search}
      </span>
      {commandMode && (
        <span
          className="shrink-0 px-1.5 py-0.5 text-2xs font-mono font-semibold uppercase tracking-label text-primary bg-primary/10 border border-primary/20 rounded"
          aria-hidden="true"
        >
          Commands
        </span>
      )}
      <input
        type="text"
        role="combobox"
        aria-label={
          commandMode
            ? "Search commands"
            : "Search devices to connect, or type > for commands"
        }
        aria-expanded
        aria-controls={hasResults ? LISTBOX_ID : undefined}
        aria-activedescendant={activeOptionId}
        aria-autocomplete="list"
        value={query}
        onChange={(e) => onQueryChange(e.target.value)}
        onKeyDown={onKeyDown}
        placeholder={
          commandMode ? "Search commands…" : "Search devices to connect…"
        }
        className="flex-1 h-12 bg-transparent text-sm text-text-primary placeholder:text-text-secondary focus:outline-none"
      />
      <kbd
        className="shrink-0 px-1.5 py-0.5 text-2xs font-mono font-semibold text-text-muted/50 bg-hover-medium border border-border rounded"
        aria-hidden="true"
      >
        ESC
      </kbd>
    </div>
  );
}
