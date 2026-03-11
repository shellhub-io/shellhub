import { ReactNode } from "react";

export default function RadioCard({
  selected,
  onClick,
  icon,
  label,
  description,
}: {
  selected: boolean;
  onClick: () => void;
  icon: ReactNode;
  label: string;
  description: string;
}) {
  return (
    <button
      type="button"
      onClick={onClick}
      className={`flex items-start gap-3 w-full px-3.5 py-3 rounded-lg border text-left transition-all ${
        selected
          ? "bg-primary/[0.06] border-primary/30 ring-1 ring-primary/10"
          : "bg-card border-border hover:border-border-light hover:bg-hover-subtle"
      }`}
    >
      <div
        className={`mt-0.5 shrink-0 w-4 h-4 rounded-full border-2 flex items-center justify-center transition-all ${
          selected ? "border-primary" : "border-text-muted/40"
        }`}
      >
        {selected && <div className="w-2 h-2 rounded-full bg-primary" />}
      </div>
      <div className="flex items-start gap-2.5 min-w-0">
        <span
          className={`mt-0.5 shrink-0 transition-colors ${selected ? "text-primary" : "text-text-muted"}`}
        >
          {icon}
        </span>
        <div className="min-w-0">
          <span
            className={`block text-sm font-medium transition-colors ${selected ? "text-text-primary" : "text-text-secondary"}`}
          >
            {label}
          </span>
          <span className="block text-2xs text-text-muted mt-0.5">
            {description}
          </span>
        </div>
      </div>
    </button>
  );
}
