import { InputHTMLAttributes, ReactNode } from "react";
import { Checkbox } from "@shellhub/design-system/primitives";
import { cn } from "@shellhub/design-system/cn";
import FieldError from "@/components/common/fields/FieldError";
import FieldHint from "@/components/common/fields/FieldHint";

type CheckboxProps = Omit<
  InputHTMLAttributes<HTMLInputElement>,
  "id" | "type" | "checked" | "onChange"
>;

type Props = CheckboxProps & {
  id: string;
  label: ReactNode;
  /** Visually hide the label (kept for screen readers via `sr-only`). */
  hideLabel?: boolean;
  /** Optional secondary text rendered below the label. */
  description?: ReactNode;
  checked: boolean;
  onChange: (checked: boolean) => void;
  hint?: string;
  error?: string;
};

export default function CheckboxField({
  id,
  label,
  hideLabel = false,
  description,
  checked,
  onChange,
  hint,
  error,
  required,
  title,
  ...rest
}: Props) {
  const errorId = `${id}-error`;
  const hintId = `${id}-hint`;
  const describedBy = error ? errorId : hint ? hintId : undefined;
  const alignment = description ? "items-start" : "items-center";

  return (
    <div>
      <label htmlFor={id} title={title} className={cn("flex gap-2.5", alignment, rest.disabled ? "cursor-not-allowed" : "cursor-pointer")}>
        <Checkbox
          {...rest}
          id={id}
          checked={checked}
          onChange={onChange}
          aria-required={required ? true : undefined}
          aria-invalid={error ? true : undefined}
          aria-describedby={describedBy}
          className={cn("shrink-0", description && "mt-0.5")}
        />
        <span className={cn(hideLabel ? "sr-only" : "min-w-0", rest.disabled && "opacity-dim")}>
          <span className={cn("block text-sm text-text-primary", description && "font-medium")}>
            {label}
          </span>
          {description && (
            <span className="block text-2xs text-text-muted mt-0.5">
              {description}
            </span>
          )}
        </span>
      </label>
      {error ? (
        <FieldError id={errorId}>{error}</FieldError>
      ) : (
        <FieldHint id={hintId}>{hint}</FieldHint>
      )}
    </div>
  );
}
