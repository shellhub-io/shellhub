import { Link } from "react-router-dom";
import {
  CheckIcon,
  ComputerDesktopIcon,
  DocumentTextIcon,
  ExclamationTriangleIcon,
  PencilIcon,
  PlayCircleIcon,
  UsersIcon,
  XMarkIcon,
} from "@heroicons/react/24/outline";
import {
  Badge,
  Button,
  Card,
  IconBadge,
  WindowChrome,
} from "@shellhub/design-system/primitives";
import { GlowOrbs } from "@shellhub/design-system/components";
import { ArrowRight } from "@/components/ArrowRight";
import { SiteLayout } from "@/components/SiteLayout";
import { CTABanner, InfoCard, Section, SectionHeader } from "@/components/marketing";
import { FeatureListItem } from "@/components/marketing/FeatureListItem";
import { Reveal, ShimmerCard, ConnectionGrid } from "../landing/components";
import { C, FONT_SANS, FONT_MONO } from "../landing/constants";

/* ═══════ Pain Points ═══════ */
const painPoints = [
  {
    title: "docker exec isn't remote",
    desc: "Docker exec requires local access to the host. For containers on remote servers you chain SSH + docker exec, adding complexity and fragility.",
    color: C.primary,
  },
  {
    title: "No access control",
    desc: "Anyone with SSH access to the Docker host can exec into any container. There's no per-container permission model.",
    color: C.red,
  },
  {
    title: "No audit trail",
    desc: "Docker exec sessions aren't logged or recorded. When something breaks inside a container, there's no record of who did what.",
    color: C.yellow,
  },
  {
    title: "Port exposure risks",
    desc: "Running sshd inside containers or exposing the Docker API creates attack surface and violates container best practices.",
    color: C.cyan,
  },
];

/* ═══════ Key Features (2x2) ═══════ */
const smallFeatures = [
  {
    icon: PlayCircleIcon,
    color: C.red,
    title: "Session Recording",
    desc: "Every container session is captured and replayable for compliance audits, post-incident review, and team training.",
  },
  {
    icon: ComputerDesktopIcon,
    color: C.green,
    title: "Web Terminal",
    desc: "Access containers directly from the browser. No SSH client, Docker CLI, or VPN required on your workstation.",
  },
  {
    icon: DocumentTextIcon,
    color: C.yellow,
    title: "SCP / SFTP into Containers",
    desc: "Transfer files in and out of containers with standard SCP and SFTP. No volume mounts or docker cp gymnastics.",
  },
  {
    icon: PencilIcon,
    color: C.blue,
    title: "Audit Logging",
    desc: "Full audit trail of every container session. Who connected, when, from where, and what they executed.",
  },
];

/* ═══════ Permissions mockup rows ═══════ */
const permRows = [
  {
    container: "api-server",
    user: "jane@co.com",
    role: "Admin",
    level: "Full",
    accent: C.green,
  },
  {
    container: "worker-01",
    user: "mike@co.com",
    role: "Operator",
    level: "Shell only",
    accent: C.primary,
  },
  {
    container: "redis-cache",
    user: "ana@co.com",
    role: "Observer",
    level: "Read only",
    accent: C.cyan,
  },
  {
    container: "postgres-db",
    user: "dev-team",
    role: "Operator",
    level: "Shell only",
    accent: C.primary,
  },
];

/* ═══════ How-it-works steps ═══════ */
const steps = [
  {
    num: "01",
    title: "Install agent on Docker host",
    desc: "Deploy the lightweight ShellHub agent alongside your containers. One binary, one line of config.",
    color: C.primary,
  },
  {
    num: "02",
    title: "Agent discovers containers",
    desc: "The agent automatically detects running containers and registers each one as an addressable SSH target.",
    color: C.cyan,
  },
  {
    num: "03",
    title: "SSH into any container by name",
    desc: "Connect to any container with a single ssh command using its human-readable name. No IPs, no ports.",
    color: C.green,
  },
];

