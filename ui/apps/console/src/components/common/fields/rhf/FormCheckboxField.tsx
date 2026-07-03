import { useController, type Control, type FieldValues, type Path } from "react-hook-form";
import CheckboxField from "@/components/common/fields/CheckboxField";
import type { ComponentProps } from "react";

type CheckboxFieldProps = Omit<
  ComponentProps<typeof CheckboxField>,
  "checked" | "onChange"
>;

type Props<T extends FieldValues> = CheckboxFieldProps & {
  name: Path<T>;
  control: Control<T>;
  /** Called on every value change, in addition to RHF's internal onChange. */
  onValueChange?: (value: boolean) => void;
};

export default function FormCheckboxField<T extends FieldValues>({
  name,
  control,
  error: errorOverride,
  onValueChange,
  ...rest
}: Props<T>) {
  const {
    field,
    fieldState: { error: fieldError },
  } = useController({ name, control });

  const resolvedError = errorOverride ?? fieldError?.message;

  return (
    <CheckboxField
      {...rest}
      checked={field.value}
      onChange={(v) => {
        field.onChange(v);
        onValueChange?.(v);
      }}
      onBlur={field.onBlur}
      error={resolvedError}
    />
  );
}
