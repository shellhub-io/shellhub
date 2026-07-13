import { ReactNode } from "react";
import { IconBadge, type Palette } from "@shellhub/design-system/primitives";
import { GlowOrbs } from "@shellhub/design-system/components";
import { cn } from "@shellhub/design-system/cn";

interface PageHeaderProps {
  icon: ReactNode;
  overline: string;
  title: string;
  description?: string;
  children?: ReactNode;
  variant?: "default" | "decorated";
  iconColor?: Palette;
}

export default function PageHeader({
  icon,
  overline,
  title,
  description,
  children,
  variant = "default",
  iconColor = "primary",
}: PageHeaderProps) {
  return (
    <div
      className={cn(
        "relative -mx-8 -mt-8 px-8 py-6 mb-8 border-b border-border",
        variant === "decorated"
          ? "animate-fade-in overflow-hidden"
          : "bg-surface page-header-band",
      )}
    >
      {variant === "decorated" && <GlowOrbs preset="corner" tone="primary" />}

      <div
        className={cn(variant === "decorated" && "relative", "flex flex-col sm:flex-row sm:items-center justify-between gap-4")}
      >
        <div className="flex items-start gap-4">
          <IconBadge size="lg" color={iconColor}>
            {icon}
          </IconBadge>
          <div>
            <p className="text-2xs font-mono font-semibold uppercase tracking-label text-primary mb-1">
              {overline}
            </p>
            <h1 className="text-xl font-semibold text-text-primary leading-tight">
              {title}
            </h1>
            {description && (
              <p className="text-sm text-text-muted mt-1 max-w-xl">
                {description}
              </p>
            )}
          </div>
        </div>
        {children && <div className="shrink-0">{children}</div>}
      </div>
    </div>
  );
}
