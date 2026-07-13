import { ReactNode } from "react";
import { cn } from "@shellhub/design-system/cn";
import { useRadioGroupContext } from "@/components/common/fields/radioGroupContext";

export default function RadioPill<T extends string>({
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
        "inline-flex items-center gap-1.5 px-3 py-1.5 text-xs font-medium rounded-md border cursor-pointer transition-all focus-within:ring-2 focus-within:ring-primary/40",
        selected
          ? "bg-primary/[0.08] border-primary/30 text-primary ring-1 ring-primary/10"
          : "bg-card border-border text-text-secondary hover:border-border-light",
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
