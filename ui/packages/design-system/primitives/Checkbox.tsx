import type { InputHTMLAttributes } from "react";
import { CheckIcon } from "@heroicons/react/16/solid";
import { cn } from "./cn";

export type CheckboxProps = {
  checked: boolean;
  onChange: (checked: boolean) => void;
  disabled?: boolean;
  id?: string;
  className?: string;
} & Omit<
  InputHTMLAttributes<HTMLInputElement>,
  "checked" | "onChange" | "disabled" | "id" | "className" | "type"
>;

const BOX_BASE =
  "flex items-center justify-center w-4 h-4 rounded border-2 transition-colors";

export function Checkbox({
  checked,
  onChange,
  disabled,
  id,
  className,
  ...rest
}: CheckboxProps) {
  return (
    <span
      className={cn(
        "inline-flex focus-within:ring-2 focus-within:ring-primary/40",
        disabled && "opacity-dim cursor-not-allowed",
        className,
      )}
    >
      <input
        {...rest}
        id={id}
        type="checkbox"
        checked={checked}
        disabled={disabled}
        onChange={(e) => onChange(e.target.checked)}
        className="sr-only"
      />
      <span
        aria-hidden="true"
        className={cn(
          BOX_BASE,
          checked
            ? "bg-primary border-primary"
            : "border-border bg-card hover:border-border-light",
        )}
      >
        {checked && <CheckIcon className="w-3 h-3 text-white" />}
      </span>
    </span>
  );
}
