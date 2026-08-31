import {
  ArrowRightEndOnRectangleIcon,
  LockClosedIcon,
  PencilIcon,
  ShieldCheckIcon,
} from "@heroicons/react/24/outline";
import { PlayIcon } from "@heroicons/react/24/solid";
import { Section, SectionHeader, InfoCard } from "@/components";
import { C } from "@shellhub/design-system/constants";

const features = [
  {
    icon: ArrowRightEndOnRectangleIcon,
    color: C.primary,
    title: "SSO / SAML",
    desc: "Single sign-on with your identity provider. Support for SAML 2.0 and OpenID Connect.",
  },
  {
    icon: LockClosedIcon,
    color: C.yellow,
    title: "Multi-Factor Authentication",
    desc: "Users can enable TOTP-based multi-factor authentication for web console login.",
  },
  {
    icon: PencilIcon,
    color: C.green,
    title: "Audit Logs",
    desc: "Complete audit trail of every action. User logins, device connections, configuration changes.",
  },
  {
    icon: PlayIcon,
    color: C.red,
    title: "Session Recording",
    desc: "Record and replay SSH sessions for compliance, training, and incident investigation.",
  },
  {
    icon: ShieldCheckIcon,
    color: C.blue,
    title: "Firewall Rules",
    desc: "Define granular connection policies per hostname or username pattern.",
  },
];

/**
 * Enterprise section listing the security and authentication features — SSO, MFA, recording
 * and audit logging.
 */
export function SecurityFeatures() {
  return (
    <Section bordered={false}>
      <SectionHeader
        eyebrow="Security & Authentication"
        title="Enterprise-grade security, built in"
        subtitle="Meet compliance requirements with SSO, MFA enforcement, session recording, and complete audit logging."
      />

      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
        {features.map((f, i) => (
          <InfoCard
            key={i}
            icon={f.icon}
            color={f.color}
            title={f.title}
            description={f.desc}
            delay={i * 0.04}
          />
        ))}
      </div>
    </Section>
  );
}
