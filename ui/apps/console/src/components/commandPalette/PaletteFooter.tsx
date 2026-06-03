/** A single footer keyboard hint: one or more keycaps plus a label. */
function KeyHint({ keys, label }: { keys: string[]; label: string }) {
  return (
    <div className="flex items-center gap-1.5">
      {keys.map((k) => (
        <kbd
          key={k}
          className="px-1.5 py-0.5 text-2xs font-mono text-text-secondary bg-hover-medium border border-border rounded"
          aria-hidden="true"
        >
          {k}
        </kbd>
      ))}
      <span className="text-xs text-text-secondary">{label}</span>
    </div>
  );
}

/** Keyboard hints at the bottom of the palette. */
export default function PaletteFooter() {
  return (
    <div className="flex items-center gap-4 px-4 py-2.5 border-t border-border bg-card/30 shrink-0">
      <KeyHint keys={["↑", "↓"]} label="navigate" />
      <KeyHint keys={["↵"]} label="select" />
      <div className="ml-auto">
        <KeyHint keys={["⌘K"]} label="toggle" />
      </div>
    </div>
  );
}
