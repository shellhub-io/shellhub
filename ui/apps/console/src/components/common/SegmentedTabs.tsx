import type { KeyboardEvent } from "react";
import { cn } from "@shellhub/design-system/cn";

/**
 * One choice in a SegmentedTabs row. value must be unique within the row: it is both the React key
 * and how the selected tab is recognised. A tab that switches a panel sets panel, and the caller
 * renders that panel with role="tabpanel", id={panel.panelId} and aria-labelledby={panel.tabId}
 * whenever the tab is selected; a tab that only filters a list leaves panel unset. A disabled tab
 * cannot be clicked and the arrow keys, Home and End pass over it.
 */
export interface SegmentedTab<T extends string> {
  label: string;
  value: T;
  panel?: { tabId: string; panelId: string };
  disabled?: boolean;
}

/**
 * A segmented row of mutually exclusive choices, announced as a tablist: a filter over a list or
 * a switch between panels. Only the selected tab is in the tab order; the arrow keys, Home and End
 * move to another enabled tab and select it at once. A value matching no tab leaves the first one
 * focusable. value should name an enabled tab: naming a disabled one leaves the row with nothing
 * in the tab order, and so does naming none while the first tab is disabled. The underline variant
 * draws the tabs as a bar over the panels instead of a pill row.
 */
export default function SegmentedTabs<T extends string>({
  tabs,
  value,
  onChange,
  label,
  variant = "segmented",
  className,
}: {
  tabs: SegmentedTab<T>[];
  value: T;
  onChange: (value: T) => void;
  label: string;
  variant?: "segmented" | "underline";
  className?: string;
}) {
  const enabled = tabs.flatMap((tab, index) => (tab.disabled ? [] : [index]));
  const selected = Math.max(
    0,
    tabs.findIndex((tab) => tab.value === value),
  );

  const moveTo = (e: KeyboardEvent<HTMLDivElement>) => {
    const at = enabled.indexOf(selected);
    const target =
      e.key === "ArrowRight"
        ? enabled[(at + 1) % enabled.length]
        : e.key === "ArrowLeft"
          ? enabled[(at - 1 + enabled.length) % enabled.length]
          : e.key === "Home"
            ? enabled[0]
            : e.key === "End"
              ? enabled[enabled.length - 1]
              : undefined;
    if (target === undefined) return;
    e.preventDefault();
    onChange(tabs[target].value);
    const buttons =
      e.currentTarget.querySelectorAll<HTMLElement>('[role="tab"]');
    buttons[target]?.focus();
  };

  return (
    // eslint-disable-next-line jsx-a11y/interactive-supports-focus -- focus sits on the tabs; the list only relays the arrow keys between them
    <div
      className={cn(
        variant === "segmented"
          ? "flex items-center h-8 bg-card border border-border rounded-md p-0.5"
          : "flex items-center gap-1 border-b border-border",
        className,
      )}
      role="tablist"
      aria-label={label}
      onKeyDown={moveTo}
    >
      {tabs.map((tab, index) => (
        <button
          type="button"
          key={tab.value}
          role="tab"
          id={tab.panel?.tabId}
          aria-controls={tab.panel?.panelId}
          aria-selected={value === tab.value}
          tabIndex={index === selected ? 0 : -1}
          disabled={tab.disabled}
          onClick={() => onChange(tab.value)}
          className={cn(
            variant === "segmented"
              ? cn(
                  "h-full px-3.5 text-xs font-medium rounded transition-all duration-150",
                  value === tab.value
                    ? "bg-primary/15 text-primary border border-primary/25"
                    : "text-text-muted hover:text-text-secondary border border-transparent",
                )
              : cn(
                  "relative px-4 py-2.5 text-xs font-semibold transition-colors focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-primary/50 rounded-t-md",
                  value === tab.value
                    ? "text-text-primary"
                    : "text-text-muted hover:text-text-secondary",
                ),
            tab.disabled && "opacity-40 cursor-not-allowed hover:text-text-muted",
          )}
        >
          {tab.label}
          {variant === "underline" && value === tab.value && (
            <span
              aria-hidden="true"
              className="absolute left-2 right-2 -bottom-px h-0.5 bg-primary rounded-full"
            />
          )}
        </button>
      ))}
    </div>
  );
}
