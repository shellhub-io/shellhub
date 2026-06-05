import { ChevronRightIcon } from "@heroicons/react/24/outline";
import { optionId, type BadgeVariant, type CommandItem } from "./items";

const badgeStyles: Record<BadgeVariant, string> = {
  green: "text-accent-green bg-accent-green/10 border-accent-green/20",
  yellow: "text-accent-yellow bg-accent-yellow/10 border-accent-yellow/20",
  red: "text-accent-red bg-accent-red/10 border-accent-red/20",
  muted: "text-text-muted bg-hover-medium border-border",
};

interface CommandRowProps {
  item: CommandItem;
  isActive: boolean;
  shaking: boolean;
  onActivate: () => void;
}

/** A single combobox option row. Non-focusable (`role="option"`); the input
 *  owns focus via `aria-activedescendant`. Hosts an optional drill-in chevron.
 *  Dimmed and inert (no click/Enter) when `item.disabled`. */
export default function CommandRow({
  item,
  isActive,
  shaking,
  onActivate,
}: CommandRowProps) {
  return (
    <div
      id={optionId(item.id)}
      role="option"
      aria-selected={isActive}
      aria-disabled={item.disabled || undefined}
      data-active={isActive}
      onClick={item.disabled ? undefined : item.onSelect}
      onMouseEnter={onActivate}
      className={`w-full flex items-center gap-3 px-4 py-2.5 text-left transition-colors duration-75 ${
        isActive
          ? "bg-primary/10"
          : item.disabled
            ? ""
            : "hover:bg-hover-subtle"
      } ${item.disabled ? "opacity-50 cursor-not-allowed" : "cursor-pointer"} ${
        shaking ? "motion-safe:animate-shake" : ""
      }`}
    >
      <span
        className={`shrink-0 ${isActive ? "text-primary" : "text-text-muted"} transition-colors duration-75`}
        aria-hidden="true"
      >
        {item.icon}
      </span>
      <div className="flex-1 min-w-0 truncate">
        <span
          className={`text-sm ${isActive ? "text-text-primary" : "text-text-secondary"} transition-colors duration-75`}
        >
          {item.label}
        </span>
        {item.sublabel && (
          <span className="text-2xs text-text-muted/50 ml-2 font-mono">
            {item.sublabel}
          </span>
        )}
      </div>
      {item.badge && (
        <span
          className={`shrink-0 text-2xs font-mono font-semibold px-1.5 py-0.5 rounded border ${badgeStyles[item.badge.variant]}`}
        >
          {item.badge.text}
        </span>
      )}
      {isActive && !item.disabled && (
        <kbd
          className="shrink-0 px-1.5 py-0.5 text-2xs font-mono text-text-muted/40 bg-hover-subtle border border-border/50 rounded"
          aria-hidden="true"
        >
          ↵
        </kbd>
      )}
      {item.onDrillIn && (
        <button
          type="button"
          tabIndex={-1}
          aria-label={`Show actions for ${item.label}`}
          onClick={(e) => {
            e.stopPropagation();
            item.onDrillIn?.();
          }}
          className={`shrink-0 p-0.5 rounded transition-colors hover:text-primary hover:bg-hover-medium ${
            isActive ? "text-primary" : "text-text-muted/40"
          }`}
        >
          <ChevronRightIcon className="w-4 h-4" />
        </button>
      )}
    </div>
  );
}
