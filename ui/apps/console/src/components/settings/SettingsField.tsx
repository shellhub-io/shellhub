import { ReactNode } from "react";
import { cn } from "@shellhub/design-system/cn";

interface SettingsFieldProps {
  title: string;
  description: ReactNode;
  stacked?: boolean;
  titleId?: string;
  children?: ReactNode;
}

/**
 * A setting that holds a value: what it is on the left and the value, or the control that edits
 * it, on the right. stacked puts the control below the text, for one that needs the full width.
 * titleId lets a control inside, such as a radio group, take the title as its name.
 */
export default function SettingsField({
  title,
  description,
  stacked = false,
  titleId,
  children,
}: SettingsFieldProps) {
  return (
    <div
      role="group"
      aria-label={title}
      className={cn(
        "flex gap-3",
        stacked ? "flex-col" : "flex-col sm:flex-row sm:items-center sm:gap-8",
      )}
    >
      <div className="min-w-0 flex-1">
        <p id={titleId} className="text-sm font-medium text-text-primary">
          {title}
        </p>
        <div className="mt-1 text-xs text-text-muted leading-relaxed">
          {description}
        </div>
      </div>
      <div className={cn(!stacked && "shrink-0")}>{children}</div>
    </div>
  );
}
