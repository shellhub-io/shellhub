import { Spinner } from "@shellhub/design-system/primitives";

/**
 * A spinner over a line of text, for a dialog step that is waiting on the server. The label is
 * announced politely each time it changes.
 */
export default function StatusMessage({ label }: { label: string }) {
  return (
    <div
      className="flex flex-col items-center gap-3 py-6"
      role="status"
      aria-live="polite"
    >
      <Spinner size="2xl" />
      <p className="text-sm text-text-muted">{label}</p>
    </div>
  );
}
