import { Reveal, ConnectionGrid } from "@shellhub/design-system/components";
import {
  Section,
  SectionHeader,
  ActionButtonGroup,
  type CTAAction,
  type SectionHeaderProps,
} from "@/components";

type GradientColor = "primary" | "accent-cyan" | "accent-green" | "accent-blue";

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

/**
 * Props of CTABanner. gradient names two palette entries rather than carrying classes, because
 * Tailwind only emits the utilities it can find spelled out in source.
 */
export interface CTABannerProps {
  eyebrow: string;
  title: string;
  subtitle: string;
  primaryAction: CTAAction;
  secondaryAction: CTAAction;
  eyebrowColor?: SectionHeaderProps["eyebrowColor"];
  gradient?: { from: GradientColor; to: GradientColor };
}

/**
 * The closing call-to-action panel of a page: eyebrow, heading, subtitle and a pair of actions
 * over a tinted gradient.
 */
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

          <div className="relative z-raised">
            <SectionHeader
              variant="cta"
              eyebrow={eyebrow}
              eyebrowColor={eyebrowColor}
              title={title}
              subtitle={subtitle}
            />

            <ActionButtonGroup
              primaryAction={primaryAction}
              secondaryAction={secondaryAction}
            />
          </div>
        </div>
      </Reveal>
    </Section>
  );
}
