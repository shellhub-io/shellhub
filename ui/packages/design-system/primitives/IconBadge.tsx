import type { ReactNode } from "react";
import { cn } from "./cn";

export type Palette =
  | "primary"
  | "green"
  | "red"
  | "yellow"
  | "blue"
  | "cyan"
  | "neutral";
export type IconBadgeSize = "sm" | "md" | "lg";

export interface IconBadgeProps {
  color?: Palette;
  size?: IconBadgeSize;
  className?: string;
  children?: ReactNode;
  [key: string]: unknown;
}

const colorClasses: Record<Palette, string> = {
  primary: "bg-primary/10 border-primary/20 text-primary",
  green: "bg-accent-green/10 border-accent-green/20 text-accent-green",
  red: "bg-accent-red/10 border-accent-red/20 text-accent-red",
  yellow: "bg-accent-yellow/10 border-accent-yellow/20 text-accent-yellow",
  blue: "bg-accent-blue/10 border-accent-blue/20 text-accent-blue",
  cyan: "bg-accent-cyan/10 border-accent-cyan/20 text-accent-cyan",
  neutral: "bg-white/[0.04] border-border text-text-secondary",
};

const sizeClasses: Record<IconBadgeSize, string> = {
  sm: "w-8 h-8",
  md: "w-10 h-10",
  lg: "w-12 h-12",
};

export function IconBadge({
  color = "primary",
  size = "md",
  className,
  children,
  ...rest
}: IconBadgeProps) {
  return (
    <div
      className={cn(
        "flex items-center justify-center border rounded-lg shrink-0",
        sizeClasses[size],
        colorClasses[color],
        className,
      )}
      {...rest}
    >
      {children}
    </div>
  );
}
