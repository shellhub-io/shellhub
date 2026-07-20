import { Link } from "react-router-dom";
import {
  CommandLineIcon,
  ComputerDesktopIcon,
  DocumentArrowUpIcon,
  GlobeAltIcon,
  MapPinIcon,
  PencilSquareIcon,
  ShieldCheckIcon,
  SignalSlashIcon,
  TagIcon,
  UsersIcon,
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
import { Reveal, ShimmerCard, ConnectionGrid } from "../landing/components";
import { SiteLayout } from "@/components/SiteLayout";
import { CTABanner, Section, SectionHeader } from "@/components/marketing";
import { C } from "../landing/constants";

/* ═══════ Pain-point data ═══════ */
const painPoints = [
  {
    color: C.primary,
    title: "Distributed locations",
    desc: "Edge servers scattered across retail stores, warehouses, cell towers, and data centers — each with unique network topology.",
    icon: <GlobeAltIcon className="w-5 h-5" style={{ color: C.primary }} />,
  },
  {
    color: C.yellow,
    title: "Unreliable connectivity",
    desc: "Intermittent or low-bandwidth links make traditional VPN tunnels unstable and impossible to maintain reliably.",
    icon: <SignalSlashIcon className="w-5 h-5" style={{ color: C.yellow }} />,
  },
  {
    color: C.red,
    title: "On-site visits are expensive",
    desc: "Dispatching a technician to a remote cell tower or warehouse for a configuration change burns time and budget.",
    icon: <MapPinIcon className="w-5 h-5" style={{ color: C.red }} />,
  },
  {
    color: C.cyan,
    title: "Security at scale",
    desc: "Managing SSH keys, firewall rules, and access policies across hundreds of edge locations is operationally complex.",
    icon: <ShieldCheckIcon className="w-5 h-5" style={{ color: C.cyan }} />,
  },
];

/* ═══════ Edge locations for network map ═══════ */
const edgeLocations = [
  { label: "Retail Store", city: "NYC", x: 72, y: 28, color: C.primary },
  { label: "Warehouse", city: "Chicago", x: 25, y: 22, color: C.green },
  { label: "Cell Tower", city: "Austin", x: 30, y: 72, color: C.yellow },
  { label: "Data Center", city: "Denver", x: 78, y: 68, color: C.cyan },
  { label: "Branch Office", city: "Seattle", x: 15, y: 50, color: C.blue },
];

/* ═══════ Scenario data ═══════ */
const scenarios = [
  {
    color: C.primary,
    title: "Retail POS maintenance",
    desc: "Remotely troubleshoot and update point-of-sale terminals across hundreds of store locations without dispatching technicians.",
    mockup: (
      <div className="mt-4 bg-surface rounded-lg border border-border p-3 font-mono text-2xs">
        <div className="flex items-center gap-2 mb-2">
          <div className="w-2 h-2 rounded-full bg-accent-green animate-pulse" />
          <span style={{ color: C.textSec }}>pos-terminal-nyc-042</span>
        </div>
        <div style={{ color: C.textMuted }}>$ systemctl status pos-service</div>
        <div style={{ color: C.green }}>
          ● pos-service.service - POS Application
        </div>
        <div style={{ color: C.green }}>
          &nbsp;&nbsp;Active: active (running) since Mon
        </div>
        <div style={{ color: C.textMuted }}>
          $ pos-cli --update-config --store=042
        </div>
        <div style={{ color: C.primary }}>
          Config updated. Restarting service...
        </div>
        <div style={{ color: C.green }}>Done. POS terminal is back online.</div>
      </div>
    ),
  },
  {
    color: C.yellow,
    title: "Cell tower firmware update",
    desc: "Push firmware updates to radio controllers on cell towers over limited-bandwidth connections with resumable transfers.",
    mockup: (
      <div className="mt-4 bg-surface rounded-lg border border-border p-3 font-mono text-2xs">
        <div className="flex items-center gap-2 mb-2">
          <div className="w-2 h-2 rounded-full bg-accent-green animate-pulse" />
          <span style={{ color: C.textSec }}>tower-austin-017</span>
        </div>
        <div style={{ color: C.textMuted }}>
          $ scp firmware-v3.2.1.bin admin@tower-017:/opt/fw/
        </div>
        <div style={{ color: C.yellow }}>
          firmware-v3.2.1.bin&nbsp;&nbsp;78%&nbsp;&nbsp;14MB&nbsp;&nbsp;2.1MB/s&nbsp;&nbsp;00:03
        </div>
        <div style={{ color: C.green }}>
          firmware-v3.2.1.bin&nbsp;&nbsp;100%&nbsp;&nbsp;18MB&nbsp;&nbsp;2.3MB/s&nbsp;&nbsp;00:00
        </div>
        <div style={{ color: C.textMuted }}>$ fw-upgrade --apply --verify</div>
        <div style={{ color: C.green }}>
          Firmware v3.2.1 applied. Radio restarting...
        </div>
      </div>
    ),
  },
  {
    color: C.cyan,
    title: "Warehouse inventory server",
    desc: "Monitor and manage inventory database servers across warehouse locations with real-time status dashboards.",
    mockup: (
      <div className="mt-4 bg-surface rounded-lg border border-border p-3">
        <div className="flex items-center justify-between mb-3">
          <span className="text-2xs font-mono" style={{ color: C.textSec }}>
            inv-server-chicago
          </span>
          <span className="px-2 py-0.5 text-2xs font-mono bg-accent-green/10 text-accent-green border border-accent-green/20 rounded-full">
            Healthy
          </span>
        </div>
        <div className="grid grid-cols-3 gap-2">
          {[
            { label: "CPU", value: "23%", color: C.green },
            { label: "Memory", value: "4.2 GB", color: C.primary },
            { label: "Disk", value: "67%", color: C.yellow },
          ].map((m) => (
            <div
              key={m.label}
              className="bg-background rounded-md p-2 text-center"
            >
              <p className="text-2xs font-mono" style={{ color: C.textMuted }}>
                {m.label}
              </p>
              <p className="text-xs font-semibold" style={{ color: m.color }}>
                {m.value}
              </p>
            </div>
          ))}
        </div>
        <div className="mt-3 pt-2 border-t border-border">
          <div className="flex items-center justify-between text-2xs font-mono">
            <span style={{ color: C.textMuted }}>
              Last scan: 142,847 items indexed
            </span>
            <span style={{ color: C.green }}>Synced</span>
          </div>
        </div>
      </div>
    ),
  },
];

/* ═══════ Component ═══════ */
export default function EdgeComputing() {
  return (
    <SiteLayout>
      {/* ───── Hero ───── */}
      <section className="relative pt-32 pb-24 overflow-hidden">
        <ConnectionGrid />
        <GlowOrbs preset="section" tone="blue" />

        <div className="max-w-7xl mx-auto px-8 relative z-10 text-center">
          <Reveal>
            <Badge shape="pill" color="blue" className="mb-6 tracking-label">
              Use Case
            </Badge>
          </Reveal>
          <Reveal>
            <h1 className="text-[clamp(2rem,5vw,3.5rem)] font-bold tracking-[-0.03em] leading-[1.1] mb-6 max-w-3xl mx-auto">
              Edge Computing{" "}
              <span className="bg-gradient-to-r from-accent-blue via-primary to-accent-cyan bg-clip-text text-transparent">
                Access
              </span>
            </h1>
          </Reveal>
          <Reveal>
            <p className="text-sm text-text-secondary max-w-lg mx-auto leading-relaxed mb-10">
              Securely manage edge servers across retail stores, warehouses,
              cell towers, and remote sites — without VPNs, static IPs, or
              on-site visits.
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

      {/* ───── Edge Network Map ───── */}
      <Section>
        <SectionHeader
          eyebrow="Network Topology"
          title="One platform, every edge location"
          subtitle="ShellHub connects to all your distributed edge infrastructure through a single control plane — no matter where your devices are."
        />

        <Reveal delay={0.1}>
          <ShimmerCard className="bg-card border border-border rounded-2xl overflow-hidden">
            <div className="relative p-8 md:p-12">
              {/* Background grid pattern */}
              <div
                className="absolute inset-0 opacity-[0.03]"
                style={{
                  backgroundImage: `radial-gradient(${C.text} 1px, transparent 1px)`,
                  backgroundSize: "24px 24px",
                }}
              />

              <svg
                viewBox="0 0 100 100"
                className="w-full max-w-3xl mx-auto relative"
                style={{ minHeight: 320 }}
              >
                {/* Connection lines from center to each location */}
                {edgeLocations.map((loc, i) => (
                  <line
                    key={`line-${i}`}
                    x1="50"
                    y1="48"
                    x2={loc.x}
                    y2={loc.y}
                    stroke={loc.color}
                    strokeWidth="0.3"
                    strokeDasharray="1.5 1"
                    opacity="0.5"
                  />
                ))}

                {/* Central ShellHub Cloud node */}
                <g>
                  <circle
                    cx="50"
                    cy="48"
                    r="6"
                    fill={C.card}
                    stroke={C.primary}
                    strokeWidth="0.5"
                  />
                  <circle
                    cx="50"
                    cy="48"
                    r="8"
                    fill="none"
                    stroke={C.primary}
                    strokeWidth="0.15"
                    opacity="0.4"
                  />
                  <circle
                    cx="50"
                    cy="48"
                    r="10.5"
                    fill="none"
                    stroke={C.primary}
                    strokeWidth="0.1"
                    opacity="0.2"
                  />
                  {/* Cloud icon */}
                  <path
                    d="M46.5 49.5a2.5 2.5 0 0 1 2.5-3 3 3 0 0 1 5.5-1 2 2 0 0 1 .5 4h-7.5z"
                    fill="none"
                    stroke={C.primary}
                    strokeWidth="0.4"
                    strokeLinecap="round"
                    strokeLinejoin="round"
                  />
                  <text
                    x="50"
                    y="56"
                    textAnchor="middle"
                    fill={C.primary}
                    fontSize="2.2"
                    fontWeight="600"
                    fontFamily="monospace"
                  >
                    ShellHub Cloud
                  </text>
                </g>

                {/* Edge location nodes */}
                {edgeLocations.map((loc, i) => (
                  <g key={`node-${i}`}>
                    {/* Outer glow */}
                    <circle
                      cx={loc.x}
                      cy={loc.y}
                      r="4"
                      fill={`${loc.color}08`}
                      stroke={`${loc.color}30`}
                      strokeWidth="0.2"
                    />
                    {/* Server icon body */}
                    <rect
                      x={loc.x - 2}
                      y={loc.y - 2}
                      width="4"
                      height="4"
                      rx="0.6"
                      fill={C.surface}
                      stroke={loc.color}
                      strokeWidth="0.3"
                    />
                    {/* Server lines */}
                    <line
                      x1={loc.x - 1}
                      y1={loc.y - 0.5}
                      x2={loc.x + 1}
                      y2={loc.y - 0.5}
                      stroke={loc.color}
                      strokeWidth="0.2"
                      opacity="0.5"
                    />
                    <line
                      x1={loc.x - 1}
                      y1={loc.y + 0.5}
                      x2={loc.x + 1}
                      y2={loc.y + 0.5}
                      stroke={loc.color}
                      strokeWidth="0.2"
                      opacity="0.5"
                    />
                    {/* Green status dot */}
                    <circle
                      cx={loc.x + 1.2}
                      cy={loc.y - 1.2}
                      r="0.5"
                      fill={C.green}
                    >
                      <animate
                        attributeName="opacity"
                        values="1;0.4;1"
                        dur="2s"
                        begin={`${i * 0.4}s`}
                        repeatCount="indefinite"
                      />
                    </circle>
                    {/* Label */}
                    <text
                      x={loc.x}
                      y={loc.y + 5.5}
                      textAnchor="middle"
                      fill={C.text}
                      fontSize="1.8"
                      fontWeight="600"
                      fontFamily="sans-serif"
                    >
                      {loc.label}
                    </text>
                    <text
                      x={loc.x}
                      y={loc.y + 7.5}
                      textAnchor="middle"
                      fill={C.textMuted}
                      fontSize="1.5"
                      fontFamily="monospace"
                    >
                      {loc.city}
                    </text>
                  </g>
                ))}
              </svg>

              {/* Legend */}
              <div className="flex flex-wrap items-center justify-center gap-4 mt-6">
                {edgeLocations.map((loc) => (
                  <div key={loc.city} className="flex items-center gap-2">
                    <div
                      className="w-2 h-2 rounded-full"
                      style={{ background: loc.color }}
                    />
                    <span className="text-2xs font-mono text-text-muted">
                      {loc.label} - {loc.city}
                    </span>
                  </div>
                ))}
              </div>
            </div>
          </ShimmerCard>
        </Reveal>
      </Section>

      {/* ───── Pain Points ───── */}
      <Section>
        <SectionHeader
          eyebrow="Challenges"
          title="Managing edge infrastructure is hard"
          subtitle="Traditional remote access tools weren't built for the realities of distributed edge computing."
        />

        <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
          {painPoints.map((p, i) => (
            <Reveal key={i} delay={i * 0.06}>
              <ShimmerCard>
                <Card hover className="p-6 h-full">
                  <div className="flex items-start gap-4">
                    <div
                      className="w-10 h-10 rounded-lg flex items-center justify-center shrink-0 border"
                      style={{
                        background: `${p.color}15`,
                        borderColor: `${p.color}25`,
                      }}
                    >
                      {p.icon}
                    </div>
                    <div>
                      <h4 className="text-sm font-semibold mb-2">{p.title}</h4>
                      <p className="text-xs text-text-secondary leading-relaxed">
                        {p.desc}
                      </p>
                    </div>
                  </div>
                </Card>
              </ShimmerCard>
            </Reveal>
          ))}
        </div>
      </Section>

      {/* ───── Key Features ───── */}
      <Section>
        <SectionHeader
          eyebrow="Features"
          title="Built for the edge"
          subtitle="ShellHub eliminates the complexity of edge access with NAT traversal, browser-based terminals, and fleet-wide management."
        />

        {/* Big highlighted card: Instant Remote Access */}
        <Reveal>
          <ShimmerCard className="mb-4">
            <div className="relative bg-card border border-primary/30 rounded-xl p-8 overflow-hidden hover:border-primary/50 transition-all duration-300 shadow-[0_0_40px_rgba(102,122,204,0.1)]">
              <div className="absolute inset-0 bg-gradient-to-br from-primary/[0.06] via-transparent to-transparent pointer-events-none" />
              <div className="relative grid lg:grid-cols-2 gap-8 items-center">
                <div>
                  <div className="flex items-center gap-2 mb-4">
                    <IconBadge color="primary">
                      <CommandLineIcon
                        className="w-5 h-5"
                        style={{ color: C.primary }}
                      />
                    </IconBadge>
                    <Badge shape="pill" color="green">
                      Core Feature
                    </Badge>
                  </div>
                  <h3 className="text-lg font-bold mb-2">
                    Instant Remote Access
                  </h3>
                  <p className="text-sm text-text-secondary leading-relaxed">
                    Connect to any edge server in seconds — behind NAT, CGNAT,
                    or restrictive firewalls. No VPN configuration, no port
                    forwarding, no static IPs required. The agent handles
                    everything.
                  </p>
                </div>

                {/* Terminal mockup */}
                <WindowChrome variant="terminal" bodyClassName="space-y-1">
                  <div>
                    <span style={{ color: C.green }}>user@ops-laptop</span>
                    <span style={{ color: C.textMuted }}>:</span>
                    <span style={{ color: C.blue }}>~</span>
                    <span style={{ color: C.textMuted }}>$</span>{" "}
                    <span style={{ color: C.text }}>
                      ssh admin@edge-nyc-01.production
                    </span>
                  </div>
                  <div style={{ color: C.textMuted }}>
                    Connecting via ShellHub tunnel...
                  </div>
                  <div style={{ color: C.green }}>
                    Connection established (NAT traversal)
                  </div>
                  <div>&nbsp;</div>
                  <div>
                    <span style={{ color: C.green }}>admin@edge-nyc-01</span>
                    <span style={{ color: C.textMuted }}>:</span>
                    <span style={{ color: C.blue }}>~</span>
                    <span style={{ color: C.textMuted }}>$</span>{" "}
                    <span style={{ color: C.text }}>uname -a</span>
                  </div>
                  <div style={{ color: C.textSec }}>
                    Linux edge-nyc-01 5.15.0 #1 SMP x86_64
                  </div>
                  <div>
                    <span style={{ color: C.green }}>admin@edge-nyc-01</span>
                    <span style={{ color: C.textMuted }}>:</span>
                    <span style={{ color: C.blue }}>~</span>
                    <span style={{ color: C.textMuted }}>$</span>{" "}
                    <span style={{ color: C.text }}>
                      systemctl status edge-service
                    </span>
                  </div>
                  <div>
                    <span style={{ color: C.green }}>●</span>{" "}
                    <span style={{ color: C.textSec }}>
                      edge-service.service - Edge Compute Service
                    </span>
                  </div>
                  <div style={{ color: C.green }}>
                    &nbsp;&nbsp;Active: active (running) since Mon 2026-02-14
                  </div>
                  <div>
                    <span style={{ color: C.green }}>admin@edge-nyc-01</span>
                    <span style={{ color: C.textMuted }}>:</span>
                    <span style={{ color: C.blue }}>~</span>
                    <span style={{ color: C.textMuted }}>$</span>{" "}
                    <span
                      className="animate-pulse"
                      style={{ color: C.primary }}
                    >
                      _
                    </span>
                  </div>
                </WindowChrome>
              </div>
            </div>
          </ShimmerCard>
        </Reveal>

        {/* 2-column: Web Terminal + SCP/SFTP */}
        <div className="grid md:grid-cols-2 gap-4 mb-4">
          <Reveal delay={0.05}>
            <ShimmerCard className="h-full">
              <Card hover className="p-6 h-full">
                <div
                  className="w-10 h-10 rounded-lg flex items-center justify-center mb-4 border"
                  style={{
                    background: `${C.green}15`,
                    borderColor: `${C.green}25`,
                  }}
                >
                  <ComputerDesktopIcon
                    className="w-5 h-5"
                    style={{ color: C.green }}
                  />
                </div>
                <h4 className="text-sm font-semibold mb-2">Web Terminal</h4>
                <p className="text-xs text-text-secondary leading-relaxed mb-4">
                  Access edge servers from any browser — no SSH client needed.
                  Perfect for field engineers using tablets or shared
                  workstations.
                </p>

                {/* Browser mockup */}
                <WindowChrome
                  variant="browser"
                  size="sm"
                  path="/terminal/edge-chi-03"
                  bodyClassName="space-y-0.5"
                >
                  <div>
                    <span style={{ color: C.green }}>admin@edge-chi-03</span>
                    <span style={{ color: C.textMuted }}>:</span>
                    <span style={{ color: C.blue }}>~</span>
                    <span style={{ color: C.textMuted }}>$</span>{" "}
                    <span style={{ color: C.text }}>df -h /data</span>
                  </div>
                  <div style={{ color: C.textSec }}>
                    Filesystem&nbsp;&nbsp;Size&nbsp;&nbsp;Used&nbsp;&nbsp;Avail&nbsp;&nbsp;Use%
                  </div>
                  <div style={{ color: C.textSec }}>
                    /dev/sda1&nbsp;&nbsp;&nbsp;500G&nbsp;&nbsp;312G&nbsp;&nbsp;188G&nbsp;&nbsp;&nbsp;62%
                  </div>
                  <div>
                    <span style={{ color: C.green }}>admin@edge-chi-03</span>
                    <span style={{ color: C.textMuted }}>:</span>
                    <span style={{ color: C.blue }}>~</span>
                    <span style={{ color: C.textMuted }}>$</span>{" "}
                    <span
                      className="animate-pulse"
                      style={{ color: C.primary }}
                    >
                      _
                    </span>
                  </div>
                </WindowChrome>
              </Card>
            </ShimmerCard>
          </Reveal>

          <Reveal delay={0.1}>
            <ShimmerCard className="h-full">
              <Card hover className="p-6 h-full">
                <div
                  className="w-10 h-10 rounded-lg flex items-center justify-center mb-4 border"
                  style={{
                    background: `${C.cyan}15`,
                    borderColor: `${C.cyan}25`,
                  }}
                >
                  <DocumentArrowUpIcon
                    className="w-5 h-5"
                    style={{ color: C.cyan }}
                  />
                </div>
                <h4 className="text-sm font-semibold mb-2">
                  SCP / SFTP File Transfer
                </h4>
                <p className="text-xs text-text-secondary leading-relaxed mb-4">
                  Transfer configuration files, firmware updates, and logs to
                  and from edge servers with progress tracking.
                </p>

                {/* File transfer mockup */}
                <div className="bg-surface rounded-lg border border-border p-3 font-mono text-2xs space-y-1.5">
                  <div style={{ color: C.textMuted }}>
                    $ scp config.yml admin@edge-den-02:/etc/app/
                  </div>
                  <div className="flex items-center gap-2">
                    <span style={{ color: C.text }}>config.yml</span>
                    <div className="flex-1 h-1.5 bg-background rounded-full overflow-hidden">
                      <div
                        className="h-full rounded-full"
                        style={{ width: "100%", background: C.green }}
                      />
                    </div>
                    <span style={{ color: C.green }}>100%</span>
                  </div>
                  <div style={{ color: C.textSec }}>
                    2.4 KB&nbsp;&nbsp;&nbsp;0:00
                  </div>
                  <div
                    className="pt-1 border-t border-border"
                    style={{ borderColor: `${C.border}80` }}
                  >
                    <div style={{ color: C.textMuted }}>
                      $ scp admin@edge-den-02:/var/log/app.log ./
                    </div>
                    <div className="flex items-center gap-2 mt-1">
                      <span style={{ color: C.text }}>app.log</span>
                      <div className="flex-1 h-1.5 bg-background rounded-full overflow-hidden">
                        <div
                          className="h-full rounded-full"
                          style={{ width: "68%", background: C.yellow }}
                        />
                      </div>
                      <span style={{ color: C.yellow }}>68%</span>
                    </div>
                    <div style={{ color: C.textSec }}>
                      45 MB&nbsp;&nbsp;&nbsp;12.3 MB/s&nbsp;&nbsp;&nbsp;ETA 0:02
                    </div>
                  </div>
                </div>
              </Card>
            </ShimmerCard>
          </Reveal>
        </div>

        {/* 3 smaller cards: Tags, RBAC, Audit */}
        <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
          <Reveal delay={0.05}>
            <ShimmerCard className="h-full">
              <Card hover className="p-6 h-full">
                <div
                  className="w-10 h-10 rounded-lg flex items-center justify-center mb-4 border"
                  style={{
                    background: `${C.yellow}15`,
                    borderColor: `${C.yellow}25`,
                  }}
                >
                  <TagIcon className="w-5 h-5" style={{ color: C.yellow }} />
                </div>
                <h4 className="text-sm font-semibold mb-2">Device Tags</h4>
                <p className="text-xs text-text-secondary leading-relaxed mb-3">
                  Organize edge servers by region, site type, or function for
                  fast filtering and batch operations.
                </p>
                <div className="flex flex-wrap gap-1.5">
                  {[
                    { label: "region:northeast", color: C.primary },
                    { label: "type:retail", color: C.green },
                    { label: "env:production", color: C.yellow },
                    { label: "rack:A3", color: C.cyan },
                  ].map((tag) => (
                    <span
                      key={tag.label}
                      className="px-2 py-0.5 text-2xs font-mono rounded-full border"
                      style={{
                        color: tag.color,
                        background: `${tag.color}10`,
                        borderColor: `${tag.color}20`,
                      }}
                    >
                      {tag.label}
                    </span>
                  ))}
                </div>
              </Card>
            </ShimmerCard>
          </Reveal>

          <Reveal delay={0.1}>
            <ShimmerCard className="h-full">
              <Card hover className="p-6 h-full">
                <div
                  className="w-10 h-10 rounded-lg flex items-center justify-center mb-4 border"
                  style={{
                    background: `${C.primary}15`,
                    borderColor: `${C.primary}25`,
                  }}
                >
                  <UsersIcon className="w-5 h-5" style={{ color: C.primary }} />
                </div>
                <h4 className="text-sm font-semibold mb-2">RBAC</h4>
                <p className="text-xs text-text-secondary leading-relaxed mb-3">
                  Give regional teams access to only their edge servers with
                  role-based controls and namespace isolation.
                </p>
                <div className="space-y-1.5">
                  {[
                    {
                      role: "NYC Ops Team",
                      access: "northeast-*",
                      color: C.primary,
                    },
                    {
                      role: "Field Engineers",
                      access: "tower-*",
                      color: C.yellow,
                    },
                    {
                      role: "Warehouse Admins",
                      access: "warehouse-*",
                      color: C.green,
                    },
                  ].map((r) => (
                    <div
                      key={r.role}
                      className="flex items-center justify-between text-2xs font-mono"
                    >
                      <span style={{ color: C.textSec }}>{r.role}</span>
                      <span style={{ color: r.color }}>{r.access}</span>
                    </div>
                  ))}
                </div>
              </Card>
            </ShimmerCard>
          </Reveal>

          <Reveal delay={0.15}>
            <ShimmerCard className="h-full">
              <Card hover className="p-6 h-full">
                <div
                  className="w-10 h-10 rounded-lg flex items-center justify-center mb-4 border"
                  style={{
                    background: `${C.green}15`,
                    borderColor: `${C.green}25`,
                  }}
                >
                  <PencilSquareIcon
                    className="w-5 h-5"
                    style={{ color: C.green }}
                  />
                </div>
                <h4 className="text-sm font-semibold mb-2">Audit Trail</h4>
                <p className="text-xs text-text-secondary leading-relaxed mb-3">
                  Full session recording and command logging across all edge
                  locations for compliance and forensics.
                </p>
                <div className="space-y-1.5">
                  {[
                    {
                      time: "14:32",
                      user: "jane",
                      device: "edge-nyc-01",
                      action: "SSH session",
                      color: C.green,
                    },
                    {
                      time: "14:28",
                      user: "mike",
                      device: "tower-aus-03",
                      action: "File transfer",
                      color: C.cyan,
                    },
                    {
                      time: "14:15",
                      user: "ana",
                      device: "wh-chi-07",
                      action: "SSH session",
                      color: C.green,
                    },
                  ].map((log, i) => (
                    <div
                      key={i}
                      className="flex items-center gap-2 text-2xs font-mono"
                    >
                      <span style={{ color: C.textMuted }}>{log.time}</span>
                      <span style={{ color: C.primary }}>{log.user}</span>
                      <span style={{ color: C.textMuted }}>&rarr;</span>
                      <span style={{ color: C.textSec }}>{log.device}</span>
                      <span
                        className="ml-auto px-1.5 py-0.5 rounded border"
                        style={{
                          color: log.color,
                          background: `${log.color}10`,
                          borderColor: `${log.color}20`,
                        }}
                      >
                        {log.action}
                      </span>
                    </div>
                  ))}
                </div>
              </Card>
            </ShimmerCard>
          </Reveal>
        </div>
      </Section>

      {/* ───── Use Case Scenarios ───── */}
      <Section>
        <SectionHeader
          eyebrow="Real-World Scenarios"
          title="How teams use ShellHub at the edge"
          subtitle="From retail to telecom to logistics, ShellHub powers remote access for edge infrastructure across industries."
        />

        <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
          {scenarios.map((s, i) => (
            <Reveal key={i} delay={i * 0.08}>
              <ShimmerCard className="h-full">
                <Card hover className="p-6 h-full">
                  <div
                    className="w-2 h-2 rounded-full mb-4"
                    style={{ background: s.color }}
                  />
                  <h4 className="text-sm font-semibold mb-2">{s.title}</h4>
                  <p className="text-xs text-text-secondary leading-relaxed">
                    {s.desc}
                  </p>
                  {s.mockup}
                </Card>
              </ShimmerCard>
            </Reveal>
          ))}
        </div>
      </Section>

      <CTABanner
        eyebrow="Ready to connect your edge?"
        title="Your edge servers, instantly accessible"
        subtitle="Install the lightweight agent on your edge servers and start managing them remotely in minutes — no infrastructure changes needed."
        primaryAction={{ label: "Get Started Free", to: "/getting-started" }}
        secondaryAction={{ label: "Compare Plans", to: "/pricing" }}
        gradient={{ from: "accent-blue", to: "primary" }}
      />
    </SiteLayout>
  );
}
