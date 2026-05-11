import { InputHTMLAttributes, ReactNode } from "react";
import {
  INPUT,
  INPUT_ERROR,
  INPUT_MONO,
  INPUT_MONO_ERROR,
  INPUT_READONLY,
} from "@/utils/styles";
import FieldLabel from "@/components/common/fields/FieldLabel";
import FieldError from "@/components/common/fields/FieldError";
import FieldHint from "@/components/common/fields/FieldHint";

type InputProps = Omit<
  InputHTMLAttributes<HTMLInputElement>,
  "id" | "value" | "onChange"
>;

type Props = InputProps & {
  id: string;
  label: ReactNode;
  /** Ignored when `hideLabel` is true. */
  labelAdornment?: ReactNode;
  /** Visually hide the label (kept for screen readers via `sr-only`). */
  hideLabel?: boolean;
  value: string;
  onChange: (v: string) => void;
  error?: string;
  errorRole?: "alert" | "status";
  hint?: string;
  appendIcon?: ReactNode;
  variant?: "default" | "mono";
};

export default function InputField({
  id,
  label,
  labelAdornment,
  hideLabel = false,
  value,
  onChange,
  error,
  errorRole,
  hint,
  appendIcon,
  variant = "default",
  type = "text",
  className,
  ...rest
}: Props) {
  const errorId = `${id}-error`;
  const hintId = `${id}-hint`;
  const describedBy = error ? errorId : hint ? hintId : undefined;
  const baseByVariant = {
    default: error ? INPUT_ERROR : INPUT,
    mono: error ? INPUT_MONO_ERROR : INPUT_MONO,
  };
  const inputClassNames = [
    baseByVariant[variant],
    appendIcon && "pr-10",
    rest.readOnly && INPUT_READONLY,
    className,
  ]
    .filter(Boolean)
    .join(" ");

  return (
    <div>
      <FieldLabel htmlFor={id} hideLabel={hideLabel} adornment={labelAdornment}>
        {label}
      </FieldLabel>
      <div className={appendIcon ? "relative" : undefined}>
        <input
          {...rest}
          id={id}
          type={type}
          value={value}
          onChange={(e) => onChange(e.target.value)}
          aria-required={rest.required ? true : undefined}
          aria-invalid={error ? true : undefined}
          aria-describedby={describedBy}
          className={inputClassNames}
        />
        {appendIcon && (
          <div className="absolute right-2.5 top-1/2 -translate-y-1/2">
            {appendIcon}
          </div>
        )}
      </div>
      {error ? (
        <FieldError id={errorId} role={errorRole}>
          {error}
        </FieldError>
      ) : (
        <FieldHint id={hintId}>{hint}</FieldHint>
      )}
    </div>
  );
}
