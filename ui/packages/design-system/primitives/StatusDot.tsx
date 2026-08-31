import { cn } from "./cn";

/**
 * Colour of an online dot. Ignored when offline, which is always muted.
 */
export type StatusDotColor = "green" | "primary" | "yellow";
/**
 * Diameter of a StatusDot: sm for a dense table row, md beside a heading.
 */
export type StatusDotSize = "sm" | "md";

/**
 * Props of StatusDot.
 */
export interface StatusDotProps {
  online?: boolean;
  color?: StatusDotColor;
  size?: StatusDotSize;
  className?: string;
}

const sizeClasses: Record<StatusDotSize, string> = {
  sm: "h-1.5 w-1.5",
  md: "h-2.5 w-2.5",
};

const glowClasses: Record<StatusDotColor, string> = {
  green: "shadow-[0_0_6px_rgba(130,165,104,0.4)]",
  primary: "shadow-[0_0_6px_rgba(102,122,204,0.4)]",
  yellow: "shadow-[0_0_6px_rgba(191,140,93,0.4)]",
};

const bgClasses: Record<StatusDotColor, string> = {
  green: "bg-accent-green",
  primary: "bg-primary",
  yellow: "bg-accent-yellow",
};

/**
 * Connectivity dot: a glowing pulse when online, a flat muted dot when not. Both carry a
 * role and a label, so the state reaches a reader that cannot see the colour.
 */
export function StatusDot({
  online = true,
  color = "green",
  size = "md",
  className,
}: StatusDotProps) {
  const sizeClass = sizeClasses[size];

  if (!online) {
    return (
      <span
        className={cn(
          "block rounded-full",
          sizeClass,
          "bg-text-muted/30",
          className,
        )}
        aria-label="Offline"
        role="img"
      />
    );
  }

  return (
    <span
      className={cn("relative flex", sizeClass, className)}
      aria-label="Online"
      role="img"
    >
      <span
        className={cn(
          "animate-ping absolute inline-flex h-full w-full rounded-full opacity-40",
          bgClasses[color],
        )}
      />
      <span
        className={cn(
          "relative inline-flex rounded-full h-full w-full",
          bgClasses[color],
          glowClasses[color],
        )}
      />
    </span>
  );
}
