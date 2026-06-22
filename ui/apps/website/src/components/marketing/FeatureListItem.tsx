import type { ReactNode } from "react";
import { cn } from "@shellhub/design-system/cn";

type Color = "muted" | "green";

const colorClasses: Record<Color, string> = {
  muted: "text-text-muted",
  green: "text-accent-green",
};

const DefaultIcon = ({ className }: { className?: string }) => (
  <svg
    xmlns="http://www.w3.org/2000/svg"
    fill="none"
    viewBox="0 0 24 24"
    strokeWidth={2}
    stroke="currentColor"
    className={cn("w-4 h-4 shrink-0", className)}
  >
    <path
      strokeLinecap="round"
      strokeLinejoin="round"
      d="m4.5 12.75 6 6 9-13.5"
    />
  </svg>
);

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
      <DefaultIcon className={colorClasses[color]} />
      {children}
    </li>
  );
}
