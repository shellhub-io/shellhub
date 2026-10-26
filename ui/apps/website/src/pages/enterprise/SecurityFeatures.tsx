import {
  ArrowRightEndOnRectangleIcon,
  LockClosedIcon,
  ClockIcon,
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
    icon: ClockIcon,
    color: C.green,
    title: "Session History",
    desc: "Every connection is on record: who connected, to which device, from where, and whether it was allowed.",
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
    title: "Access Policies",
    desc: "Default-deny rules that grant members, roles or API keys a set of devices, down to the unix login.",
  },
];

/**
 * Enterprise section listing the security and authentication features: SSO, MFA, recording,
 * session history and access policies.
 */
export function SecurityFeatures() {
  return (
    <Section bordered={false}>
      <SectionHeader
        eyebrow="Security & Authentication"
        title="Enterprise-grade security, built in"
        subtitle="Meet compliance requirements with SSO, MFA enforcement, session recording and a record of every connection."
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
