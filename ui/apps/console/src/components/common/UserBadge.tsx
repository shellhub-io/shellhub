import { ReactNode } from "react";
import { cn } from "@shellhub/design-system/cn";
import { getInitials } from "@/utils/string";

/**
 * A monogram plus a name and email, for showing who a row belongs to. By
 * default it stacks the name over the email; `short` shows only the primary
 * label (the name, or the email when there is no name). Falls back gracefully
 * when either is missing.
 */
export default function UserBadge({
  name,
  email,
  short = false,
  trailing,
  secondary,
  className,
}: {
  name?: string | null;
  email?: string | null;
  short?: boolean;
  trailing?: ReactNode;
  secondary?: ReactNode;
  className?: string;
}) {
  const primary = name || email || "—";
  const secondaryLine = secondary ?? (name && email ? email : undefined);

  return (
    <span className={cn("inline-flex items-center gap-2.5 min-w-0", className)}>
      <span className="grid place-items-center w-8 h-8 rounded-lg shrink-0 bg-card border border-border text-2xs font-bold font-mono text-text-muted">
        {getInitials(primary) || "?"}
      </span>
      <span className="flex flex-col min-w-0">
        <span className="flex items-center gap-1.5 text-sm font-medium text-text-primary truncate">
          {primary}
          {trailing}
        </span>
        {!short && secondaryLine && (
          <span className="text-xs text-text-muted truncate">
            {secondaryLine}
          </span>
        )}
      </span>
    </span>
  );
}
