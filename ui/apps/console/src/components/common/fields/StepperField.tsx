import { ComponentProps, useState } from "react";
import { cn } from "@shellhub/design-system/cn";
import { IconButton } from "@shellhub/design-system/primitives";
import NumericInput from "@/components/common/fields/NumericInput";

type StepperFieldProps = Pick<
  ComponentProps<typeof NumericInput>,
  "id" | "label" | "readOnly" | "className"
> & {
  value: number;
  onChange: (value: number) => void;
  min: number;
  max?: number;
  size?: "sm" | "md";
};

const SIZE = {
  sm: { bar: "h-9", btn: "w-9 text-base" },
  md: { bar: "h-11", btn: "w-10 text-lg" },
} as const;

/**
 * A number field with increment and decrement buttons, clamped to its range.
 */
export default function StepperField({
  value,
  onChange,
  min,
  max,
  size = "md",
  className,
  ...rest
}: StepperFieldProps) {
  const { readOnly } = rest;

  const [displayValue, setDisplayValue] = useState(String(value));

  const [prevValue, setPrevValue] = useState(value);
  if (value !== prevValue) {
    setPrevValue(value);
    setDisplayValue(String(value));
  }

  const step = (delta: number) => {
    const next = value + delta;
    setDisplayValue(String(next));
    onChange(next);
  };

  const handleInput = (raw: string) => {
    setDisplayValue(raw);
    const n = parseInt(raw, 10);
    if (n >= min && (max === undefined || n <= max)) {
      onChange(n);
    }
  };

  const handleBlur = () => {
    if (!displayValue) {
      setDisplayValue(String(min));
      onChange(min);
      return;
    }
    setDisplayValue(String(value));
  };

  const atMin = value <= min;
  const atMax = max !== undefined && value >= max;
  const s = SIZE[size];

  const btnClass = cn(s.btn, "h-full rounded-none p-0");

  return (
    <div
      className={cn(
        "inline-flex items-center overflow-hidden",
        s.bar,
        "bg-card border border-border rounded-lg",
        className,
      )}
    >
      <IconButton
        variant="ghost"
        aria-label="Decrease"
        disabled={atMin}
        onClick={() => step(-1)}
        className={btnClass}
      >
        −
      </IconButton>
      <NumericInput
        {...rest}
        hideLabel
        value={readOnly ? String(value) : displayValue}
        onChange={handleInput}
        onBlur={readOnly ? undefined : handleBlur}
        className="border-0 bg-transparent rounded-none px-0 py-0 text-center font-mono text-sm font-semibold text-text-primary focus:ring-0 focus:border-transparent"
      />
      <IconButton
        variant="ghost"
        aria-label="Increase"
        disabled={atMax}
        onClick={() => step(1)}
        className={btnClass}
      >
        +
      </IconButton>
    </div>
  );
}
