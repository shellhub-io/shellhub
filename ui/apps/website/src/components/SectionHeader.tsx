import type { ReactNode } from "react";
import { Reveal } from "@shellhub/design-system/components";
import { cn } from "@shellhub/design-system/cn";

type Size = "section" | "sub" | "cta";
type Align = "center" | "left";
type EyebrowColor = "primary" | "cyan" | "green";
type Variant = "cta";

const sizeClasses: Record<Size, string> = {
  section: "text-[clamp(1.75rem,4vw,3rem)]",
  sub: "text-[clamp(1.75rem,4vw,2.5rem)]",
  cta: "text-[clamp(1.5rem,3vw,2.25rem)]",
};

const eyebrowColorClasses: Record<EyebrowColor, string> = {
  primary: "text-primary",
  cyan: "text-accent-cyan",
  green: "text-accent-green",
};

export interface SectionHeaderProps {
  title: ReactNode;
  eyebrow?: ReactNode;
  subtitle?: ReactNode;
  size?: Size;
  align?: Align;
  eyebrowColor?: EyebrowColor;
  variant?: Variant;
  reveal?: boolean;
  className?: string;
  subtitleClassName?: string;
}

export function SectionHeader({
  title,
  eyebrow,
  subtitle,
  size,
  align = "center",
  eyebrowColor = "primary",
  variant,
  reveal,
  className,
  subtitleClassName,
}: SectionHeaderProps) {
  const isCenter = align === "center";
  const isCta = variant === "cta";
  const resolvedSize = size ?? (isCta ? "cta" : "section");
  const resolvedReveal = reveal ?? !isCta;
  const defaultMargin = isCta ? "mb-0" : "mb-14";

  const content = (
    <div
      className={cn(
        isCenter && "text-center mx-auto",
        !resolvedReveal && cn(defaultMargin, className),
      )}
    >
      {eyebrow && (
        <p
          className={cn(
            "text-2xs font-mono font-semibold uppercase tracking-label mb-3",
            eyebrowColorClasses[eyebrowColor],
          )}
        >
          {eyebrow}
        </p>
      )}
      <h2
        className={cn(
          "font-bold tracking-[-0.03em] leading-tight text-text-primary",
          sizeClasses[resolvedSize],
        )}
      >
        {title}
      </h2>
      {subtitle && (
        <p
          className={cn(
            "mt-4 text-sm text-text-secondary leading-relaxed",
            isCenter && (isCta ? "max-w-md mx-auto mb-8" : "max-w-lg mx-auto"),
            subtitleClassName,
          )}
        >
          {subtitle}
        </p>
      )}
    </div>
  );

  if (!resolvedReveal) {
    return content;
  }

  return <Reveal className={cn(defaultMargin, className)}>{content}</Reveal>;
}
