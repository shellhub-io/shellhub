import { ReactNode } from "react";
import { cn } from "@shellhub/design-system/cn";
import { useRadioGroupContext } from "@/components/common/fields/radioGroupContext";

/**
 * One option of a RadioGroupField, drawn as a small tile: a picture of the choice above its
 * label. It reads the group from context, so it only works inside one.
 */
export default function RadioTile<T extends string>({
  value,
  picture,
  label,
}: {
  value: T;
  picture: ReactNode;
  label: string;
}) {
  const { name, value: groupValue, onChange } = useRadioGroupContext();
  const selected = groupValue === value;

  return (
    <label
      className={cn(
        "block rounded-lg border p-2 cursor-pointer transition-all duration-150 focus-within:ring-2 focus-within:ring-primary/40",
        selected
          ? "border-primary/40 bg-primary/[0.06] ring-1 ring-primary/10 text-primary"
          : "border-border hover:border-border-light bg-hover-subtle text-text-muted",
      )}
    >
      <input
        type="radio"
        name={name}
        value={value}
        checked={selected}
        aria-label={label}
        onChange={() => onChange(value)}
        className="sr-only"
      />
      <span aria-hidden="true" className="block">
        {picture}
      </span>
      <span
        aria-hidden="true"
        className="mt-1.5 block font-mono text-[11px] text-center truncate"
      >
        {label}
      </span>
    </label>
  );
}
