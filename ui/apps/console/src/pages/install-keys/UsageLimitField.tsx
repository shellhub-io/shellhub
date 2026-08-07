import { useState } from "react";
import StepperField from "@/components/common/fields/StepperField";
import { LABEL } from "@/utils/styles";

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

export default function UsageLimitField({
  value,
  onChange,
}: {
  value: number;
  onChange: (value: number) => void;
}) {
  const mode = modeFromValue(value);
  const [lastCount, setLastCount] = useState(value >= 2 ? value : 5);

  const [prevValue, setPrevValue] = useState(value);
  if (value !== prevValue) {
    setPrevValue(value);
    if (value >= 2) setLastCount(value);
  }

  const goLimited = () => {
    const n = Math.max(2, lastCount);
    onChange(n);
  };

  const handleStepperChange = (n: number) => {
    setLastCount(n);
    onChange(n);
  };

  const capBase =
    "px-4 flex items-center gap-1.5 text-xs font-medium transition-colors select-none cursor-pointer whitespace-nowrap";
  const capOn = "bg-primary/[0.13] text-primary";
  const capOff =
    "text-text-secondary hover:text-text-primary hover:bg-hover-medium";

  return (
    <div>
      <span className={LABEL}>Usage limit</span>
      <div className="flex @sm/drawer:flex-row flex-col bg-card border border-border rounded-lg overflow-hidden">
        <button
          type="button"
          onClick={() => onChange(1)}
          className={`${capBase} h-11 justify-center ${mode === "single" ? capOn : capOff}`}
        >
          <span className="text-xs leading-none">1&times;</span> Single-use
        </button>

        <div
          className={`flex items-center @sm/drawer:flex-1 min-w-0 h-11 @sm/drawer:border-x border-y @sm/drawer:border-y-0 border-border transition-colors ${
            mode === "limited" ? "bg-primary/[0.07]" : ""
          }`}
        >
          {mode === "limited" ? (
            <StepperField
              id="usage-limit-stepper"
              label="Number of devices"
              value={value}
              onChange={handleStepperChange}
              min={2}
              className="border-0 bg-transparent rounded-none w-full h-full [&>div]:flex-1 [&>div]:min-w-0"
            />
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
          className={`${capBase} h-11 justify-center ${mode === "unlimited" ? capOn : capOff}`}
        >
          <span className="text-[15px] leading-none">&#8734;</span> Unlimited
        </button>
      </div>
      <p className="text-2xs text-text-muted mt-2">{helperFor(value)}</p>
    </div>
  );
}
