import {
  ArrowRightIcon,
  CheckIcon,
  CommandLineIcon,
  DocumentTextIcon,
  FolderIcon,
  PencilIcon,
  ShieldCheckIcon,
  TagIcon,
} from "@heroicons/react/24/outline";
import { Badge, WindowChrome } from "@shellhub/design-system/primitives";
import {
  ConnectionGrid,
  GlowOrbs,
  Reveal,
  ShimmerCard,
} from "@shellhub/design-system/components";
import { SiteLayout } from "@/components/SiteLayout";
import {
  ActionButtonGroup,
  CTABanner,
  InfoCard,
  Section,
  SectionHeader,
  type CTAAction,
} from "@/components/marketing";
import { docsUrl } from "@/links";
import { C } from "@shellhub/design-system/constants";

const primaryAction: CTAAction = {
  label: "Get Started Free",
  to: "/getting-started",
};
const secondaryAction: CTAAction = {
  label: "Read the Docs",
  href: docsUrl,
  external: true,
};

const painPoints = [
  {
    title: "VPN tunnels in CI pipelines",
    desc: "Establishing VPN connections from ephemeral CI runners to target devices adds latency, complexity, and fragile points of failure to every deployment.",
    color: C.primary,
  },
  {
    title: "SSH key sprawl",
    desc: "Distributing, rotating, and revoking SSH keys across CI systems, bastion hosts, and hundreds of target devices is a security and operational burden.",
    color: C.yellow,
  },
  {
    title: "No stable device addressing",
    desc: "Devices behind NAT or CGNAT have no fixed addresses. Ansible inventories and Terraform providers simply cannot reach them.",
    color: C.red,
  },
  {
    title: "Limited deployment visibility",
    desc: "When automated pipelines SSH into devices there is often no centralized log of what ran where -- making debugging failed rollouts painful.",
    color: C.cyan,
  },
];

const features = [
  {
    icon: CommandLineIcon,
    color: C.primary,
    title: "Standard SSH Transport",
    desc: "Ansible, Terraform, and CI/CD tools connect through ShellHub using standard SSH -- no custom plugins or connectors required.",
  },
  {
    icon: DocumentTextIcon,
    color: C.cyan,
    title: "SCP/SFTP Transfers",
    desc: "Push build artifacts, configuration files, and firmware updates to remote devices as part of your deployment pipeline.",
  },
  {
    icon: PencilIcon,
    color: C.green,
    title: "Audit Logging",
    desc: "Every automated session is recorded -- see exactly which pipeline ran what command on which device and when.",
  },
  {
    icon: TagIcon,
    color: C.yellow,
    title: "Device Tags",
    desc: "Target deployments to specific device groups using tags -- deploy to 'staging' or 'production' fleets with a single inventory.",
  },
  {
    icon: ShieldCheckIcon,
    color: C.primary,
    title: "Firewall Rules",
    desc: "Restrict CI runner access to only the devices they need with IP-based and identity-based firewall policies.",
  },
  {
    icon: FolderIcon,
    color: C.green,
    title: "Namespaces",
    desc: "Isolate staging and production environments in separate namespaces with independent access policies and device groups.",
  },
];

/* ------------------------------------------------------------------ */
/*  Ansible terminal lines                                             */
/* ------------------------------------------------------------------ */

const ansibleLines: {
  text: string;
  color?: string;
  dim?: boolean;
  indent?: boolean;
}[] = [
  { text: "$ ansible-playbook -i inventory deploy.yml", color: C.green },
  { text: "" },
  {
    text: "PLAY [Deploy to edge devices] ***********************",
    color: C.cyan,
  },
  { text: "" },
  {
    text: "TASK [Gathering Facts] ******************************",
    color: C.cyan,
  },
  { text: "ok: [device-01.shellhub]", color: C.green, indent: true },
  { text: "ok: [device-02.shellhub]", color: C.green, indent: true },
  { text: "" },
  {
    text: "TASK [Update firmware] ******************************",
    color: C.cyan,
  },
  { text: "changed: [device-01.shellhub]", color: C.yellow, indent: true },
  { text: "changed: [device-02.shellhub]", color: C.yellow, indent: true },
  { text: "" },
  {
    text: "TASK [Restart service] ******************************",
    color: C.cyan,
  },
  { text: "ok: [device-01.shellhub]", color: C.green, indent: true },
  { text: "ok: [device-02.shellhub]", color: C.green, indent: true },
  { text: "" },
  {
    text: "PLAY RECAP ******************************************",
    color: C.cyan,
  },
  {
    text: "device-01.shellhub  : ok=3  changed=1  failed=0",
    color: C.green,
    indent: true,
  },
  {
    text: "device-02.shellhub  : ok=3  changed=1  failed=0",
    color: C.green,
    indent: true,
  },
];

