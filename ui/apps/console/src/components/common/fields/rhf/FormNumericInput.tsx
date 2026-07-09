import { useController, type Control, type FieldValues, type Path } from "react-hook-form";
import NumericInput from "@/components/common/fields/NumericInput";
import type { ComponentProps } from "react";

type NumericInputProps = Omit<
  ComponentProps<typeof NumericInput>,
  "value" | "onChange"
>;

type Props<T extends FieldValues> = NumericInputProps & {
  name: Path<T>;
  control: Control<T>;
};

export default function FormNumericInput<T extends FieldValues>({
  name,
  control,
  error: errorOverride,
  ...rest
}: Props<T>) {
  const {
    field,
    fieldState: { error: fieldError },
  } = useController({ name, control });

  const resolvedError = errorOverride ?? fieldError?.message;

  return (
    <NumericInput
      {...rest}
      value={field.value}
      onChange={(v) => {
        field.onChange(v);
      }}
      onBlur={field.onBlur}
      error={resolvedError}
    />
  );
}