/* ═══════ Architecture SVG ═══════ */
function ArchitectureDiagram() {
  return (
    <svg
      viewBox="0 0 900 320"
      fill="none"
      className="w-full"
      xmlns="http://www.w3.org/2000/svg"
    >
      {/* Background grid */}
      <defs>
        <pattern
          id="cm-grid"
          width="40"
          height="40"
          patternUnits="userSpaceOnUse"
        >
          <path
            d="M 40 0 L 0 0 0 40"
            fill="none"
            stroke={C.border}
            strokeWidth="0.5"
            opacity="0.4"
          />
        </pattern>
        <linearGradient id="cm-line-grad" x1="0" y1="0" x2="1" y2="0">
          <stop offset="0%" stopColor={C.primary} stopOpacity="0.6" />
          <stop offset="100%" stopColor={C.cyan} stopOpacity="0.6" />
        </linearGradient>
        <filter id="cm-glow">
          <feGaussianBlur stdDeviation="4" result="blur" />
          <feMerge>
            <feMergeNode in="blur" />
            <feMergeNode in="SourceGraphic" />
          </feMerge>
        </filter>
      </defs>
      <rect width="900" height="320" fill={C.card} rx="12" />
      <rect width="900" height="320" fill="url(#cm-grid)" rx="12" />

      {/* ── User ── */}
      <rect
        x="30"
        y="115"
        width="130"
        height="90"
        rx="10"
        fill={C.surface}
        stroke={C.border}
        strokeWidth="1"
      />
      <svg
        x="72"
        y="128"
        width="28"
        height="28"
        viewBox="0 0 24 24"
        fill="none"
        stroke={C.textSec}
        strokeWidth="1.5"
        strokeLinecap="round"
      >
        <path d="M17 21v-2a4 4 0 00-4-4H5a4 4 0 00-4 4v2" />
        <circle cx="9" cy="7" r="4" />
      </svg>
      <text
        x="95"
        y="175"
        textAnchor="middle"
        fill={C.text}
        fontSize="12"
        fontWeight="600"
        fontFamily={FONT_SANS}
      >
        User
      </text>
      <text
        x="95"
        y="192"
        textAnchor="middle"
        fill={C.textMuted}
        fontSize="9"
        fontFamily={FONT_MONO}
      >
        ssh user@ctr
      </text>

      {/* Arrow 1 */}
      <line
        x1="160"
        y1="160"
        x2="250"
        y2="160"
        stroke="url(#cm-line-grad)"
        strokeWidth="1.5"
        strokeDasharray="6 4"
      />
      <polygon
        points="250,155 260,160 250,165"
        fill={C.primary}
        opacity="0.7"
      />

      {/* ── ShellHub Gateway ── */}
      <rect
        x="260"
        y="105"
        width="160"
        height="110"
        rx="10"
        fill={C.surface}
        stroke={C.primary}
        strokeWidth="1"
        opacity="0.9"
      />
      <rect
        x="260"
        y="105"
        width="160"
        height="110"
        rx="10"
        fill={`${C.primary}08`}
      />
      <svg
        x="316"
        y="118"
        width="28"
        height="28"
        viewBox="0 0 24 24"
        fill="none"
        stroke={C.primary}
        strokeWidth="1.5"
        strokeLinecap="round"
      >
        <path d="M12 22s8-4 8-10V5l-8-3-8 3v7c0 6 8 10 8 10z" />
      </svg>
      <text
        x="340"
        y="165"
        textAnchor="middle"
        fill={C.text}
        fontSize="12"
        fontWeight="600"
        fontFamily={FONT_SANS}
      >
        ShellHub Gateway
      </text>
      <text
        x="340"
        y="180"
        textAnchor="middle"
        fill={C.textMuted}
        fontSize="9"
        fontFamily={FONT_MONO}
      >
        auth + routing
      </text>
      <text
        x="340"
        y="200"
        textAnchor="middle"
        fill={C.textMuted}
        fontSize="9"
        fontFamily={FONT_MONO}
      >
        session recording
      </text>

      {/* Arrow 2 */}
      <line
        x1="420"
        y1="160"
        x2="510"
        y2="160"
        stroke="url(#cm-line-grad)"
        strokeWidth="1.5"
        strokeDasharray="6 4"
      />
      <polygon points="510,155 520,160 510,165" fill={C.cyan} opacity="0.7" />

      {/* ── Docker Host ── */}
      <rect
        x="520"
        y="40"
        width="360"
        height="260"
        rx="12"
        fill={C.surface}
        stroke={C.border}
        strokeWidth="1"
      />
      <text
        x="700"
        y="68"
        textAnchor="middle"
        fill={C.text}
        fontSize="12"
        fontWeight="600"
        fontFamily={FONT_SANS}
      >
        Docker Host
      </text>

      {/* Agent badge inside Docker Host */}
      <rect
        x="610"
        y="80"
        width="180"
        height="28"
        rx="6"
        fill={`${C.primary}12`}
        stroke={`${C.primary}30`}
        strokeWidth="1"
      />
      <svg
        x="618"
        y="85"
        width="16"
        height="16"
        viewBox="0 0 24 24"
        fill="none"
        stroke={C.primary}
        strokeWidth="1.5"
        strokeLinecap="round"
      >
        <rect x="2" y="2" width="20" height="20" rx="5" />
        <path d="M12 8v8M8 12h8" />
      </svg>
      <text
        x="700"
        y="99"
        textAnchor="middle"
        fill={C.primary}
        fontSize="10"
        fontWeight="600"
        fontFamily={FONT_MONO}
      >
        ShellHub Agent
      </text>

      {/* ── Containers ── */}
      {[
        { y: 125, name: "api-server", color: C.green },
        { y: 185, name: "worker-01", color: C.cyan },
        { y: 245, name: "redis-cache", color: C.yellow },
      ].map((ctr) => (
        <g key={ctr.name}>
          <rect
            x="570"
            y={ctr.y}
            width="260"
            height="48"
            rx="8"
            fill={C.card}
            stroke={C.border}
            strokeWidth="1"
          />
          <rect
            x="580"
            y={ctr.y + 10}
            width="28"
            height="28"
            rx="6"
            fill={`${ctr.color}15`}
            stroke={`${ctr.color}30`}
            strokeWidth="1"
          />
          <svg
            x="585"
            y={ctr.y + 15}
            width="18"
            height="18"
            viewBox="0 0 24 24"
            fill="none"
            stroke={ctr.color}
            strokeWidth="1.5"
            strokeLinecap="round"
          >
            <path d="M21 16V8a2 2 0 00-1-1.73l-7-4a2 2 0 00-2 0l-7 4A2 2 0 002 8v8a2 2 0 001 1.73l7 4a2 2 0 002 0l7-4A2 2 0 0021 16z" />
          </svg>
          <text
            x="620"
            y={ctr.y + 29}
            fill={C.text}
            fontSize="11"
            fontWeight="500"
            fontFamily={FONT_SANS}
          >
            {ctr.name}
          </text>
          <rect
            x="750"
            y={ctr.y + 14}
            width="60"
            height="20"
            rx="4"
            fill={`${ctr.color}12`}
            stroke={`${ctr.color}25`}
            strokeWidth="0.5"
          />
          <text
            x="780"
            y={ctr.y + 28}
            textAnchor="middle"
            fill={ctr.color}
            fontSize="8"
            fontWeight="600"
            fontFamily={FONT_MONO}
          >
            running
          </text>
          {/* Arrow from agent to container */}
          <line
            x1="540"
            y1={ctr.y + 24}
            x2="568"
            y2={ctr.y + 24}
            stroke={C.border}
            strokeWidth="1"
            strokeDasharray="3 3"
          />
        </g>
      ))}
    </svg>
  );
}

