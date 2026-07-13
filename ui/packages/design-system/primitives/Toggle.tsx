import type { ButtonHTMLAttributes } from "react";
import { cn } from "./cn";

export type ToggleProps = {
  enabled: boolean;
  onChange: (enabled: boolean) => void;
  disabled?: boolean;
  className?: string;
} & Omit<
  ButtonHTMLAttributes<HTMLButtonElement>,
  "onChange" | "disabled" | "className" | "type" | "role"
>;

const TRACK_BASE =
  "relative inline-flex h-5 w-9 items-center rounded-full transition-colors " +
  "focus:ring-2 focus:ring-primary/30 focus:ring-offset-2 focus:ring-offset-card";
const THUMB_BASE =
  "inline-block h-3.5 w-3.5 rounded-full bg-white shadow-sm transition-transform";

export function Toggle({
  enabled,
  onChange,
  disabled,
  className,
  ...rest
}: ToggleProps) {
  return (
    <button
      {...rest}
      type="button"
      role="switch"
      aria-checked={enabled}
      disabled={disabled}
      onClick={() => onChange(!enabled)}
      className={cn(
        TRACK_BASE,
        enabled ? "bg-primary" : "bg-border",
        disabled && "opacity-dim cursor-not-allowed",
        className,
      )}
    >
      <span
        aria-hidden="true"
        className={cn(
          THUMB_BASE,
          enabled ? "translate-x-[18px]" : "translate-x-0.5",
        )}
      />
    </button>
  );
}