/* ------------------------------------------------------------------ */
/*  Terraform HCL lines                                                */
/* ------------------------------------------------------------------ */

const terraformLines: { text: string; color?: string; dim?: boolean }[] = [
  { text: 'resource "shellhub_device" "edge_fleet" {', color: C.primary },
  { text: "  count      = var.fleet_size", dim: true },
  { text: '  namespace  = "production"', dim: true },
  { text: '  hostname   = "edge-${count.index + 1}"', dim: true },
  { text: "" },
  { text: "  tags = {", dim: true },
  { text: '    environment = "production"', color: C.green },
  { text: "    region      = var.region", color: C.green },
  { text: '    managed_by  = "terraform"', color: C.green },
  { text: "  }", dim: true },
  { text: "" },
  { text: "  connection {", dim: true },
  { text: '    type = "ssh"', color: C.cyan },
  { text: '    host = "${self.sshid}@shellhub.io"', color: C.cyan },
  { text: "  }", dim: true },
  { text: "}", color: C.primary },
];

/* ------------------------------------------------------------------ */
/*  CI / CD pipeline steps                                             */
/* ------------------------------------------------------------------ */

const pipelineSteps = [
  { label: "Build", status: "done" as const },
  { label: "Test", status: "done" as const },
  { label: "Deploy via ShellHub", status: "done" as const },
];

const pipelineDeployLines: { text: string; color?: string; dim?: boolean }[] = [
  { text: "deploy:", color: C.primary },
  { text: "  runs-on: ubuntu-latest", dim: true },
  { text: "  steps:", dim: true },
  { text: "    - name: Deploy to edge fleet", color: C.cyan },
  { text: "      run: |", dim: true },
  { text: "        ssh user@device-01.prod@shellhub.io \\", color: C.green },
  {
    text: '          "cd /opt/app && git pull && make restart"',
    color: C.green,
  },
  { text: "        ssh user@device-02.prod@shellhub.io \\", color: C.green },
  {
    text: '          "cd /opt/app && git pull && make restart"',
    color: C.green,
  },
];

/* ------------------------------------------------------------------ */
/*  Helpers                                                            */
/* ------------------------------------------------------------------ */

function CodeLine({
  text,
  color,
  dim,
  indent,
}: {
  text: string;
  color?: string;
  dim?: boolean;
  indent?: boolean;
}) {
  if (text === "") return <div className="h-3" />;
  return (
    <div
      className="font-mono text-2xs leading-relaxed whitespace-pre"
      style={{
        color: dim ? C.textMuted : color || C.text,
        paddingLeft: indent ? "1rem" : undefined,
      }}
    >
      {text}
    </div>
  );
}

/* ------------------------------------------------------------------ */
/*  Page                                                               */
/* ------------------------------------------------------------------ */

