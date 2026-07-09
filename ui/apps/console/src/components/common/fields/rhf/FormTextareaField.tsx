import type { TextareaHTMLAttributes } from "react";
import { useController, type Control, type FieldValues, type Path } from "react-hook-form";
import { INPUT } from "@/utils/styles";
import FieldLabel from "@/components/common/fields/FieldLabel";
import FieldError from "@/components/common/fields/FieldError";
import { cn } from "@shellhub/design-system/cn";

type TextareaProps = Omit<
  TextareaHTMLAttributes<HTMLTextAreaElement>,
  "id" | "value" | "onChange"
>;

type Props<T extends FieldValues> = TextareaProps & {
  id: string;
  label: string;
  name: Path<T>;
  control: Control<T>;
  error?: string;
  /** Called on every value change, in addition to RHF's internal onChange. */
  onValueChange?: (value: string) => void;
};

export default function FormTextareaField<T extends FieldValues>({
  id,
  label,
  name,
  control,
  error: errorOverride,
  onValueChange,
  className,
  ...rest
}: Props<T>) {
  const {
    field,
    fieldState: { error: fieldError },
  } = useController({ name, control });

  const resolvedError = errorOverride ?? fieldError?.message;
  const errorId = `${id}-error`;

  return (
    <div>
      <FieldLabel htmlFor={id}>{label}</FieldLabel>
      <textarea
        {...rest}
        id={id}
        value={field.value}
        onChange={(e) => {
          field.onChange(e.target.value);
          onValueChange?.(e.target.value);
        }}
        onBlur={field.onBlur}
        aria-invalid={resolvedError ? true : undefined}
        aria-describedby={resolvedError ? errorId : undefined}
        className={cn(INPUT, className)}
      />
      <FieldError id={errorId}>{resolvedError}</FieldError>
    </div>
  );
}
