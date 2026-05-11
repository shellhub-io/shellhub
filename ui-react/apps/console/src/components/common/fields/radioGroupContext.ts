import { createContext, useContext } from "react";

interface RadioGroupContextValue {
  name: string;
  value: string;
  onChange: (value: string) => void;
}

export const RadioGroupContext = createContext<RadioGroupContextValue | null>(
  null,
);

export function useRadioGroupContext(): RadioGroupContextValue {
  const ctx = useContext(RadioGroupContext);
  if (!ctx) {
    throw new Error("RadioCard must be used within a RadioGroupField");
  }
  return ctx;
}
