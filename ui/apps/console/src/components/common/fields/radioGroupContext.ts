import { createContext, useContext } from "react";

interface RadioGroupContextValue {
  name: string;
  value: string;
  onChange: (value: string) => void;
}

/**
 * Carries the group's value, name and change handler to its options, so an option does not have
 * to be given them one prop at a time.
 */
export const RadioGroupContext = createContext<RadioGroupContextValue | null>(
  null,
);

/**
 * The enclosing radio group. Throws outside one rather than returning null, so an option placed
 * in the wrong tree fails where the mistake is.
 */
export function useRadioGroupContext(): RadioGroupContextValue {
  const ctx = useContext(RadioGroupContext);
  if (!ctx) {
    throw new Error("RadioCard must be used within a RadioGroupField");
  }
  return ctx;
}
