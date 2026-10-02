import { useId, type ReactNode } from "react";
import { Link } from "react-router-dom";
import { ArrowRightIcon } from "@heroicons/react/24/outline";
import { Card } from "@shellhub/design-system/primitives";
import { cn } from "@shellhub/design-system/cn";
import { LABEL_BASE } from "@/utils/styles";
import { formatCount } from "@/utils/count";

/**
 * A titled dashboard card. The header has a fixed height so titles line up across columns whether
 * or not an action sits beside them.
 */
export function Panel({
  title,
  action,
  children,
}: {
  title: string;
  action?: ReactNode;
  children: ReactNode;
}) {
  const titleId = useId();

  return (
    <Card as="section" aria-labelledby={titleId} className="rounded-lg">
      <header className="flex items-center justify-between gap-3 h-12 px-5">
        <h2 id={titleId} className={LABEL_BASE}>
          {title}
        </h2>
        {action}
      </header>
      {children}
    </Card>
  );
}

/**
 * A panel's own failure, announced as an alert, so one failing request leaves the rest of the
 * page usable.
 */
export function PanelError({ children }: { children: ReactNode }) {
  return (
    <p role="alert" className="px-5 pb-4 text-xs text-accent-red">
      {children}
    </p>
  );
}

/**
 * A labelled value inside a panel. With to, the whole row is a link to the page behind the value.
 */
export function Row({
  label,
  to,
  children,
}: {
  label: string;
  to?: string;
  children: ReactNode;
}) {
  const body = (
    <>
      <span className="text-sm text-text-muted">{label}</span>
      <span className="flex items-center gap-2 text-sm font-medium text-text-primary min-w-0">
        {children}
        {to && (
          <ArrowRightIcon className="w-3.5 h-3.5 text-text-muted group-hover:text-primary transition-colors shrink-0" />
        )}
      </span>
    </>
  );
  const className =
    "flex items-center justify-between gap-4 px-5 py-3 border-t border-border";

  return to ? (
    <Link
      to={to}
      className={cn(className, "group hover:bg-hover-subtle transition-colors")}
    >
      {body}
    </Link>
  ) : (
    <div className={className}>{body}</div>
  );
}

/**
 * What a panel shows for a value that comes from a request: a dash once the request failed, even
 * if it still carries a value, an ellipsis until it answers, then children.
 */
export function Loaded({
  isLoading,
  isError,
  children,
}: {
  isLoading: boolean;
  isError: boolean;
  children: ReactNode;
}) {
  if (isError) return <>—</>;
  if (isLoading) return <>…</>;
  return <>{children}</>;
}

/**
 * A figure from a request, grouped for the reader's locale, under the same placeholders as Loaded.
 * A value still undefined after the request answered reads as loading.
 */
export function Count({
  value,
  isLoading,
  isError,
}: {
  value?: number;
  isLoading: boolean;
  isError: boolean;
}) {
  return (
    <span className="font-mono tabular-nums">
      <Loaded isLoading={isLoading || value === undefined} isError={isError}>
        {value !== undefined && formatCount(value)}
      </Loaded>
    </span>
  );
}
