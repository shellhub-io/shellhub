import { cn } from "@shellhub/design-system/cn";
import {
  CheckIcon,
  ComputerDesktopIcon,
  LockClosedIcon,
  PencilIcon,
  PlayCircleIcon,
  ShieldCheckIcon,
  UsersIcon,
} from "@heroicons/react/24/outline";
import { PlayIcon } from "@heroicons/react/24/solid";
import { Badge, Card, WindowChrome } from "@shellhub/design-system/primitives";
import {
  ConnectionGrid,
  GlowOrbs,
  Reveal,
  ShimmerCard,
} from "@shellhub/design-system/components";
import {
  ActionButtonGroup,
  CTABanner,
  InfoCard,
  Section,
  SectionHeader,
  SiteLayout,
  type CTAAction,
} from "@/components";
import { C } from "@shellhub/design-system/constants";

const primaryAction: CTAAction = {
  label: "Get Started Free",
  to: "/getting-started",
};
const secondaryAction: CTAAction = { label: "View Pricing", to: "/pricing" };

/* ─── Pain-point cards ─────────────────────────────────────────────── */

const painPoints = [
  {
    title: "No audit trail",
    desc: "When support engineers SSH into customer devices, there is no record of what was done — making incident review impossible.",
    color: C.yellow,
  },
  {
    title: "Overprivileged access",
    desc: "Support teams often share root credentials or use a single SSH key for all devices, creating security blind spots.",
    color: C.red,
  },
  {
    title: "VPN complexity",
    desc: "Setting up VPN tunnels for each support session is slow, error-prone, and does not scale across customers or regions.",
    color: C.primary,
  },
  {
    title: "No session visibility",
    desc: "Managers cannot see active sessions or verify that support engineers are following proper procedures in real time.",
    color: C.cyan,
  },
];

/* ─── Feature cards ────────────────────────────────────────────────── */

const features = [
  {
    icon: PlayCircleIcon,
    color: C.cyan,
    title: "Session Recording",
    desc: "Every support session is recorded and can be replayed for quality assurance, training, and compliance audits.",
  },
  {
    icon: PencilIcon,
    color: C.green,
    title: "Audit Logging",
    desc: "Complete audit trail with timestamps, user identity, and session details for every connection made to any device.",
  },
  {
    icon: UsersIcon,
    color: C.primary,
    title: "Role-Based Access",
    desc: "Assign support engineers access to specific devices or groups — revoke access instantly when the ticket is closed.",
  },
  {
    icon: ComputerDesktopIcon,
    color: C.green,
    title: "Web Terminal",
    desc: "Support engineers access devices straight from the browser — no SSH client setup required on their workstations.",
  },
  {
    icon: LockClosedIcon,
    color: C.yellow,
    title: "MFA for SSH",
    desc: "Require multi-factor authentication for support sessions — adding a security layer beyond SSH keys alone.",
  },
  {
    icon: ShieldCheckIcon,
    color: C.primary,
    title: "Firewall Rules",
    desc: "Restrict support access by IP, time window, or device group with granular, time-based firewall policies.",
  },
];

/* ─── Audit log mock data ──────────────────────────────────────────── */

const auditRows = [
  {
    time: "14:32:08",
    initials: "JD",
    name: "john@company.com",
    action: "SSH session",
    device: "prod-server-03",
    duration: "12m 15s",
    color: C.primary,
  },
  {
    time: "14:18:41",
    initials: "AL",
    name: "ana@company.com",
    action: "File transfer",
    device: "edge-gw-eu-01",
    duration: "3m 42s",
    color: C.cyan,
  },
  {
    time: "13:55:19",
    initials: "MS",
    name: "mike@company.com",
    action: "SSH session",
    device: "staging-db-02",
    duration: "8m 04s",
    color: C.green,
  },
  {
    time: "13:40:03",
    initials: "JD",
    name: "john@company.com",
    action: "SSH session",
    device: "iot-gateway-07",
    duration: "5m 51s",
    color: C.yellow,
  },
];

/* ─── Workflow steps ───────────────────────────────────────────────── */

