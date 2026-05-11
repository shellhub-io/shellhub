import { ComponentProps } from "react";
import InputField from "@/components/common/fields/InputField";

type Props = Omit<
  ComponentProps<typeof InputField>,
  "type" | "inputMode"
> & {
  allowNegative?: boolean;
};

export default function NumericInput({
  onChange,
  allowNegative = false,
  ...rest
}: Props) {
  const regex = allowNegative ? /^-?\d*$/ : /^\d*$/;
  const handleChange = (v: string) => {
    if (regex.test(v)) onChange(v);
  };

  return (
    <InputField
      {...rest}
      type="text"
      inputMode="numeric"
      onChange={handleChange}
    />
  );
}
