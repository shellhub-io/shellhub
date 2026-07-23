import { Fragment } from "react";
import { CheckIcon, XMarkIcon } from "@heroicons/react/24/outline";
import { Section, SectionHeader } from "@/components";
import { Reveal } from "@shellhub/design-system/components";

type FeatureValue = boolean | string;

interface Feature {
  name: string;
  community: FeatureValue;
  cloud: FeatureValue;
  enterprise: FeatureValue;
}

interface Category {
  label: string;
  features: Feature[];
}

const categories: Category[] = [
  {
    label: "Core",
    features: [
      { name: "SSH access", community: true, cloud: true, enterprise: true },
      { name: "Web terminal", community: true, cloud: true, enterprise: true },
      {
        name: "SCP/SFTP file transfer",
        community: true,
        cloud: true,
        enterprise: true,
      },
      {
        name: "Docker container access",
        community: true,
        cloud: true,
        enterprise: true,
      },
      {
        name: "Device tagging",
        community: true,
        cloud: true,
        enterprise: true,
      },
    ],
  },
  {
    label: "Security",
    features: [
      {
        name: "Multi-factor authentication",
        community: true,
        cloud: true,
        enterprise: true,
      },
      {
        name: "Firewall rules",
        community: true,
        cloud: true,
        enterprise: true,
      },
      {
        name: "Session recording",
        community: false,
        cloud: true,
        enterprise: true,
      },
      { name: "Audit logs", community: false, cloud: true, enterprise: true },
      {
        name: "MFA enforcement",
        community: false,
        cloud: false,
        enterprise: true,
      },
      { name: "SSO / SAML", community: false, cloud: false, enterprise: true },
      {
        name: "LDAP / Active Directory",
        community: false,
        cloud: false,
        enterprise: true,
      },
    ],
  },
  {
    label: "Management",
    features: [
      { name: "Namespaces", community: true, cloud: true, enterprise: true },
      {
        name: "Role-based access control",
        community: true,
        cloud: true,
        enterprise: true,
      },
      { name: "API keys", community: true, cloud: true, enterprise: true },
      { name: "Admin panel", community: false, cloud: false, enterprise: true },
      {
        name: "Billing management",
        community: false,
        cloud: false,
        enterprise: true,
      },
    ],
  },
  {
    label: "Deployment",
    features: [
      {
        name: "Self-hosted (Docker)",
        community: true,
        cloud: false,
        enterprise: true,
      },
      { name: "Cloud hosted", community: false, cloud: true, enterprise: true },
      {
        name: "Managed infrastructure",
        community: false,
        cloud: true,
        enterprise: true,
      },
      { name: "On-premises", community: false, cloud: false, enterprise: true },
      {
        name: "Kubernetes / Helm",
        community: false,
        cloud: false,
        enterprise: true,
      },
    ],
  },
  {
    label: "Support",
    features: [
      {
        name: "Community (GitHub)",
        community: true,
        cloud: true,
        enterprise: true,
      },
      {
        name: "Email support",
        community: false,
        cloud: true,
        enterprise: true,
      },
      {
        name: "Priority support & SLA",
        community: false,
        cloud: false,
        enterprise: true,
      },
      {
        name: "Dedicated account manager",
        community: false,
        cloud: false,
        enterprise: true,
      },
      {
        name: "Onboarding assistance",
        community: false,
        cloud: false,
        enterprise: true,
      },
    ],
  },
];

function CellValue({ value }: { value: FeatureValue }) {
  if (typeof value === "string") {
    return <span className="text-xs text-text-secondary">{value}</span>;
  }
  if (value) {
    return (
      <CheckIcon
        className="w-4 h-4 text-accent-green mx-auto"
        strokeWidth={2}
        aria-hidden="true"
      />
    );
  }
  return (
    <XMarkIcon
      className="w-4 h-4 text-text-muted/40 mx-auto"
      strokeWidth={2}
      aria-hidden="true"
    />
  );
}

export function ComparisonTable() {
  return (
    <Section>
      <SectionHeader eyebrow="Compare Plans" title="Feature comparison" />

      <Reveal>
        <div className="overflow-x-auto -mx-8 px-8">
          <table className="w-full min-w-[640px]">
            <thead>
              <tr className="border-b border-border">
                <th
                  className="text-left py-4 pr-4 text-sm font-medium text-text-secondary w-[40%]"
                  scope="col"
                >
                  <span className="sr-only">Feature</span>
                </th>
                <th className="py-4 px-4 text-center text-sm font-semibold w-[20%]">
                  Community
                </th>
                <th className="py-4 px-4 text-center text-sm font-semibold w-[20%]">
                  <span className="text-primary">Cloud</span>
                </th>
                <th className="py-4 px-4 text-center text-sm font-semibold w-[20%]">
                  Enterprise
                </th>
              </tr>
            </thead>
            <tbody>
              {categories.map((cat) => (
                <Fragment key={cat.label}>
                  <tr>
                    <td colSpan={4} className="pt-8 pb-3">
                      <span className="text-2xs font-mono font-semibold uppercase tracking-label text-primary">
                        {cat.label}
                      </span>
                    </td>
                  </tr>
                  {cat.features.map((f) => (
                    <tr
                      key={f.name}
                      className="border-b border-border/50 hover:bg-white/[0.01] transition-colors"
                    >
                      <td className="py-3 pr-4 text-sm text-text-secondary">
                        {f.name}
                      </td>
                      <td className="py-3 px-4 text-center">
                        <CellValue value={f.community} />
                      </td>
                      <td className="py-3 px-4 text-center">
                        <CellValue value={f.cloud} />
                      </td>
                      <td className="py-3 px-4 text-center">
                        <CellValue value={f.enterprise} />
                      </td>
                    </tr>
                  ))}
                </Fragment>
              ))}
            </tbody>
          </table>
        </div>
      </Reveal>
    </Section>
  );
}
