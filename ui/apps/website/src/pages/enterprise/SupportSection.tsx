import { Card, IconBadge } from "@shellhub/design-system/primitives";
import { Section, SectionHeader } from "@/components/marketing";
import { FeatureListItem } from "@/components/marketing/FeatureListItem";
import { Reveal, ShimmerCard } from "../landing/components";

export function SupportSection() {
  return (
    <Section bordered={false}>
      <SectionHeader
        eyebrow="Support"
        title="Support that matches your needs"
        subtitle="From community forums to dedicated account managers, choose the level of support your team needs."
      />

      <div className="grid md:grid-cols-2 gap-6">
        <Reveal delay={0}>
          <ShimmerCard className="h-full">
            <Card hover className="p-8 h-full">
              <div className="flex items-center gap-3 mb-6">
                <div className="w-10 h-10 rounded-lg bg-white/[0.04] border border-border flex items-center justify-center">
                  <svg
                    className="w-5 h-5 text-text-secondary"
                    fill="none"
                    viewBox="0 0 24 24"
                    stroke="currentColor"
                    strokeWidth={1.5}
                  >
                    <path
                      strokeLinecap="round"
                      strokeLinejoin="round"
                      d="M20.25 8.511c.884.284 1.5 1.128 1.5 2.097v4.286c0 1.136-.847 2.1-1.98 2.193-.34.027-.68.052-1.02.072v3.091l-3-3c-1.354 0-2.694-.055-4.02-.163a2.115 2.115 0 0 1-.825-.242m9.345-8.334a2.126 2.126 0 0 0-.476-.095 48.64 48.64 0 0 0-8.048 0c-1.131.094-1.976 1.057-1.976 2.192v4.286c0 .837.46 1.58 1.155 1.951m9.345-8.334V6.637c0-1.621-1.152-3.026-2.76-3.235A48.455 48.455 0 0 0 11.25 3c-2.115 0-4.198.137-6.24.402-1.608.209-2.76 1.614-2.76 3.235v6.226c0 1.621 1.152 3.026 2.76 3.235.577.075 1.157.14 1.74.194V21l4.155-4.155"
                    />
                  </svg>
                </div>
                <div>
                  <h3 className="text-sm font-bold">Community</h3>
                  <p className="text-2xs text-text-muted">Free & Open Source</p>
                </div>
              </div>

              <ul className="space-y-3">
                {[
                  "GitHub Issues & Discussions",
                  "Community Discord server",
                  "Public documentation",
                  "Community-driven bug fixes",
                ].map((item) => (
                  <FeatureListItem key={item} color="muted">
                    {item}
                  </FeatureListItem>
                ))}
              </ul>
            </Card>
          </ShimmerCard>
        </Reveal>

        <Reveal delay={0.1}>
          <ShimmerCard className="h-full">
            <div className="relative bg-card border border-primary/30 rounded-xl p-8 h-full hover:border-primary/50 transition-all duration-300 shadow-[0_0_40px_rgba(102,122,204,0.1)] overflow-hidden">
              <div className="absolute inset-0 bg-gradient-to-br from-primary/[0.06] via-transparent to-transparent pointer-events-none" />
              <div className="relative">
                <div className="flex items-center gap-3 mb-6">
                  <IconBadge color="primary">
                    <svg
                      className="w-5 h-5 text-primary"
                      fill="none"
                      viewBox="0 0 24 24"
                      stroke="currentColor"
                      strokeWidth={1.5}
                    >
                      <path
                        strokeLinecap="round"
                        strokeLinejoin="round"
                        d="M9 12.75 11.25 15 15 9.75m-3-7.036A11.959 11.959 0 0 1 3.598 6 11.99 11.99 0 0 0 3 9.749c0 5.592 3.824 10.29 9 11.623 5.176-1.332 9-6.03 9-11.622 0-1.31-.21-2.571-.598-3.751h-.152c-3.196 0-6.1-1.248-8.25-3.285z"
                      />
                    </svg>
                  </IconBadge>
                  <div>
                    <h3 className="text-sm font-bold">Enterprise</h3>
                    <p className="text-2xs text-primary">Priority Support</p>
                  </div>
                </div>

                <ul className="space-y-3">
                  {[
                    "Dedicated account manager",
                    "Priority ticket queue with SLA",
                    "Private Slack or Teams channel",
                    "Onboarding & migration assistance",
                    "Custom integration support",
                    "Quarterly business reviews",
                  ].map((item) => (
                    <FeatureListItem key={item} color="green">
                      {item}
                    </FeatureListItem>
                  ))}
                </ul>
              </div>
            </div>
          </ShimmerCard>
        </Reveal>
      </div>
    </Section>
  );
}
