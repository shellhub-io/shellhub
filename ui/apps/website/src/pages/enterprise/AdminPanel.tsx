import { CheckIcon } from "@heroicons/react/24/outline";
import { WindowChrome } from "@shellhub/design-system/primitives";
import { Section, SectionHeader } from "@/components";
import { Reveal, ShimmerCard } from "@shellhub/design-system/components";

const capabilities = [
  {
    label: "User management",
    desc: "Create, invite, and manage users across your organization",
  },
  {
    label: "Namespace administration",
    desc: "Organize devices into namespaces with fine-grained access",
  },
  {
    label: "Role-based access control",
    desc: "Assign roles and permissions at namespace and device level",
  },
  {
    label: "Billing & license management",
    desc: "View usage, manage subscriptions, and track licenses",
  },
  {
    label: "Global settings",
    desc: "Configure security policies, session limits, and defaults",
  },
  {
    label: "API key management",
    desc: "Create and revoke API keys for automation and integrations",
  },
];

export function AdminPanel() {
  return (
    <Section>
      <div className="grid lg:grid-cols-2 gap-12 items-center">
        <div>
          <SectionHeader
            align="left"
            size="sub"
            className="mb-8"
            eyebrow="Admin Panel"
            title="Manage everything from the browser"
            subtitle="No more CLI commands for admin tasks. The Enterprise admin panel gives your team a complete web interface for user management, namespace administration, and security policies."
          />

          <div className="space-y-3">
            {capabilities.map((cap, i) => (
              <Reveal key={i} delay={i * 0.04}>
                <div className="flex items-start gap-3">
                  <CheckIcon
                    className="w-4 h-4 text-accent-green shrink-0 mt-0.5"
                    strokeWidth={2}
                    aria-hidden="true"
                  />
                  <div>
                    <p className="text-sm font-medium text-text-primary">
                      {cap.label}
                    </p>
                    <p className="text-xs text-text-secondary leading-relaxed">
                      {cap.desc}
                    </p>
                  </div>
                </div>
              </Reveal>
            ))}
          </div>
        </div>

        <Reveal delay={0.1}>
          <ShimmerCard>
            <WindowChrome
              variant="browser"
              path="/admin"
              bodyClassName="font-sans"
            >
              <div className="space-y-3">
                <div className="flex items-center justify-between p-3 bg-surface rounded-lg border border-border">
                  <div className="flex items-center gap-3">
                    <div className="w-8 h-8 rounded-full bg-primary/15 flex items-center justify-center text-2xs font-semibold text-primary">
                      JD
                    </div>
                    <div>
                      <p className="text-xs font-medium">Jane Doe</p>
                      <p className="text-2xs text-text-muted">
                        jane@company.com
                      </p>
                    </div>
                  </div>
                  <span className="px-2 py-0.5 text-2xs font-mono bg-accent-green/10 text-accent-green border border-accent-green/20 rounded-full">
                    Admin
                  </span>
                </div>

                <div className="flex items-center justify-between p-3 bg-surface rounded-lg border border-border">
                  <div className="flex items-center gap-3">
                    <div className="w-8 h-8 rounded-full bg-accent-blue/15 flex items-center justify-center text-2xs font-semibold text-accent-blue">
                      MS
                    </div>
                    <div>
                      <p className="text-xs font-medium">Mike Smith</p>
                      <p className="text-2xs text-text-muted">
                        mike@company.com
                      </p>
                    </div>
                  </div>
                  <span className="px-2 py-0.5 text-2xs font-mono bg-primary/10 text-primary border border-primary/20 rounded-full">
                    Operator
                  </span>
                </div>

                <div className="flex items-center justify-between p-3 bg-surface rounded-lg border border-border">
                  <div className="flex items-center gap-3">
                    <div className="w-8 h-8 rounded-full bg-accent-cyan/15 flex items-center justify-center text-2xs font-semibold text-accent-cyan">
                      AL
                    </div>
                    <div>
                      <p className="text-xs font-medium">Ana Lima</p>
                      <p className="text-2xs text-text-muted">
                        ana@company.com
                      </p>
                    </div>
                  </div>
                  <span className="px-2 py-0.5 text-2xs font-mono bg-white/[0.04] text-text-muted border border-border rounded-full">
                    Observer
                  </span>
                </div>
              </div>

              <div className="mt-4 pt-4 border-t border-border flex items-center justify-between">
                <span className="text-2xs text-text-muted">
                  3 users in production namespace
                </span>
                <span className="text-2xs text-primary font-medium">
                  Manage &rarr;
                </span>
              </div>
            </WindowChrome>
          </ShimmerCard>
        </Reveal>
      </div>
    </Section>
  );
}
