import { ReactNode } from "react";
import SavedMark from "./SavedMark";

interface SettingsSwitchCardProps {
  icon: ReactNode;
  title: string;
  description: ReactNode;
  control: ReactNode;
  error?: string;
  saved?: boolean;
  children?: ReactNode;
}

/**
 * A setting in a card of its own, with its control on the right: a switch, or a button that opens
 * the editor for a value too long to edit in place. children are the settings that only apply
 * while it is on, shown in a panel attached below. saved says, for a moment after a change is
 * stored, that it was.
 */
export default function SettingsSwitchCard({
  icon,
  title,
  description,
  control,
  error,
  saved = false,
  children,
}: SettingsSwitchCardProps) {
  return (
    <div role="group" aria-label={title}>
      <div className="flex flex-wrap items-center justify-between gap-x-6 gap-y-3 px-5 py-4 rounded-xl border border-border bg-card">
        <div className="min-w-0 flex-1 basis-60">
          <p className="flex items-center gap-2 text-sm font-medium text-text-primary">
            <span
              aria-hidden="true"
              className="shrink-0 text-text-muted [&>svg]:w-4 [&>svg]:h-4"
            >
              {icon}
            </span>
            {title}
            {saved && <SavedMark />}
          </p>
          <div className="mt-1 max-w-md text-xs text-text-muted leading-relaxed">
            {description}
          </div>
        </div>
        <div className="shrink-0">{control}</div>
      </div>
      {children && (
        <div className="mx-1 px-5 py-4 rounded-b-xl border border-t-0 border-border bg-surface">
          {children}
        </div>
      )}
      {error && (
        <p role="alert" className="mt-2 text-2xs text-accent-red">
          {error}
        </p>
      )}
    </div>
  );
}
