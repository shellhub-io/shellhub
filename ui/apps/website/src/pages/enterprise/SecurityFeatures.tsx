import {
  ArrowRightEndOnRectangleIcon,
  UsersIcon,
  LockClosedIcon,
  PencilIcon,
  ShieldCheckIcon,
} from "@heroicons/react/24/outline";
import { PlayIcon } from "@heroicons/react/24/solid";
import { Card } from "@shellhub/design-system/primitives";
import { Section, SectionHeader } from "@/components/marketing";
import { Reveal, ShimmerCard } from "../landing/components";
import { C } from "../landing/constants";

const features = [
  {
    icon: (
      <ArrowRightEndOnRectangleIcon
        width="20"
        height="20"
        stroke={C.primary}
        strokeWidth="1.5"
        aria-hidden="true"
      />
    ),
    color: C.primary,
    title: "SSO / SAML",
    desc: "Single sign-on with your identity provider. Support for SAML 2.0 and OpenID Connect.",
  },
  {
    icon: (
      <UsersIcon
        width="20"
        height="20"
        stroke={C.cyan}
        strokeWidth="1.5"
        aria-hidden="true"
      />
    ),
    color: C.cyan,
    title: "LDAP / Active Directory",
    desc: "Integrate with your existing directory service for centralized user management.",
  },
  {
    icon: (
      <LockClosedIcon
        width="20"
        height="20"
        stroke={C.yellow}
        strokeWidth="1.5"
        aria-hidden="true"
      />
    ),
    color: C.yellow,
    title: "MFA Enforcement",
    desc: "Require multi-factor authentication for all users or specific roles across the organization.",
  },
  {
    icon: (
      <PencilIcon
        width="20"
        height="20"
        stroke={C.green}
        strokeWidth="1.5"
        aria-hidden="true"
      />
    ),
    color: C.green,
    title: "Audit Logs",
    desc: "Complete audit trail of every action. User logins, device connections, configuration changes.",
  },
  {
    icon: (
      <PlayIcon
        width="20"
        height="20"
        style={{ color: C.red }}
        aria-hidden="true"
      />
    ),
    color: C.red,
    title: "Session Recording",
    desc: "Record and replay SSH sessions for compliance, training, and incident investigation.",
  },
  {
    icon: (
      <ShieldCheckIcon
        width="20"
        height="20"
        stroke={C.blue}
        strokeWidth="1.5"
        aria-hidden="true"
      />
    ),
    color: C.blue,
    title: "Firewall Rules",
    desc: "Define granular connection policies per device, namespace, or user group.",
  },
];

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
          <Reveal key={i} delay={i * 0.04}>
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
                <h4 className="text-sm font-semibold mb-2">{f.title}</h4>
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
