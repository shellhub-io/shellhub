import { InputHTMLAttributes, ReactNode } from "react";
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
      <label
        htmlFor={id}
        title={title}
        className={`flex ${alignment} gap-2.5 ${
          rest.disabled ? "cursor-not-allowed opacity-60" : "cursor-pointer"
        }`}
      >
        <input
          {...rest}
          id={id}
          type="checkbox"
          checked={checked}
          onChange={(e) => onChange(e.target.checked)}
          aria-required={required ? true : undefined}
          aria-invalid={error ? true : undefined}
          aria-describedby={describedBy}
          className={`${description ? "mt-0.5 " : ""}shrink-0 w-4 h-4 rounded border-border bg-card text-primary focus:ring-1 focus:ring-primary/30 disabled:cursor-not-allowed`}
        />
        <span className={hideLabel ? "sr-only" : "min-w-0"}>
          <span
            className={`block text-sm ${
              description ? "font-medium" : ""
            } text-text-primary`}
          >
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
