import { icons, LISTBOX_ID } from "./items";

interface PaletteHeaderProps {
  query: string;
  hasResults: boolean;
  activeOptionId: string | undefined;
  onQueryChange: (value: string) => void;
  onKeyDown: (e: React.KeyboardEvent) => void;
}

/** Search header: the leading glyph plus the combobox input. */
export default function PaletteHeader({
  query,
  hasResults,
  activeOptionId,
  onQueryChange,
  onKeyDown,
}: PaletteHeaderProps) {
  return (
    <div className="flex items-center gap-3 px-4 border-b border-border shrink-0">
      <span className="text-text-muted shrink-0" aria-hidden="true">
        {icons.search}
      </span>
      <input
        type="text"
        role="combobox"
        aria-label="Search pages, devices, and sessions"
        aria-expanded
        aria-controls={hasResults ? LISTBOX_ID : undefined}
        aria-activedescendant={activeOptionId}
        aria-autocomplete="list"
        value={query}
        onChange={(e) => onQueryChange(e.target.value)}
        onKeyDown={onKeyDown}
        placeholder="Search pages, devices, sessions..."
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
