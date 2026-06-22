import { Card } from "@shellhub/design-system/primitives";
import { Section, SectionHeader } from "@/components/marketing";
import { Reveal, ShimmerCard } from "../landing/components";
import { C } from "../landing/constants";

const features = [
  {
    icon: (
      <svg
        width="20"
        height="20"
        viewBox="0 0 24 24"
        fill="none"
        stroke={C.primary}
        strokeWidth="1.5"
        strokeLinecap="round"
      >
        <path d="M15 3h4a2 2 0 0 1 2 2v14a2 2 0 0 1-2 2h-4" />
        <polyline points="10 17 15 12 10 7" />
        <line x1="15" y1="12" x2="3" y2="12" />
      </svg>
    ),
    color: C.primary,
    title: "SSO / SAML",
    desc: "Single sign-on with your identity provider. Support for SAML 2.0 and OpenID Connect.",
  },
  {
    icon: (
      <svg
        width="20"
        height="20"
        viewBox="0 0 24 24"
        fill="none"
        stroke={C.cyan}
        strokeWidth="1.5"
        strokeLinecap="round"
      >
        <path d="M17 21v-2a4 4 0 0 0-4-4H5a4 4 0 0 0-4 4v2" />
        <circle cx="9" cy="7" r="4" />
        <path d="M23 21v-2a4 4 0 0 0-3-3.87" />
        <path d="M16 3.13a4 4 0 0 1 0 7.75" />
      </svg>
    ),
    color: C.cyan,
    title: "LDAP / Active Directory",
    desc: "Integrate with your existing directory service for centralized user management.",
  },
  {
    icon: (
      <svg
        width="20"
        height="20"
        viewBox="0 0 24 24"
        fill="none"
        stroke={C.yellow}
        strokeWidth="1.5"
        strokeLinecap="round"
      >
        <rect x="3" y="11" width="18" height="11" rx="2" />
        <path d="M7 11V7a5 5 0 0 1 10 0v4" />
        <circle cx="12" cy="16" r="1" />
      </svg>
    ),
    color: C.yellow,
    title: "MFA Enforcement",
    desc: "Require multi-factor authentication for all users or specific roles across the organization.",
  },
  {
    icon: (
      <svg
        width="20"
        height="20"
        viewBox="0 0 24 24"
        fill="none"
        stroke={C.green}
        strokeWidth="1.5"
        strokeLinecap="round"
      >
        <path d="M12 20h9" />
        <path d="M16.5 3.5a2.121 2.121 0 0 1 3 3L7 19l-4 1 1-4L16.5 3.5z" />
      </svg>
    ),
    color: C.green,
    title: "Audit Logs",
    desc: "Complete audit trail of every action. User logins, device connections, configuration changes.",
  },
  {
    icon: (
      <svg
        width="20"
        height="20"
        viewBox="0 0 24 24"
        fill="none"
        stroke={C.red}
        strokeWidth="1.5"
        strokeLinecap="round"
      >
        <circle cx="12" cy="12" r="10" />
        <polygon points="10 8 16 12 10 16 10 8" />
      </svg>
    ),
    color: C.red,
    title: "Session Recording",
    desc: "Record and replay SSH sessions for compliance, training, and incident investigation.",
  },
  {
    icon: (
      <svg
        width="20"
        height="20"
        viewBox="0 0 24 24"
        fill="none"
        stroke={C.blue}
        strokeWidth="1.5"
        strokeLinecap="round"
      >
        <path d="M12 22s8-4 8-10V5l-8-3-8 3v7c0 6 8 10 8 10z" />
      </svg>
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
