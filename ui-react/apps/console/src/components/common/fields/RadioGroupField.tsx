import { ReactNode, useId } from "react";
import { RadioGroupContext } from "@/components/common/fields/radioGroupContext";
import { LABEL } from "@/utils/styles";

type Props<T extends string> = {
  id?: string;
  value: T;
  onChange: (v: T) => void;
  containerClassName?: string;
  children: ReactNode;
} & (
  | { label: string; labelledBy?: never }
  | { label?: never; labelledBy: string }
);

export default function RadioGroupField<T extends string>({
  id,
  label,
  labelledBy,
  value,
  onChange,
  containerClassName = "space-y-2",
  children,
}: Props<T>) {
  const autoId = useId();
  const groupId = id ?? autoId;
  const internalLabelId = `${groupId}-label`;
  const ariaLabelledBy = labelledBy ?? internalLabelId;

  return (
    <div>
      {label && (
        <span id={internalLabelId} className={LABEL}>
          {label}
        </span>
      )}
      <div
        role="radiogroup"
        aria-labelledby={ariaLabelledBy}
        className={containerClassName}
      >
        <RadioGroupContext.Provider
          value={{
            name: groupId,
            value,
            onChange: onChange as (v: string) => void,
          }}
        >
          {children}
        </RadioGroupContext.Provider>
      </div>
    </div>
  );
}
