import {
  ChatBubbleLeftRightIcon,
  ShieldCheckIcon,
} from "@heroicons/react/24/outline";
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
                  <ChatBubbleLeftRightIcon className="w-5 h-5 text-text-secondary" />
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
                    <ShieldCheckIcon className="w-5 h-5 text-primary" />
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
