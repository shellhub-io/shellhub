import { useController, type Control, type FieldValues, type Path } from "react-hook-form";
import InputField from "@/components/common/fields/InputField";
import type { ReactNode, InputHTMLAttributes } from "react";

type InputProps = Omit<
  InputHTMLAttributes<HTMLInputElement>,
  "id" | "value" | "onChange"
>;

type Props<T extends FieldValues> = InputProps & {
  name: Path<T>;
  control: Control<T>;
  id: string;
  label: ReactNode;
  labelAdornment?: ReactNode;
  hideLabel?: boolean;
  error?: string;
  errorRole?: "alert" | "status";
  hint?: string;
  appendIcon?: ReactNode;
  prependIcon?: ReactNode;
  variant?: "default" | "mono";
  /** Called on every value change, in addition to RHF's internal onChange. */
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
