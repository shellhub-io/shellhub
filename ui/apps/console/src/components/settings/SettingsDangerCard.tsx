import { ReactNode } from "react";

/**
 * An action that can't be undone, in a red card: what it does on the left, the control that
 * starts it on the right, or below on a narrow window. action is expected to ask before acting.
 */
export default function SettingsDangerCard({
  title,
  description,
  action,
}: {
  title: string;
  description: ReactNode;
  action: ReactNode;
}) {
  return (
    <div
      role="group"
      aria-label={title}
      className="flex flex-col sm:flex-row sm:items-center gap-4 sm:gap-8 px-5 py-4 rounded-xl border border-accent-red/30 bg-accent-red/[0.04]"
    >
      <div className="min-w-0 flex-1">
        <p className="text-sm font-medium text-text-primary">{title}</p>
        <p className="mt-1 text-xs text-text-muted leading-relaxed">
          {description}
        </p>
      </div>
      <div className="shrink-0 self-start sm:self-auto">{action}</div>
    </div>
  );
}
