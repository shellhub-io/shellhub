import { cn } from "@shellhub/design-system/cn";

type HighlightCardColor =
  | "primary"
  | "accent-cyan"
  | "accent-green"
  | "accent-blue"
  | "accent-yellow";

interface HighlightCardProps extends React.ComponentPropsWithoutRef<"div"> {
  color: HighlightCardColor;
}

const colorStyles: Record<
  HighlightCardColor,
  { card: string; gradient: string }
> = {
  primary: {
    card: "border-primary/30 hover:border-primary/50 shadow-[0_0_40px_rgba(102,122,204,0.1)]",
    gradient: "from-primary/[0.06]",
  },
  "accent-cyan": {
    card: "border-accent-cyan/30 hover:border-accent-cyan/50 shadow-[0_0_40px_rgba(78,154,163,0.08)]",
    gradient: "from-accent-cyan/[0.06]",
  },
  "accent-green": {
    card: "border-accent-green/30 hover:border-accent-green/50 shadow-[0_0_40px_rgba(130,165,104,0.08)]",
    gradient: "from-accent-green/[0.06]",
  },
  "accent-blue": {
    card: "border-accent-blue/30 hover:border-accent-blue/50 shadow-[0_0_40px_rgba(86,162,225,0.06)]",
    gradient: "from-accent-blue/[0.06]",
  },
  "accent-yellow": {
    card: "border-accent-yellow/30 hover:border-accent-yellow/50 shadow-[0_0_40px_rgba(191,140,93,0.06)]",
    gradient: "from-accent-yellow/[0.06]",
  },
};

export function HighlightCard({
  color,
  className,
  children,
  ...props
}: HighlightCardProps) {
  const styles = colorStyles[color];

  return (
    <div
      className={cn(
        "relative bg-card border rounded-xl overflow-hidden transition-all duration-300",
        styles.card,
        className,
      )}
      {...props}
    >
      <div
        className={cn(
          "absolute inset-0 bg-gradient-to-br via-transparent to-transparent pointer-events-none",
          styles.gradient,
        )}
      />
      {children}
    </div>
  );
}

export type { HighlightCardColor, HighlightCardProps };
