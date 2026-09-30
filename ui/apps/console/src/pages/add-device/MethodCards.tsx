import { useState } from "react";
import { ChevronDownIcon, ChevronUpIcon } from "@heroicons/react/24/outline";
import { cn } from "@shellhub/design-system/cn";
import RadioCard from "@/components/common/fields/RadioCard";
import RadioGroupField from "@/components/common/fields/RadioGroupField";
import { METHODS, type Method } from "@/pages/install/methods";

const INITIAL_VISIBLE = 3;

/**
 * The installation methods as cards with what each does, the first few shown and the rest behind
 * a toggle. A chosen method from the rest stays visible when the list folds.
 */
export default function MethodCards({
  labelledBy,
  value,
  onChange,
}: {
  labelledBy: string;
  value: Method;
  onChange: (method: Method) => void;
}) {
  const [showAll, setShowAll] = useState(false);
  const selected = METHODS.find((m) => m.id === value)!;
  const base = METHODS.slice(0, INITIAL_VISIBLE);
  const visible = showAll
    ? METHODS
    : base.includes(selected)
      ? base
      : [...base, selected];

  return (
    <>
      <RadioGroupField
        labelledBy={labelledBy}
        value={value}
        onChange={onChange}
      >
        {visible.map((m) => (
          <RadioCard
            key={m.id}
            value={m.id}
            icon={m.icon}
            label={m.label}
            description={m.description}
            adornment={
              m.tag && (
                <span
                  className={cn(
                    "px-1.5 py-0.5 text-3xs font-bold uppercase tracking-wider rounded border",
                    m.tag === "Manual"
                      ? "bg-accent-yellow/10 text-accent-yellow border-accent-yellow/20"
                      : "bg-accent-green/15 text-accent-green border-accent-green/20",
                  )}
                >
                  {m.tag}
                </span>
              )
            }
          />
        ))}
      </RadioGroupField>
      <button
        type="button"
        onClick={() => setShowAll(!showAll)}
        className="flex items-center justify-center gap-1.5 w-full py-2 text-2xs font-mono text-text-muted hover:text-primary transition-colors"
      >
        {showAll ? (
          <ChevronUpIcon className="w-3 h-3" strokeWidth={2} />
        ) : (
          <ChevronDownIcon className="w-3 h-3" strokeWidth={2} />
        )}
        {showAll ? "Show fewer" : "Show all methods"}
      </button>
    </>
  );
}
