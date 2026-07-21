import {
  CommandLineIcon,
  DocumentTextIcon,
  LockClosedIcon,
  PencilSquareIcon,
  UsersIcon,
  ServerIcon,
} from "@heroicons/react/24/outline";
import { Section, SectionHeader, InfoCard } from "@/components/marketing";
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
          { icon: CommandLineIcon, color: C.primary, title: "Native SSH Support", desc: "Use your standard SSH client. No proprietary tools or plugins required.", delay: 0 },
          { icon: DocumentTextIcon, color: C.cyan, title: "SCP/SFTP File Transfer", desc: "Transfer files to and from remote devices with SCP and SFTP.", delay: 0.06 },
          { icon: LockClosedIcon, color: C.yellow, title: "Multi-Factor Auth", desc: "Require TOTP-based MFA for SSH connections. Works with any authenticator app.", delay: 0.12 },
          { icon: PencilSquareIcon, color: C.green, title: "Audit Logging", desc: "Full audit trail of every connection, command, and session. Export logs for compliance.", delay: 0 },
          { icon: UsersIcon, color: C.primary, title: "Team Management & RBAC", desc: "Invite team members, assign roles, and control who can access which devices.", delay: 0.06 },
          { icon: ServerIcon, color: C.blue, title: "Docker Container Access", desc: "SSH directly into Docker containers running on remote hosts. No docker exec needed.", delay: 0.12 },
        ].map((f, i) => (
          <InfoCard
            key={i}
            icon={f.icon}
            color={f.color}
            title={f.title}
            description={f.desc}
            delay={f.delay}
          />
        ))}
      </div>
    </Section>
  );
}
