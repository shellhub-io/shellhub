import { useState } from "react";
import { LABEL } from "@/utils/styles";

/**
 * The enrollment budget maps to a single `usage_limit` int: 1 is single-use, N
 * (>=2) enrolls N devices, 0 is unlimited (reusable forever). This control is
 * one bar: a "Single-use" cap on the left, a hand-editable number with a stepper
 * in the middle (>=2), and an "Unlimited" cap on the right. Editing the middle
 * (typing or stepping) implies the "limited" mode.
 */

type Mode = "single" | "limited" | "unlimited";

function modeFromValue(value: number): Mode {
  if (value === 1) return "single";
  if (value === 0) return "unlimited";
  return "limited";
}

function helperFor(value: number): string {
  if (value === 0) return "Reusable. Registers any number of devices.";
  if (value === 1) return "Single-use. Spent after one device.";
  return `Registers up to ${value} devices.`;
}

const clamp = (n: number) => Math.max(2, n);

// Hide the native number spinners so the middle reads as plain text in the bar.
const NO_SPINNER =
  "[appearance:textfield] [&::-webkit-outer-spin-button]:appearance-none [&::-webkit-inner-spin-button]:appearance-none";

export default function UsageLimitField({
  value,
  onChange,
}: {
  value: number;
  onChange: (value: number) => void;
}) {
  const mode = modeFromValue(value);
  // The middle keeps its own string so multi-digit typing isn't clamped mid-edit;
  // it also remembers a count to restore when toggling back from single/unlimited.
  const [countStr, setCountStr] = useState(String(value >= 2 ? value : 5));

  // Resync the middle when the value changes from outside (edit prefill / reset),
  // using React's adjust-state-on-prop-change during render (no effect needed).
  const [prevValue, setPrevValue] = useState(value);
  if (value !== prevValue) {
    setPrevValue(value);
    if (value >= 2) setCountStr(String(value));
  }

  const goLimited = () => {
    const n = clamp(parseInt(countStr, 10) || 2);
    setCountStr(String(n));
    onChange(n);
  };

  const step = (delta: number) => {
    const n = clamp((parseInt(countStr, 10) || 2) + delta);
    setCountStr(String(n));
    onChange(n);
  };

  const handleInput = (raw: string) => {
    setCountStr(raw);
    const n = parseInt(raw, 10);
    if (!Number.isNaN(n) && n >= 2) onChange(n);
  };

  const handleBlur = () => {
    const n = parseInt(countStr, 10);
    if (Number.isNaN(n) || n < 2) {
      setCountStr("2");
      if (mode === "limited") onChange(2);
    }
  };

  const capBase =
    "px-4 flex items-center gap-1.5 text-xs font-medium transition-colors select-none cursor-pointer whitespace-nowrap";
  const capOn = "bg-primary/[0.13] text-primary";
  const capOff =
    "text-text-secondary hover:text-text-primary hover:bg-hover-medium";
  const stepBtn =
    "w-10 h-full grid place-items-center text-lg text-text-secondary hover:text-text-primary hover:bg-hover-medium transition-colors";

  return (
    <div>
      <span className={LABEL}>Usage limit</span>
      <div className="flex items-stretch h-11 bg-card border border-border rounded-lg overflow-hidden">
        <button
          type="button"
          onClick={() => onChange(1)}
          className={`${capBase} ${mode === "single" ? capOn : capOff}`}
        >
          <span className="text-xs leading-none">1&times;</span> Single-use
        </button>

        <div
          className={`flex items-center flex-1 min-w-0 border-x border-border transition-colors ${
            mode === "limited" ? "bg-primary/[0.07]" : ""
          }`}
        >
          {mode === "limited" ? (
            <>
              <button
                type="button"
                aria-label="Decrease"
                onClick={() => step(-1)}
                className={stepBtn}
              >
                −
              </button>
              <input
                type="number"
                min={2}
                aria-label="Number of devices"
                value={countStr}
                onChange={(e) => handleInput(e.target.value)}
                onBlur={handleBlur}
                className={`flex-1 min-w-0 text-center bg-transparent font-mono text-sm font-semibold text-text-primary outline-none ${NO_SPINNER}`}
              />
              <button
                type="button"
                aria-label="Increase"
                onClick={() => step(1)}
                className={stepBtn}
              >
                +
              </button>
            </>
          ) : (
            <button
              type="button"
              aria-label="Set a custom device limit"
              onClick={goLimited}
              className="flex-1 h-full grid place-items-center text-sm font-medium text-text-muted hover:text-text-primary transition-colors"
            >
              Custom
            </button>
          )}
        </div>

        <button
          type="button"
          onClick={() => onChange(0)}
          className={`${capBase} ${mode === "unlimited" ? capOn : capOff}`}
        >
          <span className="text-[15px] leading-none">&#8734;</span> Unlimited
        </button>
      </div>
      <p className="text-2xs text-text-muted mt-2">{helperFor(value)}</p>
    </div>
  );
}
