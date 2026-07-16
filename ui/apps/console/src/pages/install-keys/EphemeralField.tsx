import CheckboxField from "@/components/common/fields/CheckboxField";

const EPHEMERAL_MIN = 1;
const EPHEMERAL_MAX = 10;

/**
 * Ephemeral lifecycle for a key's devices: when on, an enrolled device is removed
 * after it stays offline past the timeout (1-10 min). A selectable card that
 * lights up when enabled and reveals the timeout stepper. The on/off stays a
 * plain checkbox so it inherits the design-system field when that lands.
 */
export default function EphemeralField({
  id,
  enabled,
  onEnabledChange,
  timeout,
  onTimeoutChange,
}: {
  // Unique per drawer: Create and Edit are both mounted at once, so a shared id would make one
  // checkbox's label toggle the other's (inert) input.
  id: string;
  enabled: boolean;
  onEnabledChange: (value: boolean) => void;
  timeout: number;
  onTimeoutChange: (value: number) => void;
}) {
  const step = (delta: number) =>
    onTimeoutChange(
      Math.min(EPHEMERAL_MAX, Math.max(EPHEMERAL_MIN, timeout + delta)),
    );

  const stepBtn =
    "w-9 h-full grid place-items-center text-base text-text-secondary hover:text-text-primary hover:bg-hover-medium transition-colors";

  return (
    <div
      className={`rounded-xl border p-4 transition-colors ${
        enabled
          ? "border-primary/40 bg-primary/[0.05]"
          : "border-border bg-transparent"
      }`}
    >
      <CheckboxField
        id={id}
        label="Ephemeral devices"
        description="Automatically remove registered devices once they stay offline past the timeout."
        checked={enabled}
        onChange={onEnabledChange}
      />
      {enabled && (
        <div className="flex items-center gap-3 mt-4 pt-4 border-t border-border/70">
          <span className="text-xs text-text-secondary">Remove after</span>
          <div className="inline-flex items-center h-9 bg-card border border-border rounded-lg overflow-hidden">
            <button
              type="button"
              aria-label="Decrease"
              onClick={() => step(-1)}
              className={stepBtn}
            >
              −
            </button>
            <span className="min-w-[2.25rem] text-center font-mono text-sm font-semibold text-text-primary">
              {timeout}
            </span>
            <button
              type="button"
              aria-label="Increase"
              onClick={() => step(1)}
              className={stepBtn}
            >
              +
            </button>
          </div>
          <span className="text-2xs text-text-muted">min offline</span>
        </div>
      )}
    </div>
  );
}