/* ═══════ Main Component ═══════ */
export default function ContainerManagement() {
  return (
    <SiteLayout>
      {/* ═══════ Hero ═══════ */}
      <section className="relative pt-32 pb-24 overflow-hidden">
        <ConnectionGrid />
        <GlowOrbs preset="section" tone="cyan" />

        <div className="max-w-7xl mx-auto px-8 relative z-10 text-center">
          <Reveal>
            <Badge shape="pill" color="cyan" className="mb-6 tracking-label">
              Use Case
            </Badge>
          </Reveal>
          <Reveal>
            <h1 className="text-[clamp(2rem,5vw,3.5rem)] font-bold tracking-[-0.03em] leading-[1.1] mb-6 max-w-3xl mx-auto">
              SSH into Docker containers{" "}
              <span className="bg-gradient-to-r from-accent-cyan via-primary to-accent-blue bg-clip-text text-transparent">
                without docker exec
              </span>
            </h1>
          </Reveal>
          <Reveal>
            <p className="text-sm text-text-secondary max-w-lg mx-auto leading-relaxed mb-10">
              Connect to containers on remote hosts with a single SSH command.
              No exposed ports, no sshd inside containers, no multi-step
              workflows. Just{" "}
              <span className="font-mono text-accent-cyan">
                ssh user@container.namespace
              </span>
              .
            </p>
          </Reveal>
          <Reveal>
            <div className="flex flex-col sm:flex-row items-center justify-center gap-3">
              <Button
                as={Link}
                to="/getting-started"
                variant="primary"
                size="xl"
                glow
                iconRight={<ArrowRight />}
              >
                Get Started Free
              </Button>
              <Button as={Link} to="/pricing" variant="outline" size="xl">
                Compare Plans
              </Button>
            </div>
          </Reveal>
        </div>
      </section>

      {/* ═══════ Before / After ═══════ */}
      <Section>
        <SectionHeader
          eyebrow="Before & After"
          title="From three steps to one"
          subtitle="See how ShellHub eliminates the multi-step dance of connecting to containers on remote hosts."
        />

        <div className="grid md:grid-cols-2 gap-6">
          {/* Without ShellHub */}
          <Reveal delay={0}>
            <ShimmerCard className="h-full">
              <Card hover className="p-8 flex flex-col h-full">
                <div className="flex items-center gap-3 mb-6">
                  <div className="w-10 h-10 rounded-lg bg-white/[0.04] border border-border flex items-center justify-center">
                    <XMarkIcon className="w-5 h-5 text-text-secondary" />
                  </div>
                  <div>
                    <h3 className="text-sm font-bold">Without ShellHub</h3>
                    <p className="text-2xs text-text-muted">
                      Manual, multi-step
                    </p>
                  </div>
                </div>

                {/* Fake terminal */}
                <WindowChrome variant="terminal" size="sm" className="flex-1">
                  <div className="space-y-2">
                    <p>
                      <span className="text-text-muted">
                        # Step 1: SSH into the Docker host
                      </span>
                    </p>
                    <p>
                      <span className="text-accent-green">$</span>{" "}
                      <span className="text-text-secondary">
                        ssh root@192.168.1.42
                      </span>
                    </p>
                    <p className="text-text-muted text-2xs">
                      Enter password...
                    </p>
                    <p>
                      <span className="text-text-muted">
                        # Step 2: Find the container
                      </span>
                    </p>
                    <p>
                      <span className="text-accent-green">$</span>{" "}
                      <span className="text-text-secondary">
                        docker ps | grep api
                      </span>
                    </p>
                    <p className="text-text-muted text-2xs">
                      abc123f8... api-server
                    </p>
                    <p>
                      <span className="text-text-muted">
                        # Step 3: Exec into it
                      </span>
                    </p>
                    <p>
                      <span className="text-accent-green">$</span>{" "}
                      <span className="text-text-secondary">
                        docker exec -it abc123 /bin/bash
                      </span>
                    </p>
                    <p className="text-text-muted text-2xs mt-2">
                      No audit. No access control. No recording.
                    </p>
                  </div>
                </WindowChrome>
              </Card>
            </ShimmerCard>
          </Reveal>

          {/* With ShellHub */}
          <Reveal delay={0.1}>
            <ShimmerCard className="h-full">
              <div className="relative bg-card border border-primary/30 rounded-xl p-8 flex flex-col h-full hover:border-primary/50 transition-all duration-300 shadow-[0_0_40px_rgba(102,122,204,0.1)] overflow-hidden">
                <div className="absolute inset-0 bg-gradient-to-br from-primary/[0.06] via-transparent to-transparent pointer-events-none" />
                <div className="relative flex-1 flex flex-col">
                  <div className="flex items-center gap-3 mb-6">
                    <IconBadge color="primary">
                      <CheckIcon className="w-5 h-5 text-primary" />
                    </IconBadge>
                    <Badge shape="pill" color="green">
                      Recommended
                    </Badge>
                  </div>

                  <h3 className="text-sm font-bold mb-1">With ShellHub</h3>
                  <p className="text-2xs text-primary mb-6">
                    One command, fully recorded
                  </p>

                  {/* Fake terminal */}
                  <WindowChrome
                    variant="terminal"
                    size="sm"
                    className="border-primary/20 flex-1"
                  >
                    <div className="space-y-2">
                      <p>
                        <span className="text-text-muted">
                          # One step. That's it.
                        </span>
                      </p>
                      <p>
                        <span className="text-accent-green">$</span>{" "}
                        <span className="text-text-primary">
                          ssh user@api-server.production
                        </span>
                      </p>
                      <p className="text-accent-green text-2xs mt-1">
                        Connected to api-server
                      </p>
                      <p className="text-text-muted text-2xs">
                        Session recorded: #ses-4a82f1
                      </p>
                      <p className="text-text-muted text-2xs">
                        Access policy: production-operators
                      </p>
                      <p className="text-text-muted text-2xs">MFA verified</p>
                      <p className="mt-3">
                        <span className="text-accent-cyan">
                          root@api-server:~#
                        </span>{" "}
                        <span className="inline-block w-2 h-3.5 bg-text-primary animate-pulse" />
                      </p>
                    </div>
                  </WindowChrome>

                  {/* Status badges */}
                  <div className="flex flex-wrap gap-2 mt-4">
                    {[
                      { label: "Recorded", color: C.green },
                      { label: "Access-controlled", color: C.primary },
                      { label: "MFA", color: C.yellow },
                      { label: "Audited", color: C.cyan },
                    ].map((b) => (
                      <span
                        key={b.label}
                        className="px-2 py-0.5 text-2xs font-mono border rounded-full"
                        style={{
                          color: b.color,
                          borderColor: `${b.color}30`,
                          background: `${b.color}10`,
                        }}
                      >
                        {b.label}
                      </span>
                    ))}
                  </div>
                </div>
              </div>
            </ShimmerCard>
          </Reveal>
        </div>
      </Section>

      {/* ═══════ Architecture Diagram ═══════ */}
      <Section>
        <SectionHeader
          eyebrow="Architecture"
          title="How container SSH works"
          subtitle="The ShellHub agent runs alongside your containers on the Docker host, routing SSH connections to individual containers through the gateway."
        />

        <Reveal>
          <ShimmerCard>
            <Card hover className="p-6 sm:p-8 overflow-hidden">
              <ArchitectureDiagram />
            </Card>
          </ShimmerCard>
        </Reveal>
      </Section>

      <Section>
        <SectionHeader
          eyebrow="The Problem"
          title="Why docker exec falls short"
          subtitle="Docker exec was designed for local debugging, not for production access at scale."
        />

        <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
          {painPoints.map((p, i) => (
            <InfoCard
              key={i}
              icon={ExclamationTriangleIcon}
              color={p.color}
              title={p.title}
              description={p.desc}
              layout="horizontal"
              delay={i * 0.06}
            />
          ))}
        </div>
      </Section>

      {/* ═══════ Key Features ═══════ */}
      <Section>
        <SectionHeader
          eyebrow="Features"
          title="Built for container workflows"
          subtitle="Everything you need to manage SSH access to containers at scale, from access control to session recording."
        />

        {/* Big feature card: Per-Container Access Control */}
        <Reveal className="mb-6">
          <ShimmerCard>
            <div className="relative bg-card border border-primary/30 rounded-xl overflow-hidden hover:border-primary/50 transition-all duration-300 shadow-[0_0_40px_rgba(102,122,204,0.1)]">
              <div className="absolute inset-0 bg-gradient-to-br from-primary/[0.06] via-transparent to-transparent pointer-events-none" />
              <div className="relative grid lg:grid-cols-2 gap-8 p-8">
                {/* Left: description */}
                <div className="flex flex-col justify-center">
                  <IconBadge color="primary" className="mb-4">
                    <UsersIcon
                      className="w-5 h-5"
                      style={{ color: C.primary }}
                    />
                  </IconBadge>
                  <h3 className="text-lg font-bold mb-2">
                    Per-Container Access Control
                  </h3>
                  <p className="text-sm text-text-secondary leading-relaxed mb-6">
                    Define who can access which containers with role-based
                    policies. Assign roles at the container level, not just the
                    host. Restrict shell access, enforce MFA, and audit every
                    connection attempt.
                  </p>
                  <ul className="space-y-2.5">
                    {[
                      "Container-level permissions, not host-level",
                      "Role-based policies per user or group",
                      "MFA enforcement per container",
                      "Deny-by-default access model",
                    ].map((item) => (
                      <FeatureListItem key={item} color="green">
                        {item}
                      </FeatureListItem>
                    ))}
                  </ul>
                </div>

                {/* Right: permissions table mockup */}
                <WindowChrome
                  variant="browser"
                  path="/containers"
                  bodyClassName="font-sans"
                >
                  {/* Table header */}
                  <div className="grid grid-cols-4 gap-2 px-5 py-2.5 border-b border-border text-2xs font-mono font-semibold text-text-muted uppercase tracking-wider">
                    <span>Container</span>
                    <span>User</span>
                    <span>Role</span>
                    <span>Access</span>
                  </div>

                  {/* Table rows */}
                  {permRows.map((row, i) => (
                    <div
                      key={i}
                      className="grid grid-cols-4 gap-2 px-5 py-3 border-b border-border last:border-b-0 hover:bg-white/[0.02] transition-colors"
                    >
                      <span className="text-xs font-mono text-text-primary truncate">
                        {row.container}
                      </span>
                      <span className="text-xs text-text-secondary truncate">
                        {row.user}
                      </span>
                      <span
                        className="px-2 py-0.5 text-2xs font-mono rounded-full w-fit border"
                        style={{
                          color: row.accent,
                          borderColor: `${row.accent}30`,
                          background: `${row.accent}10`,
                        }}
                      >
                        {row.role}
                      </span>
                      <span className="text-xs text-text-muted">
                        {row.level}
                      </span>
                    </div>
                  ))}

                  <div className="px-5 py-3 flex items-center justify-between border-t border-border">
                    <span className="text-2xs text-text-muted">
                      4 policies in production
                    </span>
                    <span className="text-2xs text-primary font-medium">
                      Manage &rarr;
                    </span>
                  </div>
                </WindowChrome>
              </div>
            </div>
          </ShimmerCard>
        </Reveal>

        <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
          {smallFeatures.map((f, i) => (
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

      {/* ═══════ How It Works ═══════ */}
      <Section>
        <SectionHeader
          eyebrow="How It Works"
          title="Three steps to container SSH"
          subtitle="Go from zero to SSH-accessible containers in minutes, not hours."
        />

        <div className="relative grid grid-cols-1 md:grid-cols-3 gap-6">
          {/* Connecting line (visible on md+) */}
          <div className="hidden md:block absolute top-[52px] left-[16.66%] right-[16.66%] h-[1px] z-0">
            <div
              className="w-full h-full"
              style={{
                background: `linear-gradient(to right, ${C.primary}40, ${C.cyan}40, ${C.green}40)`,
              }}
            />
          </div>

          {steps.map((s, i) => (
            <Reveal key={i} delay={i * 0.08}>
              <div className="relative text-center">
                {/* Step number circle */}
                <div
                  className="relative z-10 w-[60px] h-[60px] rounded-full mx-auto mb-6 flex items-center justify-center border text-lg font-bold font-mono"
                  style={{
                    background: `${s.color}12`,
                    borderColor: `${s.color}30`,
                    color: s.color,
                  }}
                >
                  {s.num}
                </div>
                <ShimmerCard>
                  <Card hover className="p-6">
                    <h4 className="text-sm font-semibold mb-2">{s.title}</h4>
                    <p className="text-xs text-text-secondary leading-relaxed">
                      {s.desc}
                    </p>
                  </Card>
                </ShimmerCard>
              </div>
            </Reveal>
          ))}
        </div>
      </Section>

      <CTABanner
        eyebrow="Ready to simplify container access?"
        title="Manage your containers remotely"
        subtitle="Install the ShellHub agent and get instant SSH access to containers on any remote host. Free to start, no credit card required."
        primaryAction={{ label: "Get Started Free", to: "/getting-started" }}
        secondaryAction={{ label: "Compare Plans", to: "/pricing" }}
        eyebrowColor="cyan"
        gradient={{ from: "accent-cyan", to: "primary" }}
      />
    </SiteLayout>
  );
}
