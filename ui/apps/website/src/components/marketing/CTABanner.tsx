import { Link } from "react-router-dom";
import { Button } from "@shellhub/design-system/primitives";
import { Reveal, ConnectionGrid } from "@shellhub/design-system/components";
import { ArrowRight } from "@/components/ArrowRight";
import { Section, SectionHeader } from "@/components/marketing";
import type { SectionHeaderProps } from "@/components/marketing";

type GradientColor = "primary" | "accent-cyan" | "accent-green" | "accent-blue";

type CTAAction =
  | { label: string; to: string; href?: never; external?: never }
  | { label: string; href: string; to?: never; external?: boolean };

const GRADIENT_FROM: Record<GradientColor, string> = {
  primary: "from-primary/[0.06]",
  "accent-cyan": "from-accent-cyan/[0.06]",
  "accent-green": "from-accent-green/[0.06]",
  "accent-blue": "from-accent-blue/[0.06]",
};

const GRADIENT_TO: Record<GradientColor, string> = {
  primary: "to-primary/[0.04]",
  "accent-cyan": "to-accent-cyan/[0.04]",
  "accent-green": "to-accent-green/[0.04]",
  "accent-blue": "to-accent-blue/[0.04]",
};

export interface CTABannerProps {
  eyebrow: string;
  title: string;
  subtitle: string;
  primaryAction: CTAAction;
  secondaryAction: CTAAction;
  eyebrowColor?: SectionHeaderProps["eyebrowColor"];
  gradient?: { from: GradientColor; to: GradientColor };
}

function ActionButton({
  action,
  variant,
}: {
  action: CTAAction;
  variant: "primary" | "outline";
}) {
  const isPrimary = variant === "primary";
  const shared = {
    variant,
    size: "xl" as const,
    glow: isPrimary || undefined,
    iconRight: isPrimary ? <ArrowRight /> : undefined,
    children: action.label,
  };

  if (action.to) {
    return <Button as={Link} to={action.to} {...shared} />;
  }

  return (
    <Button
      as="a"
      href={action.href}
      {...shared}
      {...(action.external && {
        target: "_blank",
        rel: "noopener noreferrer",
      })}
    />
  );
}

export function CTABanner({
  eyebrow,
  title,
  subtitle,
  primaryAction,
  secondaryAction,
  eyebrowColor,
  gradient = { from: "primary", to: "accent-cyan" },
}: CTABannerProps) {
  return (
    <Section>
      <Reveal>
        <div className="relative bg-card border border-border rounded-2xl p-12 text-center overflow-hidden">
          <ConnectionGrid />
          <div
            className={`absolute inset-0 bg-gradient-to-br ${GRADIENT_FROM[gradient.from]} via-transparent ${GRADIENT_TO[gradient.to]} pointer-events-none`}
          />

          <div className="relative z-10">
            <SectionHeader
              variant="cta"
              eyebrow={eyebrow}
              eyebrowColor={eyebrowColor}
              title={title}
              subtitle={subtitle}
            />

            <div className="flex flex-col sm:flex-row items-center justify-center gap-3">
              <ActionButton action={primaryAction} variant="primary" />
              <ActionButton action={secondaryAction} variant="outline" />
            </div>
          </div>
        </div>
      </Reveal>
    </Section>
  );
}
