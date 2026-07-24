import { useRef, useState } from "react";
import {
  CalendarDaysIcon,
  CheckIcon,
  ChevronDownIcon,
} from "@heroicons/react/24/outline";
import { addDays, startOfMonth } from "date-fns";
import { DayPicker } from "react-day-picker";
import { useClickOutside } from "@/hooks/useClickOutside";
import { useEscapeKey } from "@/hooks/useEscapeKey";
import { formatDateShort } from "@/utils/date";
import { LABEL } from "@/utils/styles";
import { defaultExpiry, startOfDayUtc } from "./helpers";

/** Quick-pick offsets from today, so the common cases skip the calendar. */
const PRESETS = [
  { label: "30 days", days: 30 },
  { label: "60 days", days: 60 },
  { label: "90 days", days: 90 },
  { label: "1 year", days: 365 },
];

/**
 * Tailwind class map for react-day-picker. The library renders a real table, so
 * with no default stylesheet imported these classes fully own the look; every
 * colour is a design-system token, so the calendar flips with the app theme.
 * Selected/today target the day button (`[&>button]`) since the modifier class
 * lands on the surrounding cell.
 */
const CALENDAR_CLASSNAMES = {
  months: "relative flex flex-col",
  month: "flex flex-col",
  nav: "absolute top-0 inset-x-0 flex items-center justify-between h-9 z-raised",
  button_previous:
    "inline-flex items-center justify-center w-7 h-7 rounded-md text-text-secondary hover:bg-hover-subtle hover:text-text-primary transition-colors disabled:opacity-30 disabled:pointer-events-none",
  button_next:
    "inline-flex items-center justify-center w-7 h-7 rounded-md text-text-secondary hover:bg-hover-subtle hover:text-text-primary transition-colors disabled:opacity-30 disabled:pointer-events-none",
  chevron: "w-4 h-4 fill-current",
  month_caption: "flex items-center justify-center h-9",
  caption_label: "text-sm font-semibold text-text-primary",
  month_grid: "w-full border-collapse",
  weekdays: "flex",
  weekday:
    "w-9 h-8 flex items-center justify-center text-2xs font-mono font-medium uppercase tracking-label text-text-muted",
  week: "flex w-full",
  day: "p-0",
  day_button:
    "inline-flex items-center justify-center w-9 h-9 rounded-md text-xs text-text-secondary hover:bg-hover-subtle hover:text-text-primary transition-colors cursor-pointer",
  today: "[&>button]:text-primary [&>button]:font-semibold",
  selected:
    "[&>button]:bg-primary [&>button]:text-white [&>button]:font-semibold [&>button]:hover:bg-primary [&>button]:hover:text-white",
  outside: "[&>button]:text-text-muted/30",
  disabled:
    "[&>button]:text-text-muted/25 [&>button]:pointer-events-none [&>button]:hover:bg-transparent",
  hidden: "invisible",
};

/**
 * Expiration picker used by both drawers. A single field-trigger shows the
 * current choice ("Never expires" or an absolute date) and opens a popover with
 * quick presets, a calendar for a custom date, and a "Never expires" option.
 * `value` is an RFC3339 date string or null (never); the calendar and presets
 * always resolve to the start of the chosen day.
 */
export default function ExpirationField({
  value,
  onChange,
}: {
  value: string | null;
  onChange: (value: string | null) => void;
}) {
  const [open, setOpen] = useState(false);
  const wrapperRef = useRef<HTMLDivElement>(null);

  useClickOutside(wrapperRef, () => setOpen(false));
  useEscapeKey(() => setOpen(false), open);

  const never = value === null;
  const selected = value ? new Date(value) : undefined;
  const selectedDay = value ? value.slice(0, 10) : null;
  // Expiry must be in the future, so the earliest selectable day is tomorrow.
  const tomorrow = addDays(new Date(), 1);

  const pick = (iso: string | null) => {
    onChange(iso);
    setOpen(false);
  };

  return (
    <div>
      <span className={LABEL}>Expiration</span>
      <div ref={wrapperRef} className="relative">
        <button
          type="button"
          onClick={() => setOpen((v) => !v)}
          aria-haspopup="dialog"
          aria-expanded={open}
          className="w-full flex items-center gap-2 px-3.5 py-2.5 bg-card border border-border rounded-lg text-sm text-text-primary hover:border-border-light focus:outline-none focus:border-primary/50 focus:ring-1 focus:ring-primary/20 transition-all"
        >
          <CalendarDaysIcon className="w-4 h-4 text-text-muted shrink-0" />
          <span className="flex-1 text-left">
            {never ? "Never expires" : formatDateShort(value ?? "")}
          </span>
          <ChevronDownIcon
            className={`w-4 h-4 text-text-muted shrink-0 transition-transform ${open ? "rotate-180" : ""}`}
          />
        </button>

        {open && (
          <div
            role="dialog"
            aria-label="Choose an expiration date"
            className="absolute left-0 top-full mt-1 z-dropdown w-[19rem] p-3 bg-surface border border-border rounded-xl shadow-2xl animate-fade-in"
          >
            <div className="flex flex-wrap gap-1.5 mb-3">
              {PRESETS.map((preset) => {
                const iso = startOfDayUtc(addDays(new Date(), preset.days));
                const active = selectedDay === iso.slice(0, 10);
                return (
                  <button
                    key={preset.days}
                    type="button"
                    onClick={() => pick(iso)}
                    className={`px-2.5 py-1 rounded-md text-2xs font-medium border transition-colors ${
                      active
                        ? "bg-primary/10 border-primary/40 text-primary"
                        : "bg-card border-border text-text-secondary hover:border-border-light hover:text-text-primary"
                    }`}
                  >
                    {preset.label}
                  </button>
                );
              })}
            </div>

            <DayPicker
              mode="single"
              selected={selected}
              onSelect={(day) => day && pick(startOfDayUtc(day))}
              disabled={{ before: tomorrow }}
              startMonth={startOfMonth(new Date())}
              defaultMonth={selected ?? new Date()}
              showOutsideDays
              classNames={CALENDAR_CLASSNAMES}
            />

            <div className="mt-3 pt-3 border-t border-border">
              <button
                type="button"
                onClick={() => pick(never ? defaultExpiry() : null)}
                className={`w-full flex items-center justify-between px-2.5 py-2 rounded-lg text-xs font-medium transition-colors ${
                  never
                    ? "bg-primary/10 text-primary"
                    : "text-text-secondary hover:bg-hover-subtle hover:text-text-primary"
                }`}
              >
                Never expires
                {never && <CheckIcon className="w-4 h-4" />}
              </button>
            </div>
          </div>
        )}
      </div>
      <p className="text-2xs text-text-muted mt-1.5">
        {never
          ? "The key never expires."
          : "The key expires at the start of the chosen day."}
      </p>
    </div>
  );
}
