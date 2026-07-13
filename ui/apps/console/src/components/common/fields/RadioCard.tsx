import { ReactNode } from "react";
import { cn } from "@shellhub/design-system/cn";
import { useRadioGroupContext } from "@/components/common/fields/radioGroupContext";

export default function RadioCard<T extends string>({
  value,
  icon,
  label,
  description,
  adornment,
}: {
  value: T;
  icon: ReactNode;
  label: string;
  description: string;
  adornment?: ReactNode;
}) {
  const { name, value: groupValue, onChange } = useRadioGroupContext();
  const selected = groupValue === value;

  return (
    <label
      className={cn(
        "flex items-start gap-3 w-full px-3.5 py-3 rounded-lg border text-left transition-all cursor-pointer focus-within:ring-2 focus-within:ring-primary/40",
        selected
          ? "bg-primary/[0.06] border-primary/30 ring-1 ring-primary/10"
          : "bg-card border-border hover:border-border-light hover:bg-hover-subtle",
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
      <div
        aria-hidden="true"
        className={cn(
          "mt-0.5 shrink-0 w-4 h-4 rounded-full border-2 flex items-center justify-center transition-all",
          selected ? "border-primary" : "border-text-muted/40",
        )}
      >
        {selected && <div className="w-2 h-2 rounded-full bg-primary" />}
      </div>
      <div className="flex items-start gap-2.5 min-w-0">
        <span
          className={cn("mt-0.5 shrink-0 transition-colors", selected ? "text-primary" : "text-text-muted")}
        >
          {icon}
        </span>
        <div className="min-w-0">
          <span className={adornment ? "flex items-center gap-2" : "block"}>
            <span
              className={cn("block min-w-0 truncate text-sm font-medium transition-colors", selected ? "text-text-primary" : "text-text-secondary")}
            >
              {label}
            </span>
            {adornment && <span className="shrink-0">{adornment}</span>}
          </span>
          <span className="block text-2xs text-text-muted mt-0.5">
            {description}
          </span>
        </div>
      </div>
    </label>
  );
}
