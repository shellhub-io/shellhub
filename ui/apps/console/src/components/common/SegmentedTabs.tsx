import type { KeyboardEvent } from "react";
import { cn } from "@shellhub/design-system/cn";

/**
 * One choice in a SegmentedTabs row. value must be unique within the row: it is both the React key
 * and how the selected tab is recognised. A tab that switches a panel sets panel, and the caller
 * renders that panel with role="tabpanel", id={panel.panelId} and aria-labelledby={panel.tabId}
 * whenever the tab is selected; a tab that only filters a list leaves panel unset.
 */
export interface SegmentedTab<T extends string> {
  label: string;
  value: T;
  panel?: { tabId: string; panelId: string };
}

/**
 * A segmented row of mutually exclusive choices, announced as a tablist: a filter over a list or
 * a switch between panels. Only the selected tab is in the tab order; the arrow keys, Home and End
 * move to another tab and select it at once. A value matching no tab leaves the first one
 * focusable.
 */
export default function SegmentedTabs<T extends string>({
  tabs,
  value,
  onChange,
  label,
}: {
  tabs: SegmentedTab<T>[];
  value: T;
  onChange: (value: T) => void;
  label: string;
}) {
  const selected = Math.max(
    0,
    tabs.findIndex((tab) => tab.value === value),
  );

  const moveTo = (e: KeyboardEvent<HTMLDivElement>) => {
    const target =
      e.key === "ArrowRight"
        ? (selected + 1) % tabs.length
        : e.key === "ArrowLeft"
          ? (selected - 1 + tabs.length) % tabs.length
          : e.key === "Home"
            ? 0
            : e.key === "End"
              ? tabs.length - 1
              : null;
    if (target === null) return;
    e.preventDefault();
    onChange(tabs[target].value);
    const buttons =
      e.currentTarget.querySelectorAll<HTMLElement>('[role="tab"]');
    buttons[target]?.focus();
  };

  return (
    // eslint-disable-next-line jsx-a11y/interactive-supports-focus -- focus sits on the tabs; the list only relays the arrow keys between them
    <div
      className="flex items-center h-8 bg-card border border-border rounded-md p-0.5"
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
          onClick={() => onChange(tab.value)}
          className={cn(
            "h-full px-3.5 text-xs font-medium rounded transition-all duration-150",
            value === tab.value
              ? "bg-primary/15 text-primary border border-primary/25"
              : "text-text-muted hover:text-text-secondary border border-transparent",
          )}
        >
          {tab.label}
        </button>
      ))}
    </div>
  );
}
