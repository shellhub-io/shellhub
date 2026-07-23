import { Link } from "react-router-dom";
import { cn } from "@shellhub/design-system/cn";
import { Button } from "@shellhub/design-system/primitives";
import {
  GlowOrbs,
  Reveal,
  ShimmerCard,
} from "@shellhub/design-system/components";
import { FeatureListItem, HighlightCard, Section } from "@/components";

const tiers = [
  {
    name: "Community",
    badge: "Open Source",
    badgeClass: "bg-white/[0.03] text-text-muted border-border",
    price: "Free",
    priceSuffix: "forever",
    desc: "For individuals and small teams getting started with remote device management.",
    features: [
      "Up to 3 devices",
      "SSH & web terminal",
      "SCP/SFTP file transfer",
      "Multi-factor authentication",
      "Firewall rules",
      "Docker container access",
      "Community support",
    ],
    cta: "Get Started",
    ctaHref: "/getting-started",
    highlighted: false,
  },
  {
    name: "Cloud",
    badge: "Popular",
    badgeClass: "bg-accent-green/10 text-accent-green border-accent-green/20",
    price: "$2",
    priceSuffix: "per device / month",
    desc: "Managed hosting with zero maintenance. Free for up to 3 devices.",
    features: [
      "Unlimited devices",
      "Everything in Community",
      "Managed infrastructure",
      "Automatic updates",
      "Session recording",
      "Audit logs",
      "Email support",
    ],
    cta: "Start Free",
    ctaHref: "/getting-started",
    highlighted: true,
  },
  {
    name: "Enterprise",
    badge: "Custom",
    badgeClass:
      "bg-accent-yellow/10 text-accent-yellow border-accent-yellow/20",
    price: "Custom",
    priceSuffix: "contact sales",
    desc: "For organizations that need full control, compliance, and dedicated support.",
    features: [
      "Everything in Cloud",
      "Admin panel",
      "SSO / SAML / LDAP",
      "MFA enforcement",
      "Dedicated support & SLA",
      "On-premises deployment",
      "Custom integrations",
    ],
    cta: "Contact Sales",
    ctaHref: "/enterprise",
    highlighted: false,
  },
];

export function TierCards() {
  return (
    <Section bordered={false} padding="md">
      <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
        {tiers.map((tier, i) => {
          const content = (
            <div className="relative">
              <div className="flex items-center gap-3 mb-4">
                <h3 className="text-lg font-bold">{tier.name}</h3>
                <span
                  className={cn(
                    "px-2 py-0.5 text-2xs font-mono font-semibold uppercase tracking-[0.1em] border rounded-full",
                    tier.badgeClass,
                  )}
                >
                  {tier.badge}
                </span>
              </div>

              <div className="mb-2">
                <span className="text-3xl font-bold">{tier.price}</span>
              </div>
              <p className="text-2xs text-text-muted mb-4">
                {tier.priceSuffix}
              </p>
              <p className="text-sm text-text-secondary leading-relaxed mb-6">
                {tier.desc}
              </p>

              <ul className="space-y-2.5 mb-8 flex-1">
                {tier.features.map((feature) => (
                  <FeatureListItem
                    key={feature}
                    color={tier.highlighted ? "green" : "muted"}
                  >
                    {feature}
                  </FeatureListItem>
                ))}
              </ul>

              <Button
                as={Link}
                to={tier.ctaHref}
                variant={tier.highlighted ? "primary" : "surface"}
                size="lg"
                glow={tier.highlighted}
                fullWidth
                className={
                  tier.highlighted
                    ? undefined
                    : "hover:scale-[1.02] active:scale-[0.98]"
                }
              >
                {tier.cta}
              </Button>
            </div>
          );

          return (
            <Reveal key={tier.name} delay={i * 0.08}>
              <ShimmerCard className="h-full">
                {tier.highlighted ? (
                  <HighlightCard
                    color="primary"
                    className="p-8 flex flex-col h-full"
                  >
                    <GlowOrbs preset="corner" tone="primary" />
                    {content}
                  </HighlightCard>
                ) : (
                  <div className="relative rounded-xl p-8 flex flex-col h-full transition-all duration-300 overflow-hidden bg-card border border-border hover:border-border-light">
                    {content}
                  </div>
                )}
              </ShimmerCard>
            </Reveal>
          );
        })}
      </div>
    </Section>
  );
}
