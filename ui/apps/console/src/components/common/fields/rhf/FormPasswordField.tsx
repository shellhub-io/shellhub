import { useController, type Control, type FieldValues, type Path } from "react-hook-form";
import PasswordField from "@/components/common/fields/PasswordField";
import type { ComponentProps } from "react";

type PasswordFieldProps = Omit<
  ComponentProps<typeof PasswordField>,
  "value" | "onChange"
>;

type Props<T extends FieldValues> = PasswordFieldProps & {
  name: Path<T>;
  control: Control<T>;
  /** Called on every value change, in addition to RHF's internal onChange. */
  onValueChange?: (value: string) => void;
};

export default function FormPasswordField<T extends FieldValues>({
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
    <PasswordField
      {...rest}
      value={field.value}
      onChange={(v) => {
        field.onChange(v);
        onValueChange?.(v);
      }}
      onBlur={field.onBlur}
      error={resolvedError}
    />
  );
}
