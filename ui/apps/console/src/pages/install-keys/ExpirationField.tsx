import { useState } from "react";
import {
  CalendarDaysIcon,
  CheckIcon,
  ChevronDownIcon,
} from "@heroicons/react/24/outline";
import { addDays, differenceInCalendarDays, startOfMonth } from "date-fns";
import { DayPicker } from "react-day-picker";
import { Dropdown } from "@shellhub/design-system/primitives";
import { cn } from "@shellhub/design-system/cn";
import { formatDateShort } from "@/utils/date";
import { LABEL } from "@/utils/styles";

const PRESETS = [
  { label: "30 days", days: 30 },
  { label: "60 days", days: 60 },
  { label: "90 days", days: 90 },
  { label: "1 year", days: 365 },
];

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
  weekdays: "flex w-full",
  weekday:
    "flex-1 h-8 flex items-center justify-center text-2xs font-mono font-medium uppercase tracking-label text-text-muted",
  week: "flex w-full",
  day: "flex-1 p-0",
  day_button:
    "inline-flex items-center justify-center w-full h-9 rounded-md text-xs text-text-secondary hover:bg-hover-subtle hover:text-text-primary transition-colors cursor-pointer",
  today: "[&>button]:text-primary [&>button]:font-semibold",
  selected:
    "[&>button]:bg-primary [&>button]:text-white [&>button]:font-semibold [&>button]:hover:bg-primary [&>button]:hover:text-white",
  outside: "[&>button]:text-text-muted/30",
  disabled:
    "[&>button]:text-text-muted/25 [&>button]:pointer-events-none [&>button]:hover:bg-transparent",
  hidden: "invisible",
};

export default function ExpirationField({
  expiresIn,
  onExpiresInChange,
}: {
  expiresIn: string;
  onExpiresInChange: (value: string) => void;
}) {
  const [open, setOpen] = useState(false);

  const days = Number(expiresIn);
  const never = days < 1;
  const targetDate = never ? undefined : addDays(new Date(), days);
  const tomorrow = addDays(new Date(), 1);

  const pick = (value: number | null) => {
    onExpiresInChange(value === null ? "-1" : String(value));
    setOpen(false);
  };

  return (
    <div>
      <span className={LABEL}>Expiration</span>
      <Dropdown mode="content" open={open} onOpenChange={setOpen}>
        <Dropdown.Trigger>
          <button
            type="button"
            className="w-full flex items-center gap-2 px-3.5 py-2.5 bg-card border border-border rounded-lg text-sm text-text-primary hover:border-border-light focus:outline-none focus:border-primary/50 focus:ring-1 focus:ring-primary/20 transition-all"
          >
            <CalendarDaysIcon className="w-4 h-4 text-text-muted shrink-0" />
            <span className="flex-1 text-left">
              {never
                ? "Never expires"
                : formatDateShort(targetDate!.toISOString())}
            </span>
            <ChevronDownIcon
              className={cn(
                "w-4 h-4 text-text-muted shrink-0 transition-transform",
                open && "rotate-180",
              )}
            />
          </button>
        </Dropdown.Trigger>

        <Dropdown.Panel aria-label="Choose an expiration date" className="p-3">
          <div className="flex flex-wrap gap-1.5 mb-3">
            {PRESETS.map((preset) => {
              const active = days === preset.days;
              return (
                <button
                  key={preset.days}
                  type="button"
                  onClick={() => pick(preset.days)}
                  className={cn(
                    "px-2.5 py-1 rounded-md text-2xs font-medium border transition-colors",
                    active
                      ? "bg-primary/10 border-primary/40 text-primary"
                      : "bg-card border-border text-text-secondary hover:border-border-light hover:text-text-primary",
                  )}
                >
                  {preset.label}
                </button>
              );
            })}
          </div>

          <DayPicker
            mode="single"
            selected={targetDate}
            onSelect={(day?: Date) => {
              if (!day) return;
              const d = differenceInCalendarDays(day, new Date());
              if (d > 0) pick(d);
            }}
            disabled={{ before: tomorrow }}
            startMonth={startOfMonth(new Date())}
            defaultMonth={targetDate ?? new Date()}
            showOutsideDays
            classNames={CALENDAR_CLASSNAMES}
          />

          <div className="mt-3 pt-3 border-t border-border">
            <button
              type="button"
              onClick={() => pick(never ? 30 : null)}
              className={cn(
                "w-full flex items-center justify-between px-2.5 py-2 rounded-lg text-xs font-medium transition-colors",
                never
                  ? "bg-primary/10 text-primary"
                  : "text-text-secondary hover:bg-hover-subtle hover:text-text-primary",
              )}
            >
              Never expires
              {never && <CheckIcon className="w-4 h-4" />}
            </button>
          </div>
        </Dropdown.Panel>
      </Dropdown>
    </div>
  );
}
