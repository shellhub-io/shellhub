import { Badge, Card, IconBadge } from "@shellhub/design-system/primitives";
import { Section, SectionHeader } from "@/components/marketing";
import { FeatureListItem } from "@/components/marketing/FeatureListItem";
import { Reveal, ShimmerCard } from "../landing/components";

export function DeploymentOptions() {
  return (
    <Section>
      <SectionHeader
        eyebrow="Deployment"
        title="Your infrastructure, your rules"
        subtitle="Choose the deployment model that fits your organization. Fully managed or on your own infrastructure."
      />

      <div className="grid md:grid-cols-2 gap-6">
        <Reveal delay={0}>
          <ShimmerCard className="h-full">
            <div className="relative bg-card border border-primary/30 rounded-xl p-8 flex flex-col h-full hover:border-primary/50 transition-all duration-300 shadow-[0_0_40px_rgba(102,122,204,0.1)] overflow-hidden">
              <div className="absolute inset-0 bg-gradient-to-br from-primary/[0.06] via-transparent to-transparent pointer-events-none" />
              <div className="relative">
                <div className="flex items-center gap-3 mb-4">
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
                        d="M2.25 15a4.5 4.5 0 0 0 4.5 4.5H18a3.75 3.75 0 0 0 1.332-7.257 3 3 0 0 0-3.758-3.848 5.25 5.25 0 0 0-10.233 2.33A4.502 4.502 0 0 0 2.25 15z"
                      />
                    </svg>
                  </IconBadge>
                  <Badge shape="pill" color="green">
                    Recommended
                  </Badge>
                </div>

                <h3 className="text-lg font-bold mb-2">Managed Cloud</h3>
                <p className="text-sm text-text-secondary leading-relaxed mb-6">
                  We handle the infrastructure. Dedicated servers, automatic
                  updates, and guaranteed uptime.
                </p>

                <ul className="space-y-2.5">
                  {[
                    "Dedicated servers for your organization",
                    "Automatic updates and patches",
                    "99.9% uptime SLA",
                    "Daily backups with point-in-time recovery",
                    "Global edge network for low latency",
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

        <Reveal delay={0.1}>
          <ShimmerCard className="h-full">
            <Card hover className="p-8 flex flex-col h-full">
              <div className="flex items-center gap-3 mb-4">
                <IconBadge color="neutral">
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
                      d="M21.75 17.25v-.228a4.5 4.5 0 0 0-.12-1.03l-2.268-9.64a3.375 3.375 0 0 0-3.285-2.602H7.923a3.375 3.375 0 0 0-3.285 2.602l-2.268 9.64a4.5 4.5 0 0 0-.12 1.03v.228m19.5 0a3 3 0 0 1-3 3H5.25a3 3 0 0 1-3-3m19.5 0a3 3 0 0 0-3-3H5.25a3 3 0 0 0-3 3m16.5 0h.008v.008h-.008v-.008Zm-3 0h.008v.008h-.008v-.008Z"
                    />
                  </svg>
                </IconBadge>
              </div>

              <h3 className="text-lg font-bold mb-2">On-Premises</h3>
              <p className="text-sm text-text-secondary leading-relaxed mb-6">
                Run ShellHub on your own infrastructure. Full data sovereignty
                and compliance control.
              </p>

              <ul className="space-y-2.5">
                {[
                  "Complete data sovereignty",
                  "Deploy on Kubernetes with Helm charts",
                  "Docker Compose for simpler setups",
                  "Air-gapped environment support",
                  "Custom integration with your toolchain",
                ].map((item) => (
                  <FeatureListItem key={item} color="muted">
                    {item}
                  </FeatureListItem>
                ))}
              </ul>
            </Card>
          </ShimmerCard>
        </Reveal>
      </div>
    </Section>
  );
}
