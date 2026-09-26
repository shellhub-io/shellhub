import { ReactNode } from "react";
import { cn } from "@shellhub/design-system/cn";

interface SettingsSectionProps {
  title: string;
  description: ReactNode;
  action?: ReactNode;
  wide?: boolean;
  children: ReactNode;
}

/**
 * One section of the settings area: a heading that says what the section governs, an optional
 * action beside it, and the settings below. The column stays narrow for reading unless the section
 * holds a table, which asks for wide.
 */
export default function SettingsSection({
  title,
  description,
  action,
  wide = false,
  children,
}: SettingsSectionProps) {
  return (
    <section className={cn(!wide && "max-w-2xl")}>
      <header className="flex flex-col sm:flex-row sm:items-start sm:justify-between gap-4 mb-6">
        <div className="min-w-0">
          <h2 className="text-lg font-semibold text-text-primary">{title}</h2>
          <p className="mt-1 text-sm text-text-muted">{description}</p>
        </div>
        {action && <div className="shrink-0">{action}</div>}
      </header>
      <div className="space-y-5">{children}</div>
    </section>
  );
}
