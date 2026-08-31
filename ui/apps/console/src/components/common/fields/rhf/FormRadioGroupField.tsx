import { type ReactNode } from "react";
import { useController, type Control, type FieldValues, type Path } from "react-hook-form";
import RadioGroupField from "@/components/common/fields/RadioGroupField";

type Props<TFieldValues extends FieldValues, TOption extends string> = {
  name: Path<TFieldValues>;
  control: Control<TFieldValues>;
  id?: string;
  containerClassName?: string;
  error?: string;
  onValueChange?: (value: TOption) => void;
  children: ReactNode;
} & (
  | { label: string; labelledBy?: never }
  | { label?: never; labelledBy: string }
);

/**
 * RadioGroupField bound to react-hook-form.
 */
export default function FormRadioGroupField<
  TFieldValues extends FieldValues,
  TOption extends string,
>({
  name,
  control,
  id,
  containerClassName,
  error: errorOverride,
  onValueChange,
  children,
  ...labelProps
}: Props<TFieldValues, TOption>) {
  const {
    field,
    fieldState: { error: fieldError },
  } = useController({ name, control });

  const resolvedError = errorOverride ?? fieldError?.message;

  return (
    <RadioGroupField<TOption>
      id={id}
      containerClassName={containerClassName}
      value={field.value}
      onChange={(v) => {
        field.onChange(v);
        onValueChange?.(v);
      }}
      error={resolvedError}
      {...labelProps}
    >
      {children}
    </RadioGroupField>
  );
}
