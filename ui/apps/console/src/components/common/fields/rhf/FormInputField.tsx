import { useController, type Control, type FieldValues, type Path } from "react-hook-form";
import InputField from "@/components/common/fields/InputField";
import type { ComponentProps } from "react";

type InputFieldProps = Omit<
  ComponentProps<typeof InputField>,
  "value" | "onChange"
>;

type Props<T extends FieldValues> = InputFieldProps & {
  name: Path<T>;
  control: Control<T>;
  onValueChange?: (value: string) => void;
};

export default function FormInputField<T extends FieldValues>({
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
    <InputField
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
