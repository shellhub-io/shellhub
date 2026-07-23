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
  className,
}: {
  name?: string | null;
  email?: string | null;
  short?: boolean;
  /** Rendered next to the primary label, e.g. a "you" chip. */
  trailing?: ReactNode;
  className?: string;
}) {
  const primary = name || email || "—";
  // The email only earns the secondary line when a name occupies the primary.
  const secondary = name && email ? email : undefined;

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
        {!short && secondary && (
          <span className="text-xs text-text-muted truncate">{secondary}</span>
        )}
      </span>
    </span>
  );
}
