import { InputHTMLAttributes } from "react";

interface NumericInputProps extends Omit<
  InputHTMLAttributes<HTMLInputElement>,
  "onChange" | "value" | "type" | "inputMode" | "defaultValue"
> {
  value: string;
  onChange: (value: string) => void;
  allowNegative?: boolean;
}

export default function NumericInput({
  value,
  onChange,
  allowNegative = false,
  ...rest
}: NumericInputProps) {
  const handleChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const newValue = e.target.value;
    const regex = allowNegative ? /^-?\d*$/ : /^\d*$/;

    if (regex.test(newValue)) onChange(newValue);
  };

  return (
    <input
      type="text"
      inputMode="numeric"
      value={value}
      onChange={handleChange}
      {...rest}
    />
  );
}
