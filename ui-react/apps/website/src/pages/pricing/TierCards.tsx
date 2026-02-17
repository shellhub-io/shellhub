import { Reveal, ShimmerCard } from "../landing/components";

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
    ctaHref: "/v2/getting-started",
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
    ctaHref: "/v2/getting-started",
    highlighted: true,
  },
  {
    name: "Enterprise",
    badge: "Custom",
    badgeClass: "bg-accent-yellow/10 text-accent-yellow border-accent-yellow/20",
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
    ctaHref: "/v2/enterprise",
    highlighted: false,
  },
];

export function TierCards() {
  return (
    <section className="py-12">
      <div className="max-w-7xl mx-auto px-8">
        <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
          {tiers.map((tier, i) => (
            <Reveal key={tier.name} delay={i * 0.08}>
              <ShimmerCard className="h-full">
                <div
                  className={`relative rounded-xl p-8 flex flex-col h-full transition-all duration-300 overflow-hidden ${
                    tier.highlighted
                      ? "bg-card border border-primary/30 hover:border-primary/50 shadow-[0_0_40px_rgba(102,122,204,0.15)]"
                      : "bg-card border border-border hover:border-border-light"
                  }`}
                >
                  {tier.highlighted && (
                    <>
                      <div className="absolute inset-0 bg-gradient-to-br from-primary/[0.08] via-primary/[0.02] to-transparent pointer-events-none" />
                      <div className="absolute top-0 right-0 w-40 h-40 bg-primary/[0.08] rounded-full -translate-y-1/2 translate-x-1/2 blur-3xl pointer-events-none" />
                    </>
                  )}

                  <div className="relative">
                    <div className="flex items-center gap-3 mb-4">
                      <h3 className="text-lg font-bold">{tier.name}</h3>
                      <span className={`px-2 py-0.5 text-2xs font-mono font-semibold uppercase tracking-[0.1em] border rounded-full ${tier.badgeClass}`}>
                        {tier.badge}
                      </span>
                    </div>

                    <div className="mb-2">
                      <span className="text-3xl font-bold">{tier.price}</span>
                    </div>
                    <p className="text-2xs text-text-muted mb-4">{tier.priceSuffix}</p>
                    <p className="text-sm text-text-secondary leading-relaxed mb-6">{tier.desc}</p>

                    <ul className="space-y-2.5 mb-8 flex-1">
                      {tier.features.map((feature) => (
                        <li key={feature} className="flex items-center gap-2.5 text-sm text-text-secondary">
                          <svg className={`w-4 h-4 shrink-0 ${tier.highlighted ? "text-accent-green" : "text-text-muted"}`} fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
                            <path strokeLinecap="round" strokeLinejoin="round" d="m4.5 12.75 6 6 9-13.5" />
                          </svg>
                          {feature}
                        </li>
                      ))}
                    </ul>

                    <a
                      href={tier.ctaHref}
                      className={`inline-flex items-center justify-center gap-2 w-full px-6 py-3 text-sm font-semibold rounded-xl transition-all duration-300 group ${
                        tier.highlighted
                          ? "bg-primary border border-primary-400/40 text-white hover:brightness-110 hover:border-primary-400/60 hover:scale-[1.02] active:scale-[0.98]"
                          : "bg-surface border border-border text-text-primary hover:border-border-light hover:bg-white/[0.04] hover:scale-[1.02] active:scale-[0.98]"
                      }`}
                    >
                      {tier.cta}
                      <svg className="w-4 h-4 group-hover:translate-x-0.5 transition-transform duration-300" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2.5}>
                        <path strokeLinecap="round" strokeLinejoin="round" d="M13.5 4.5 21 12m0 0-7.5 7.5M21 12H3" />
                      </svg>
                    </a>
                  </div>
                </div>
              </ShimmerCard>
            </Reveal>
          ))}
        </div>
      </div>
    </section>
  );
}