export default function DevopsCiCd() {
  return (
    <SiteLayout>
      {/* ── Hero ─────────────────────────────────────────────────── */}
      <section className="relative pt-32 pb-24 overflow-hidden">
        <ConnectionGrid />
        <GlowOrbs preset="section" tone="primary" />

        <div className="max-w-7xl mx-auto px-8 relative z-10 text-center">
          <Reveal>
            <Badge shape="pill" color="green" className="mb-6 tracking-label">
              Use Case
            </Badge>
          </Reveal>
          <Reveal>
            <h1 className="text-[clamp(2rem,5vw,3.5rem)] font-bold tracking-[-0.03em] leading-[1.1] mb-6 max-w-4xl mx-auto">
              DevOps & CI/CD{" "}
              <span className="bg-gradient-to-r from-primary via-accent-blue to-accent-cyan bg-clip-text text-transparent">
                Automation
              </span>
            </h1>
          </Reveal>
          <Reveal>
            <p className="text-sm text-text-secondary max-w-lg mx-auto leading-relaxed mb-10">
              Automate deployments to remote and edge devices with Ansible,
              Terraform, and CI/CD pipelines -- using standard SSH through
              ShellHub.
            </p>
          </Reveal>
          <Reveal>
            <ActionButtonGroup
              primaryAction={primaryAction}
              secondaryAction={secondaryAction}
            />
          </Reveal>
        </div>
      </section>

      {/* ── Ansible Integration ──────────────────────────────────── */}
      <Section>
        <div className="grid lg:grid-cols-2 gap-12 items-center">
          <div>
            <SectionHeader
              align="left"
              size="sub"
              className="mb-8"
              eyebrow="Ansible Integration"
              title="Use ShellHub as your SSH transport for Ansible"
              subtitle="Point your Ansible inventory at ShellHub device identifiers and run playbooks against devices behind NAT, CGNAT, or firewalls -- no VPN required. ShellHub acts as a transparent SSH gateway so existing playbooks work without modification."
            />

            <div className="space-y-3">
              {[
                {
                  label: "Standard SSH connection",
                  desc: "No custom connection plugins -- Ansible uses its built-in SSH transport",
                },
                {
                  label: "NAT traversal included",
                  desc: "Reach devices that have no public IP or inbound ports",
                },
                {
                  label: "Fleet targeting with tags",
                  desc: "Build dynamic inventories based on ShellHub device tags",
                },
                {
                  label: "Session-level audit trail",
                  desc: "Every Ansible task is logged with full session recording",
                },
              ].map((item, i) => (
                <Reveal key={i} delay={i * 0.04}>
                  <div className="flex items-start gap-3">
                    <CheckIcon
                      className="w-4 h-4 text-accent-green shrink-0 mt-0.5"
                      strokeWidth={2}
                    />
                    <div>
                      <p className="text-sm font-medium text-text-primary">
                        {item.label}
                      </p>
                      <p className="text-xs text-text-secondary leading-relaxed">
                        {item.desc}
                      </p>
                    </div>
                  </div>
                </Reveal>
              ))}
            </div>
          </div>

          <Reveal delay={0.1}>
            <ShimmerCard>
              <WindowChrome variant="terminal" title="ansible-playbook">
                <div className="space-y-0 overflow-x-auto">
                  {ansibleLines.map((line, i) => (
                    <CodeLine key={i} {...line} />
                  ))}
                </div>
              </WindowChrome>
            </ShimmerCard>
          </Reveal>
        </div>
      </Section>

      {/* ── Terraform Integration ────────────────────────────────── */}
      <Section>
        <div className="grid lg:grid-cols-2 gap-12 items-center">
          <Reveal delay={0.1} className="order-2 lg:order-1">
            <ShimmerCard>
              <WindowChrome variant="terminal" title="main.tf">
                <div className="space-y-0 overflow-x-auto">
                  {terraformLines.map((line, i) => (
                    <CodeLine key={i} {...line} />
                  ))}
                </div>
              </WindowChrome>
            </ShimmerCard>
          </Reveal>

          <div className="order-1 lg:order-2">
            <SectionHeader
              align="left"
              size="sub"
              className="mb-8"
              eyebrow="Terraform Integration"
              title="Infrastructure as code for device provisioning"
              subtitle="Define your device fleet declaratively with Terraform. Provision namespaces, assign tags, configure firewall rules, and manage access policies -- all versioned in Git and applied through your standard IaC workflow."
            />

            <div className="space-y-3">
              {[
                {
                  label: "Declarative device management",
                  desc: "Define fleet topology and access policies as HCL resources",
                },
                {
                  label: "GitOps-friendly",
                  desc: "Version your device configuration alongside application code",
                },
                {
                  label: "Drift detection",
                  desc: "Terraform plan shows configuration differences before applying",
                },
                {
                  label: "Reproducible environments",
                  desc: "Spin up identical staging and production fleets from the same config",
                },
              ].map((item, i) => (
                <Reveal key={i} delay={i * 0.04}>
                  <div className="flex items-start gap-3">
                    <CheckIcon
                      className="w-4 h-4 text-accent-green shrink-0 mt-0.5"
                      strokeWidth={2}
                    />
                    <div>
                      <p className="text-sm font-medium text-text-primary">
                        {item.label}
                      </p>
                      <p className="text-xs text-text-secondary leading-relaxed">
                        {item.desc}
                      </p>
                    </div>
                  </div>
                </Reveal>
              ))}
            </div>
          </div>
        </div>
      </Section>

      {/* ── CI/CD Pipeline ───────────────────────────────────────── */}
      <Section>
        <SectionHeader
          eyebrow="CI/CD Pipeline"
          title="Deploy to remote devices from any CI system"
          subtitle="GitHub Actions, GitLab CI, Jenkins -- any pipeline that can run SSH can deploy through ShellHub."
        />

        <Reveal delay={0.1}>
          <ShimmerCard className="max-w-3xl mx-auto">
            <WindowChrome
              variant="terminal"
              title=".github/workflows/deploy.yml"
              className="border-primary/30 shadow-[0_0_40px_rgba(102,122,204,0.1)]"
            >
              {/* Pipeline step indicators */}
              <div className="flex items-center gap-6 mb-6 pb-5 border-b border-border">
                {pipelineSteps.map((step, i) => (
                  <div key={i} className="flex items-center gap-2">
                    <div className="w-5 h-5 rounded-full bg-accent-green/15 border border-accent-green/30 flex items-center justify-center">
                      <CheckIcon
                        className="w-3 h-3 text-accent-green"
                        strokeWidth={2.5}
                      />
                    </div>
                    <span className="text-2xs font-mono text-text-secondary">
                      {step.label}
                    </span>
                    {i < pipelineSteps.length - 1 && (
                      <ArrowRightIcon
                        className="w-3 h-3 text-text-muted ml-2"
                        strokeWidth={2}
                      />
                    )}
                  </div>
                ))}
              </div>

              {/* Workflow YAML */}
              <div className="space-y-0 mb-6 overflow-x-auto">
                {pipelineDeployLines.map((line, i) => (
                  <CodeLine key={i} {...line} />
                ))}
              </div>

              {/* Status bar */}
              <div className="pt-4 border-t border-border flex items-center justify-between">
                <div className="flex items-center gap-2">
                  <div className="w-2 h-2 rounded-full bg-accent-green animate-pulse" />
                  <span className="text-2xs font-mono text-accent-green">
                    Pipeline succeeded
                  </span>
                </div>
                <span className="text-2xs text-text-muted font-mono">
                  2 devices updated in 34s
                </span>
              </div>
            </WindowChrome>
          </ShimmerCard>
        </Reveal>
      </Section>

      {/* ── Pain Points ──────────────────────────────────────────── */}
      <Section>
        <SectionHeader
          eyebrow="The Problem"
          title="Why automating remote deployments is hard"
          subtitle="Traditional approaches to device automation break down when devices are behind NAT, on unstable networks, or spread across locations."
        />

        <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
          {painPoints.map((p, i) => (
            <InfoCard
              key={i}
              color={p.color}
              title={p.title}
              description={p.desc}
              layout="dot"
              delay={i * 0.06}
            />
          ))}
        </div>
      </Section>

      {/* ── Key Features ─────────────────────────────────────────── */}
      <Section>
        <SectionHeader
          eyebrow="Capabilities"
          title="ShellHub features for DevOps teams"
          subtitle="Everything your automation toolchain needs to securely deploy to remote and edge devices at scale."
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

      <CTABanner
        eyebrow="Ready to automate?"
        title="Automate your device deployments today"
        subtitle="Connect ShellHub to your CI/CD pipeline and start deploying to remote devices in minutes -- no VPN, no static IPs, no hassle."
        primaryAction={primaryAction}
        secondaryAction={secondaryAction}
      />
    </SiteLayout>
  );
}