const workflowSteps = [
  {
    num: "01",
    title: "Ticket Created",
    desc: "A customer reports an issue. The support ticket is created and assigned to an engineer with the right permissions.",
    color: C.yellow,
  },
  {
    num: "02",
    title: "Engineer Connects via ShellHub",
    desc: "The engineer opens a browser-based terminal and connects to the device instantly — no VPN, no key sharing.",
    color: C.primary,
  },
  {
    num: "03",
    title: "Session Recorded & Logged",
    desc: "Every keystroke is recorded. The audit log captures who connected, when, what was done, and for how long.",
    color: C.green,
  },
];

/* ─── Page ─────────────────────────────────────────────────────────── */

export default function RemoteSupport() {
  return (
    <SiteLayout>
      {/* ── Hero ──────────────────────────────────────────────────── */}
      <section className="relative pt-32 pb-24 overflow-hidden">
        <ConnectionGrid />
        <GlowOrbs preset="section" tone="yellow" />

        <div className="max-w-7xl mx-auto px-8 relative z-raised text-center">
          <Reveal>
            <Badge shape="pill" color="yellow" className="mb-6">
              Use Case
            </Badge>
          </Reveal>
          <Reveal>
            <h1 className="text-[clamp(2rem,5vw,3.5rem)] font-bold tracking-[-0.03em] leading-[1.1] mb-6 max-w-3xl mx-auto">
              Remote Support &{" "}
              <span className="bg-gradient-to-r from-accent-yellow via-primary to-accent-cyan bg-clip-text text-transparent">
                Troubleshooting
              </span>
            </h1>
          </Reveal>
          <Reveal>
            <p className="text-sm text-text-secondary max-w-lg mx-auto leading-relaxed mb-10">
              Give your support team secure, audited access to customer devices
              — with session recording, full audit trails, and complete
              accountability for every connection.
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

      {/* ── Session Recording Mockup ─────────────────────────────── */}
      <Section>
        <div className="grid lg:grid-cols-2 gap-12 items-center">
          <div>
            <SectionHeader
              align="left"
              size="sub"
              className="mb-8"
              eyebrowColor="cyan"
              eyebrow="Session Recording"
              title="Record every session. Replay any time."
              subtitle="Every SSH session is automatically captured and stored. Replay sessions to review what happened during a support interaction, verify that procedures were followed, or train new engineers on troubleshooting workflows."
            />

            <div className="space-y-3">
              {[
                {
                  label: "Full terminal replay",
                  desc: "Watch sessions frame by frame, exactly as they happened",
                },
                {
                  label: "Searchable archive",
                  desc: "Find sessions by user, device, date range, or keyword",
                },
                {
                  label: "Compliance-ready exports",
                  desc: "Export session recordings for external audits and compliance reports",
                },
                {
                  label: "Incident investigation",
                  desc: "Pinpoint exactly what command caused an issue, down to the second",
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

          {/* Session player mockup */}
          <Reveal delay={0.1}>
            <ShimmerCard>
              <WindowChrome variant="terminal" title="Session Replay">
                {/* Terminal area */}
                <div className="bg-[#111214] rounded-lg border border-border p-4 font-mono text-2xs leading-relaxed mb-4">
                  <p className="text-accent-green">
                    john@prod-server-03:~$
                    <span className="text-text-primary ml-2">
                      ls -la /var/log/nginx/
                    </span>
                  </p>
                  <p className="text-text-muted mt-1">total 2184</p>
                  <p className="text-text-muted">
                    drwxr-xr-x 2 root root 4096 Feb 14 09:00 .
                  </p>
                  <p className="text-text-muted">
                    -rw-r----- 1 root adm 842560 Feb 14 14:31 access.log
                  </p>
                  <p className="text-text-muted">
                    -rw-r----- 1 root adm 194720 Feb 14 14:28 error.log
                  </p>
                  <p className="text-accent-green mt-2">
                    john@prod-server-03:~$
                    <span className="text-text-primary ml-2">
                      systemctl status nginx
                    </span>
                  </p>
                  <p className="text-text-muted mt-1">
                    <span className="text-accent-green">●</span> nginx.service -
                    A high performance web server
                  </p>
                  <p className="text-text-muted">
                    {"   "}Loaded: loaded (/lib/systemd/system/nginx.service;
                    enabled)
                  </p>
                  <p className="text-text-muted">
                    {"   "}Active:{" "}
                    <span className="text-accent-green">active (running)</span>{" "}
                    since Fri 2026-02-14 09:00:12 UTC
                  </p>
                  <p className="text-accent-green mt-2">
                    john@prod-server-03:~$
                    <span className="text-text-primary ml-2 animate-pulse">
                      _
                    </span>
                  </p>
                </div>

                {/* Playback controls */}
                <div className="flex items-center gap-3 bg-surface rounded-lg border border-border p-3">
                  <button
                    type="button"
                    aria-label="Play"
                    className="w-8 h-8 rounded-lg bg-accent-cyan/10 border border-accent-cyan/20 flex items-center justify-center shrink-0"
                  >
                    <PlayIcon
                      className="w-3.5 h-3.5"
                      style={{ color: C.cyan }}
                      aria-hidden="true"
                    />
                  </button>

                  {/* Timeline */}
                  <div className="flex-1 relative">
                    <div className="h-1.5 bg-border rounded-full overflow-hidden">
                      <div
                        className="h-full rounded-full bg-gradient-to-r from-accent-cyan to-primary"
                        style={{ width: "30%" }}
                      />
                    </div>
                  </div>

                  <span className="text-2xs text-text-muted font-mono shrink-0">
                    00:03:42 / 00:12:15
                  </span>
                </div>

                {/* Metadata */}
                <div className="mt-4 pt-4 border-t border-border">
                  <p className="text-2xs text-text-muted font-mono">
                    User:{" "}
                    <span className="text-text-secondary">
                      john@company.com
                    </span>
                    <span className="mx-2 text-border">|</span>
                    Device:{" "}
                    <span className="text-text-secondary">prod-server-03</span>
                    <span className="mx-2 text-border">|</span>
                    Duration:{" "}
                    <span className="text-text-secondary">12m 15s</span>
                  </p>
                </div>
              </WindowChrome>
            </ShimmerCard>
          </Reveal>
        </div>
      </Section>

      {/* ── Audit Trail Mockup ───────────────────────────────────── */}
      <Section>
        <div className="grid lg:grid-cols-2 gap-12 items-center">
          {/* Audit log table mockup */}
          <Reveal delay={0.1} className="order-2 lg:order-1">
            <ShimmerCard>
              <WindowChrome variant="terminal" title="Audit Log">
                {/* Table header */}
                <div className="grid grid-cols-[60px_1fr_1fr_1fr_70px] gap-2 px-3 py-2 mb-1">
                  <span className="text-2xs font-mono font-semibold text-text-muted uppercase tracking-wider">
                    Time
                  </span>
                  <span className="text-2xs font-mono font-semibold text-text-muted uppercase tracking-wider">
                    User
                  </span>
                  <span className="text-2xs font-mono font-semibold text-text-muted uppercase tracking-wider">
                    Action
                  </span>
                  <span className="text-2xs font-mono font-semibold text-text-muted uppercase tracking-wider">
                    Device
                  </span>
                  <span className="text-2xs font-mono font-semibold text-text-muted uppercase tracking-wider">
                    Duration
                  </span>
                </div>

                {/* Rows */}
                <div className="space-y-1">
                  {auditRows.map((row, i) => (
                    <div
                      key={i}
                      className={cn(
                        "grid grid-cols-[60px_1fr_1fr_1fr_70px] gap-2 px-3 py-2.5 rounded-lg items-center",
                        i % 2 === 0 ? "bg-surface" : "bg-transparent",
                      )}
                    >
                      <span className="text-2xs font-mono text-text-muted">
                        {row.time}
                      </span>
                      <div className="flex items-center gap-2 min-w-0">
                        <div
                          className="w-6 h-6 rounded-full flex items-center justify-center text-[9px] font-semibold shrink-0"
                          style={{
                            background: `${row.color}20`,
                            color: row.color,
                          }}
                        >
                          {row.initials}
                        </div>
                        <span className="text-2xs text-text-secondary truncate">
                          {row.name}
                        </span>
                      </div>
                      <span className="text-2xs text-text-secondary font-mono">
                        {row.action}
                      </span>
                      <span className="text-2xs text-text-secondary font-mono truncate">
                        {row.device}
                      </span>
                      <span className="text-2xs text-text-muted font-mono">
                        {row.duration}
                      </span>
                    </div>
                  ))}
                </div>

                {/* Footer */}
                <div className="mt-4 pt-4 border-t border-border flex items-center justify-between">
                  <span className="text-2xs text-text-muted">
                    Showing 4 of 1,247 entries
                  </span>
                  <span className="text-2xs text-primary font-medium">
                    View all &rarr;
                  </span>
                </div>
              </WindowChrome>
            </ShimmerCard>
          </Reveal>

          {/* Text */}
          <div className="order-1 lg:order-2">
            <SectionHeader
              align="left"
              size="sub"
              className="mb-8"
              eyebrowColor="green"
              eyebrow="Audit Trail"
              title="Full accountability for every connection"
              subtitle="Every session is logged with the user identity, target device, timestamps, and duration. Build compliance reports, investigate incidents, and prove to auditors exactly who did what, when, and for how long."
            />

            <div className="space-y-3">
              {[
                {
                  label: "Immutable log entries",
                  desc: "Audit records cannot be altered or deleted by users",
                },
                {
                  label: "Filter and search",
                  desc: "Find events by user, device, date, or action type instantly",
                },
                {
                  label: "Compliance reporting",
                  desc: "Generate reports for SOC 2, ISO 27001, and internal audits",
                },
                {
                  label: "Real-time monitoring",
                  desc: "View active sessions and who is currently connected to which device",
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

      <Section>
        <SectionHeader
          eyebrow="The Problem"
          title="Why traditional remote support falls short"
          subtitle="Legacy workflows leave gaps in security, visibility, and compliance that put your organization at risk."
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

      <Section>
        <SectionHeader
          eyebrow="Features"
          title="Built for audited, accountable support"
          subtitle="Everything your support team needs to connect securely, troubleshoot efficiently, and maintain a complete compliance record."
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

      {/* ── Support Workflow ──────────────────────────────────────── */}
      <Section>
        <SectionHeader
          eyebrow="Workflow"
          title="From ticket to resolution — fully audited"
          subtitle="A streamlined support workflow that captures everything for compliance without slowing your team down."
        />

        <div className="grid grid-cols-1 md:grid-cols-3 gap-6 relative">
          {/* Connecting arrows (desktop only) */}
          <div className="hidden md:block absolute top-1/2 left-[calc(33.33%-12px)] w-[calc(33.33%+24px)] -translate-y-1/2 pointer-events-none z-base">
            <svg
              className="w-full h-8"
              viewBox="0 0 400 32"
              fill="none"
              preserveAspectRatio="none"
            >
              <line
                x1="0"
                y1="16"
                x2="180"
                y2="16"
                stroke={C.border}
                strokeWidth="1"
                strokeDasharray="6 4"
              />
              <polygon points="180,12 188,16 180,20" fill={C.textMuted} />
              <line
                x1="210"
                y1="16"
                x2="390"
                y2="16"
                stroke={C.border}
                strokeWidth="1"
                strokeDasharray="6 4"
              />
              <polygon points="390,12 398,16 390,20" fill={C.textMuted} />
            </svg>
          </div>

          {workflowSteps.map((step, i) => (
            <Reveal key={i} delay={i * 0.08}>
              <ShimmerCard className="h-full">
                <Card hover className="relative p-6 h-full overflow-hidden">
                  <div
                    className="absolute top-0 left-0 w-full h-0.5"
                    style={{
                      background: `linear-gradient(90deg, ${step.color}00, ${step.color}, ${step.color}00)`,
                    }}
                  />
                  <div className="flex items-center gap-3 mb-4">
                    <span
                      className="w-10 h-10 rounded-lg flex items-center justify-center text-xs font-bold font-mono border"
                      style={{
                        background: `${step.color}15`,
                        borderColor: `${step.color}25`,
                        color: step.color,
                      }}
                    >
                      {step.num}
                    </span>
                    <h4 className="text-sm font-semibold">{step.title}</h4>
                  </div>
                  <p className="text-xs text-text-secondary leading-relaxed">
                    {step.desc}
                  </p>
                </Card>
              </ShimmerCard>
            </Reveal>
          ))}
        </div>
      </Section>

      <CTABanner
        eyebrow="Ready to get started?"
        title="Secure support starts here"
        subtitle="Enable audited, accountable remote support for your team in minutes. No VPN required, no SSH keys to manage, no audit gaps."
        primaryAction={primaryAction}
        secondaryAction={secondaryAction}
      />
    </SiteLayout>
  );
}
