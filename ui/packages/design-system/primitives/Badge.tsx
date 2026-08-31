import type { ReactNode } from "react";
import type { Palette } from "./IconBadge";
import { cn } from "./cn";

/**
 * The palette a Badge may take. Neutral is excluded because a badge on a neutral fill has no
 * contrast against the surface it sits on.
 */
export type BadgeColor = Exclude<Palette, "neutral">;
/**
 * Rounded reads as a label in running text; pill is bordered, monospaced and uppercased, for a
 * status or a count.
 */
export type BadgeShape = "rounded" | "pill";

/**
 * Props of Badge. Anything else given is spread onto the span, so a caller can attach data
 * attributes and ARIA without the component naming them.
 */
export interface BadgeProps {
  color?: BadgeColor;
  shape?: BadgeShape;
  className?: string;
  children?: ReactNode;
  [key: string]: unknown;
}

const roundedColorClasses: Record<BadgeColor, string> = {
  primary: "bg-primary/10 text-primary",
  green: "bg-accent-green/10 text-accent-green",
  red: "bg-accent-red/10 text-accent-red",
  yellow: "bg-accent-yellow/10 text-accent-yellow",
  blue: "bg-accent-blue/10 text-accent-blue",
  cyan: "bg-accent-cyan/10 text-accent-cyan",
};

const pillColorClasses: Record<BadgeColor, string> = {
  primary: "bg-primary/10 text-primary border-primary/20",
  green: "bg-accent-green/10 text-accent-green border-accent-green/20",
  red: "bg-accent-red/10 text-accent-red border-accent-red/20",
  yellow: "bg-accent-yellow/10 text-accent-yellow border-accent-yellow/20",
  blue: "bg-accent-blue/10 text-accent-blue border-accent-blue/20",
  cyan: "bg-accent-cyan/10 text-accent-cyan border-accent-cyan/20",
};

/**
 * Small inline label carrying a state or a category next to the thing it describes.
 */
export function Badge({
  color = "primary",
  shape = "rounded",
  className,
  children,
  ...rest
}: BadgeProps) {
  const isRounded = shape === "rounded";

  return (
    <span
      className={cn(
        "inline-flex items-center gap-1 text-2xs",
        isRounded
          ? cn("px-1.5 py-0.5 rounded font-medium", roundedColorClasses[color])
          : cn(
              "px-2 py-0.5 rounded-full font-mono font-semibold uppercase tracking-compact border",
              pillColorClasses[color],
            ),
        className,
      )}
      {...rest}
    >
      {children}
    </span>
  );
}
