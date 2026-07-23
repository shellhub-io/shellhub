import type { ComponentType, ReactNode, SVGProps } from "react";
import { Card } from "@shellhub/design-system/primitives";
import { Reveal, ShimmerCard } from "@shellhub/design-system/components";

export interface InfoCardProps {
  color: string;
  title: string;
  description: string;
  icon?: ComponentType<SVGProps<SVGSVGElement>>;
  layout?: "vertical" | "horizontal" | "dot";
  delay?: number;
  children?: ReactNode;
}

export function InfoCard({
  color,
  title,
  description,
  icon: Icon,
  layout = "vertical",
  delay = 0,
  children,
}: InfoCardProps) {
  const marker =
    layout === "dot" ? (
      <div
        className="w-2 h-2 rounded-full mb-4"
        style={{ background: color }}
      />
    ) : (
      <div
        className={`w-10 h-10 rounded-lg flex items-center justify-center border ${layout === "horizontal" ? "shrink-0" : "mb-4"}`}
        style={{
          background: `${color}15`,
          borderColor: `${color}25`,
        }}
      >
        {Icon && <Icon className="w-5 h-5" style={{ color }} />}
      </div>
    );

  const text = (
    <>
      <h4 className="text-sm font-semibold mb-2">{title}</h4>
      <p className="text-xs text-text-secondary leading-relaxed">
        {description}
      </p>
      {children}
    </>
  );

  return (
    <Reveal delay={delay}>
      <ShimmerCard className="h-full">
        <Card hover className="p-6 h-full">
          {layout === "horizontal" ? (
            <div className="flex items-start gap-4">
              {marker}
              <div>{text}</div>
            </div>
          ) : (
            <>
              {marker}
              {text}
            </>
          )}
        </Card>
      </ShimmerCard>
    </Reveal>
  );
}
