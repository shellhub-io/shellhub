import { ChevronLeftIcon } from "@heroicons/react/24/outline";
import { IconButton } from "@shellhub/design-system/primitives";
import { cn } from "@shellhub/design-system/cn";
import type { NormalizedDevice } from "@/hooks/useDevices";
import { LISTBOX_ID, icons } from "./items";

interface PaletteHeaderProps {
  inputRef: React.RefObject<HTMLInputElement | null>;
  query: string;
  drillDevice: NormalizedDevice | null;
  commandMode: boolean;
  hasResults: boolean;
  activeOptionId: string | undefined;
  onQueryChange: (value: string) => void;
  onKeyDown: (e: React.KeyboardEvent<HTMLInputElement>) => void;
  onBack: () => void;
}

/** Mode-aware header: a back-button + device breadcrumb while drilled in, the
 *  search/command glyph otherwise, plus the combobox input. */
export default function PaletteHeader({
  inputRef,
  query,
  drillDevice,
  commandMode,
  hasResults,
  activeOptionId,
  onQueryChange,
  onKeyDown,
  onBack,
}: PaletteHeaderProps) {
  return (
    <div className="flex items-center gap-3 px-4 border-b border-border shrink-0">
      {drillDevice ? (
        <>
          <IconButton
            size="sm"
            onClick={onBack}
            aria-label="Back to devices"
            className="-ml-1"
          >
            <ChevronLeftIcon className="w-5 h-5" />
          </IconButton>
          <div className="shrink-0 flex items-center gap-1.5 min-w-0">
            <span className="text-text-muted" aria-hidden="true">
              {icons.devices}
            </span>
            <span className="text-sm font-medium text-text-primary truncate max-w-[10rem]">
              {drillDevice.name}
            </span>
          </div>
        </>
      ) : (
        <>
          <span
            className={cn("shrink-0", commandMode ? "text-primary" : "text-text-muted")}
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
        </>
      )}
      <input
        ref={inputRef}
        type="text"
        role="combobox"
        aria-label={
          drillDevice
            ? `Search actions for ${drillDevice.name}`
            : commandMode
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
          drillDevice
            ? "Search actions…"
            : commandMode
              ? "Search commands…"
              : "Search devices to connect…"
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
