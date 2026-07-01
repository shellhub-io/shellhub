import {
  CommandLineIcon,
  DocumentTextIcon,
  LockClosedIcon,
  PencilSquareIcon,
  UsersIcon,
  ServerIcon,
} from "@heroicons/react/24/outline";
import { Card } from "@shellhub/design-system/primitives";
import { Section, SectionHeader } from "@/components/marketing";
import { Reveal, ShimmerCard } from "./components";
import { C } from "./constants";

export function FeatureGrid() {
  return (
    <Section bordered={false}>
      <SectionHeader
        eyebrow="Features"
        title="Everything you need to manage remote devices."
      />

      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
        {[
          {
            icon: (
              <CommandLineIcon
                width="20"
                height="20"
                stroke={C.primary}
                strokeWidth="1.5"
              />
            ),
            color: C.primary,
            title: "Native SSH Support",
            desc: "Use your standard SSH client. No proprietary tools or plugins required.",
            delay: 0,
          },
          {
            icon: (
              <DocumentTextIcon
                width="20"
                height="20"
                stroke={C.cyan}
                strokeWidth="1.5"
              />
            ),
            color: C.cyan,
            title: "SCP/SFTP File Transfer",
            desc: "Transfer files to and from remote devices with SCP and SFTP.",
            delay: 0.06,
          },
          {
            icon: (
              <LockClosedIcon
                width="20"
                height="20"
                stroke={C.yellow}
                strokeWidth="1.5"
              />
            ),
            color: C.yellow,
            title: "Multi-Factor Auth",
            desc: "Require TOTP-based MFA for SSH connections. Works with any authenticator app.",
            delay: 0.12,
          },
          {
            icon: (
              <PencilSquareIcon
                width="20"
                height="20"
                stroke={C.green}
                strokeWidth="1.5"
              />
            ),
            color: C.green,
            title: "Audit Logging",
            desc: "Full audit trail of every connection, command, and session. Export logs for compliance.",
            delay: 0,
          },
          {
            icon: (
              <UsersIcon
                width="20"
                height="20"
                stroke={C.primary}
                strokeWidth="1.5"
              />
            ),
            color: C.primary,
            title: "Team Management & RBAC",
            desc: "Invite team members, assign roles, and control who can access which devices.",
            delay: 0.06,
          },
          {
            icon: (
              <ServerIcon
                width="20"
                height="20"
                stroke={C.blue}
                strokeWidth="1.5"
              />
            ),
            color: C.blue,
            title: "Docker Container Access",
            desc: "SSH directly into Docker containers running on remote hosts. No docker exec needed.",
            delay: 0.12,
          },
        ].map((f, i) => (
          <Reveal key={i} delay={f.delay}>
            <ShimmerCard>
              <Card hover className="p-6 h-full">
                <div
                  className="w-10 h-10 rounded-lg flex items-center justify-center mb-4 border"
                  style={{
                    background: `${f.color}15`,
                    borderColor: `${f.color}25`,
                  }}
                >
                  {f.icon}
                </div>
                <h4 className="text-sm font-semibold mb-2 group-hover:text-primary transition-colors">
                  {f.title}
                </h4>
                <p className="text-xs text-text-secondary leading-relaxed">
                  {f.desc}
                </p>
              </Card>
            </ShimmerCard>
          </Reveal>
        ))}
      </div>
    </Section>
  );
}
