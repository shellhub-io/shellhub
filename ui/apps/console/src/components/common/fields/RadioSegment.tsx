import { ReactNode } from "react";
import { cn } from "@shellhub/design-system/cn";
import { useRadioGroupContext } from "@/components/common/fields/radioGroupContext";

export default function RadioSegment<T extends string>({
  value,
  label,
  icon,
}: {
  value: T;
  label: string;
  icon?: ReactNode;
}) {
  const { name, value: groupValue, onChange } = useRadioGroupContext();
  const selected = groupValue === value;

  return (
    <label
      className={cn(
        "flex-1 flex items-center justify-center gap-1.5 px-3 py-1.5 rounded-md text-xs font-medium cursor-pointer transition-all focus-within:ring-2 focus-within:ring-primary/40",
        selected
          ? "bg-primary/10 text-primary border border-primary/20"
          : "text-text-secondary hover:text-text-primary border border-transparent",
      )}
    >
      <input
        type="radio"
        name={name}
        value={value}
        checked={selected}
        onChange={() => onChange(value)}
        className="sr-only"
      />
      {icon}
      {label}
    </label>
  );
}
