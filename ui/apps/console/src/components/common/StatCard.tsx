import { ReactNode } from "react";
import { Link } from "react-router-dom";
import { Button, Card } from "@shellhub/design-system/primitives";
import { cn } from "@shellhub/design-system/cn";

type StatCardAction = { label: string } & (
  { to: string } | { onClick: () => void }
);

/**
 * StatCardProps describes one card, so a page can hold its cards as data and spread each one in.
 * Without action the card shows only the figure.
 */
export interface StatCardProps {
  icon: ReactNode;
  title: string;
  value: number | string;
  accent?: string;
  action?: StatCardAction;
}

/**
 * A dashboard figure with its label and, given an action, a way through to the list behind it.
 */
export default function StatCard({
  icon,
  title,
  value,
  accent,
  action,
}: StatCardProps) {
  return (
    <Card className="h-full rounded-lg p-6 flex flex-col items-center text-center group hover:border-primary/30 transition-all duration-300">
      <div
        aria-hidden="true"
        className="w-14 h-14 rounded-xl bg-primary/10 border border-primary/20 flex items-center justify-center text-primary mb-5"
      >
        {icon}
      </div>

      <p className="text-2xs font-mono font-medium uppercase tracking-label text-text-muted mb-2">
        {title}
      </p>

      <p
        className={cn(
          "text-4xl font-mono font-bold tabular-nums",
          action && "mb-5",
          accent ?? "text-text-primary",
        )}
      >
        {value}
      </p>

      {action && "onClick" in action && (
        <Button variant="ghost" size="sm" onClick={action.onClick}>
          {action.label} &rarr;
        </Button>
      )}
      {action && "to" in action && (
        <Button variant="ghost" size="sm" as={Link} to={action.to}>
          {action.label} &rarr;
        </Button>
      )}
    </Card>
  );
}
