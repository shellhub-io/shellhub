import type { ReactNode } from "react";
import { CheckIcon } from "@heroicons/react/24/outline";
import { cn } from "@shellhub/design-system/cn";

type Color = "muted" | "green";

const colorClasses: Record<Color, string> = {
  muted: "text-text-muted",
  green: "text-accent-green",
};

export interface FeatureListItemProps {
  children: ReactNode;
  color?: Color;
  className?: string;
}

export function FeatureListItem({
  children,
  color = "muted",
  className,
}: FeatureListItemProps) {
  return (
    <li
      className={cn(
        "flex gap-2.5 text-sm text-text-secondary items-center",
        className,
      )}
    >
      <CheckIcon
        strokeWidth={2}
        className={cn("w-4 h-4 shrink-0", colorClasses[color])}
      />
      {children}
    </li>
  );
}
