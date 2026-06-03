import BaseDialog from "@/components/common/BaseDialog";
import { useCommandPalette } from "@/hooks/useCommandPalette";
import { LISTBOX_ID, optionId } from "./items";
import PaletteHeader from "./PaletteHeader";
import FeedbackBanner from "./FeedbackBanner";
import CommandRow from "./CommandRow";
import PaletteFooter from "./PaletteFooter";

/**
 * Cmd/Ctrl+K command palette. A thin presentational shell over
 * `useCommandPalette()` — connection-first by default (connect/restore a
 * device), with a ">"-prefixed command mode for page navigation.
 */
export default function CommandPalette() {
  const {
    open,
    listRef,
    query,
    commandMode,
    sections,
    hasResults,
    indexById,
    safeIndex,
    activeItem,
    rejectMessage,
    shakeId,
    onQueryChange,
    setActiveIndex,
    handleKeyDown,
    close,
  } = useCommandPalette();

  if (!open) return null;

  return (
    <BaseDialog
      open={open}
      onClose={close}
      size="xl"
      aria-label="Command palette"
      className="overflow-hidden sm:max-h-[85vh]"
    >
      <PaletteHeader
        query={query}
        commandMode={commandMode}
        hasResults={hasResults}
        activeOptionId={activeItem ? optionId(activeItem.id) : undefined}
        onQueryChange={onQueryChange}
        onKeyDown={handleKeyDown}
      />

      <FeedbackBanner key={rejectMessage} message={rejectMessage} />

      <div
        ref={listRef}
        className="flex-1 min-h-0 overflow-y-auto overscroll-contain"
      >
        {!hasResults ? (
          <div className="px-4 py-10 text-center" role="status">
            <p className="text-sm text-text-muted">
              {commandMode ? "No commands match" : "No devices match"}
            </p>
            <p className="text-2xs text-text-muted/50 mt-1">
              {commandMode ? "Try a different command" : "Type > for commands"}
            </p>
          </div>
        ) : (
          <div id={LISTBOX_ID} role="listbox" aria-label="Results">
            {Array.from(sections.entries()).map(([section, sectionItems]) => (
              <div key={section}>
                <div className="px-4 pt-3 pb-1.5">
                  <p className="text-2xs font-mono font-semibold uppercase tracking-label text-text-muted/50">
                    {section}
                  </p>
                </div>
                {sectionItems.map((item) => {
                  const idx = indexById.get(item.id) ?? -1;
                  return (
                    <CommandRow
                      key={item.id}
                      item={item}
                      isActive={idx === safeIndex}
                      shaking={shakeId === item.id}
                      onActivate={() => setActiveIndex(idx)}
                    />
                  );
                })}
              </div>
            ))}
          </div>
        )}
      </div>

      <PaletteFooter />
    </BaseDialog>
  );
}
