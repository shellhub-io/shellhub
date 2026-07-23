import {
  ArrowsPointingOutIcon,
  BoltIcon,
  CheckIcon,
  EnvelopeIcon,
  GlobeAltIcon,
  LockClosedIcon,
  ShieldCheckIcon,
  ShieldExclamationIcon,
  XMarkIcon,
} from "@heroicons/react/24/outline";
import {
  Badge,
  Card,
  IconBadge,
  WindowChrome,
} from "@shellhub/design-system/primitives";
import {
  ConnectionGrid,
  GlowOrbs,
  Reveal,
  ShimmerCard,
} from "@shellhub/design-system/components";
import {
  ActionButtonGroup,
  ArrowMarker,
  CTABanner,
  HighlightCard,
  InfoCard,
  Section,
  SectionHeader,
  SiteLayout,
} from "@/components";
import { C, FONT_SANS, FONT_MONO } from "@shellhub/design-system/constants";

const techDetails = [
  {
    icon: LockClosedIcon,
    color: C.primary,
    title: "TLS Encryption",
    desc: "All traffic between agents and the gateway is encrypted with TLS 1.3. No plaintext data ever leaves a device.",
  },
  {
    icon: EnvelopeIcon,
    color: C.cyan,
    title: "WebSocket Tunnels",
    desc: "Persistent WebSocket connections over port 443 ensure reliable communication even through restrictive proxies.",
  },
  {
    icon: ArrowsPointingOutIcon,
    color: C.green,
    title: "Reverse SSH",
    desc: "The agent initiates the connection outbound, then the gateway multiplexes inbound SSH sessions back through the same tunnel.",
  },
  {
    icon: ShieldCheckIcon,
    color: C.yellow,
    title: "NAT Traversal",
    desc: "Outbound-only connections mean devices behind any NAT, CGNAT, or carrier-grade firewall are reachable without port forwarding.",
  },
  {
    icon: GlobeAltIcon,
    color: C.blue,
    title: "Agent Auto-Update",
    desc: "Agents check for updates automatically and apply them without downtime. Always running the latest secure version.",
  },
  {
    icon: BoltIcon,
    color: C.red,
    title: "Lightweight Footprint",
    desc: "The agent binary is under 10 MB and uses minimal CPU and memory. Designed for constrained embedded devices.",
  },
];

const shellhubAdvantages = [
  { label: "No public IPs required", has: true },
  { label: "No firewall changes", has: true },
  { label: "No client software to install", has: true },
  { label: "Works behind CGNAT", has: true },
  { label: "Auto-reconnect on failure", has: true },
  { label: "Zero config on device network", has: true },
  { label: "Session recording built-in", has: true },
  { label: "Centralized audit logging", has: true },
];

const vpnLimitations = [
  { label: "Requires public IP or relay", has: false },
  { label: "Firewall rules per device", has: false },
  { label: "VPN client on every machine", has: false },
  { label: "Breaks behind CGNAT", has: false },
  { label: "Manual reconnection", has: false },
  { label: "Complex network configuration", has: false },
  { label: "Separate recording tools needed", has: false },
  { label: "Logging requires extra setup", has: false },
];

/* ------------------------------------------------------------------ */
/*  Page                                                               */
/* ------------------------------------------------------------------ */

export default function HowItWorks() {
  return (
    <SiteLayout>
      {/* ── Hero ─────────────────────────────────────────────────── */}
      <section className="relative pt-32 pb-24 overflow-hidden">
        <ConnectionGrid />
        <GlowOrbs preset="section" tone="cyan" />

        <div className="max-w-7xl mx-auto px-8 relative z-10 text-center">
          <Reveal>
            <Badge shape="pill" color="primary" className="mb-6 tracking-label">
              How It Works
            </Badge>
          </Reveal>
          <Reveal>
            <h1 className="text-[clamp(2rem,5vw,3.5rem)] font-bold tracking-[-0.03em] leading-[1.1] mb-6 max-w-3xl mx-auto">
              From anywhere to{" "}
              <span className="bg-gradient-to-r from-primary via-accent-blue to-accent-cyan bg-clip-text text-transparent">
                any device
              </span>
            </h1>
          </Reveal>
          <Reveal>
            <p className="text-sm text-text-secondary max-w-lg mx-auto leading-relaxed mb-10">
              ShellHub replaces VPNs, public IPs, and firewall rules with a
              single secure gateway. Install an agent, and SSH in from anywhere
              in the world.
            </p>
          </Reveal>
          <Reveal>
            <ActionButtonGroup
              primaryAction={{
                label: "Get Started Free",
                to: "/getting-started",
              }}
              secondaryAction={{
                label: "See the Architecture",
                href: "#architecture",
              }}
            />
          </Reveal>
        </div>
      </section>

      {/* ── Architecture Diagram ─────────────────────────────────── */}
      <Section id="architecture">
        <SectionHeader
          eyebrow="Architecture"
          title="The full picture, from user to device"
          subtitle="Every connection flows through the ShellHub gateway where it is authenticated, encrypted, and logged before reaching the target device."
        />

        <Reveal>
          <ShimmerCard>
            <Card className="p-6 lg:p-10 overflow-x-auto">
              <svg
                viewBox="0 0 960 340"
                fill="none"
                xmlns="http://www.w3.org/2000/svg"
                className="w-full h-auto min-w-[720px]"
              >
                <defs>
                  <ArrowMarker id="hw-a-pri" fill={C.primary} />
                  <ArrowMarker id="hw-a-grn" fill={C.green} />
                  <ArrowMarker id="hw-a-dim" fill={`${C.primary}60`} />
                </defs>

                {/* ── User ── */}
                <text
                  x="70"
                  y="24"
                  fontFamily={FONT_SANS}
                  fontSize="11"
                  fill={C.textMuted}
                  textAnchor="middle"
                  letterSpacing=".1em"
                >
                  YOU
                </text>
                <rect
                  x="20"
                  y="36"
                  width="100"
                  height="90"
                  rx="12"
                  fill={C.card}
                  stroke={C.border}
                  strokeWidth="1.2"
                />
                <circle
                  cx="70"
                  cy="62"
                  r="14"
                  stroke={C.primary}
                  strokeWidth="1.5"
                  fill="none"
                />
                <circle cx="70" cy="58" r="4.5" fill={C.primary} />
                <path
                  d="M70 63 L70 67 M63 65 L77 65"
                  stroke={C.primary}
                  strokeWidth="1.5"
                  strokeLinecap="round"
                />
                <text
                  x="70"
                  y="100"
                  fontFamily={FONT_MONO}
                  fontSize="9"
                  fill={C.textSec}
                  textAnchor="middle"
                >
                  Any location
                </text>
                <text
                  x="70"
                  y="114"
                  fontFamily={FONT_MONO}
                  fontSize="8"
                  fill={C.textMuted}
                  textAnchor="middle"
                >
                  laptop / browser
                </text>

                {/* ── Arrow: User to Gateway ── */}
                <line
                  x1="125"
                  y1="80"
                  x2="260"
                  y2="80"
                  stroke={C.primary}
                  strokeWidth="1.5"
                  strokeDasharray="6 4"
                  markerEnd="url(#hw-a-pri)"
                />
                <rect
                  x="155"
                  y="58"
                  width="68"
                  height="16"
                  rx="4"
                  fill={C.bg}
                />
                <text
                  x="189"
                  y="70"
                  fontFamily={FONT_MONO}
                  fontSize="8"
                  fill={C.primary}
                  textAnchor="middle"
                >
                  SSH / HTTPS
                </text>

                {/* ── ShellHub Cloud ── */}
                <rect
                  x="265"
                  y="20"
                  width="330"
                  height="290"
                  rx="16"
                  fill={`${C.primary}08`}
                  stroke={C.primary}
                  strokeWidth="1.5"
                />
                <text
                  x="430"
                  y="14"
                  fontFamily={FONT_SANS}
                  fontSize="12"
                  fill={C.primary}
                  textAnchor="middle"
                  fontWeight="600"
                  letterSpacing=".1em"
                >
                  SHELLHUB CLOUD
                </text>

                {/* Logo badge */}
                <rect
                  x="395"
                  y="38"
                  width="70"
                  height="36"
                  rx="10"
                  fill={C.primaryDim}
                  stroke={C.primary}
                  strokeWidth="1"
                />
                <text
                  x="430"
                  y="62"
                  fontFamily={FONT_MONO}
                  fontSize="16"
                  fill={C.primary}
                  textAnchor="middle"
                  fontWeight="700"
                >
                  SH
                </text>

                {/* Internal modules */}
                {[
                  {
                    x: 285,
                    y: 100,
                    label: "Authentication",
                    color: C.primary,
                    icon: "lock",
                  },
                  {
                    x: 285,
                    y: 140,
                    label: "TLS Encryption",
                    color: C.primary,
                    icon: "shield",
                  },
                  {
                    x: 285,
                    y: 180,
                    label: "Session Recording",
                    color: C.cyan,
                    icon: "rec",
                  },
                  {
                    x: 285,
                    y: 220,
                    label: "Connection Router",
                    color: C.green,
                    icon: "route",
                  },
                  {
                    x: 445,
                    y: 100,
                    label: "Firewall Rules",
                    color: C.yellow,
                    icon: "fire",
                  },
                  {
                    x: 445,
                    y: 140,
                    label: "Audit Logging",
                    color: C.green,
                    icon: "log",
                  },
                  {
                    x: 445,
                    y: 180,
                    label: "Team RBAC",
                    color: C.primary,
                    icon: "team",
                  },
                  {
                    x: 445,
                    y: 220,
                    label: "Device Registry",
                    color: C.blue,
                    icon: "device",
                  },
                ].map((m, i) => (
                  <g key={i}>
                    <rect
                      x={m.x}
                      y={m.y}
                      width="130"
                      height="30"
                      rx="6"
                      fill={C.card}
                      stroke={C.border}
                    />
                    <circle
                      cx={m.x + 16}
                      cy={m.y + 15}
                      r="5"
                      fill={`${m.color}30`}
                      stroke={m.color}
                      strokeWidth=".8"
                    />
                    <text
                      x={m.x + 32}
                      y={m.y + 19}
                      fontFamily={FONT_SANS}
                      fontSize="9"
                      fill={C.textSec}
                    >
                      {m.label}
                    </text>
                  </g>
                ))}

                {/* Gateway label */}
                <text
                  x="430"
                  y="278"
                  fontFamily={FONT_MONO}
                  fontSize="8"
                  fill={C.textMuted}
                  textAnchor="middle"
                >
                  CLOUD OR SELF-HOSTED
                </text>

                {/* ── NAT Wall ── */}
                <rect
                  x="630"
                  y="30"
                  width="8"
                  height="270"
                  rx="4"
                  fill={C.border}
                />
                <rect
                  x="630"
                  y="30"
                  width="8"
                  height="270"
                  rx="4"
                  fill={`${C.red}12`}
                />
                <text
                  x="634"
                  y="22"
                  fontFamily={FONT_MONO}
                  fontSize="9"
                  fill={C.textMuted}
                  textAnchor="middle"
                >
                  NAT
                </text>

                {/* ── Arrow: Gateway to NAT ── */}
                <line
                  x1="600"
                  y1="115"
                  x2="626"
                  y2="115"
                  stroke={C.green}
                  strokeWidth="1.5"
                  markerEnd="url(#hw-a-grn)"
                />

                {/* ── Arrows: NAT to Devices ── */}
                <line
                  x1="642"
                  y1="85"
                  x2="678"
                  y2="65"
                  stroke={`${C.primary}60`}
                  strokeWidth="1.2"
                  strokeDasharray="4 3"
                  markerEnd="url(#hw-a-dim)"
                />
                <line
                  x1="642"
                  y1="125"
                  x2="678"
                  y2="140"
                  stroke={`${C.primary}60`}
                  strokeWidth="1.2"
                  strokeDasharray="4 3"
                  markerEnd="url(#hw-a-dim)"
                />
                <line
                  x1="642"
                  y1="175"
                  x2="678"
                  y2="215"
                  stroke={`${C.primary}60`}
                  strokeWidth="1.2"
                  strokeDasharray="4 3"
                  markerEnd="url(#hw-a-dim)"
                />
                <line
                  x1="642"
                  y1="220"
                  x2="678"
                  y2="280"
                  stroke={`${C.primary}60`}
                  strokeWidth="1.2"
                  strokeDasharray="4 3"
                  markerEnd="url(#hw-a-dim)"
                />

                <text
                  x="658"
                  y="170"
                  fontFamily={FONT_MONO}
                  fontSize="7"
                  fill={`${C.primary}50`}
                  textAnchor="middle"
                  transform="rotate(-90,658,170)"
                >
                  NAT Traversal
                </text>

                {/* ── YOUR DEVICES label ── */}
                <text
                  x="790"
                  y="22"
                  fontFamily={FONT_SANS}
                  fontSize="11"
                  fill={C.textMuted}
                  textAnchor="middle"
                  letterSpacing=".1em"
                >
                  YOUR DEVICES
                </text>

                {/* ── Devices ── */}
                {[
                  {
                    y: 36,
                    icon: "Pi",
                    iconBg: C.green,
                    label: "Raspberry Pi",
                    sub: "armv7 / aarch64",
                    begin: "0s",
                  },
                  {
                    y: 112,
                    icon: "srv",
                    iconBg: C.primary,
                    label: "Linux Server",
                    sub: "Ubuntu / Debian / RHEL",
                    begin: ".5s",
                  },
                  {
                    y: 188,
                    icon: "dk",
                    iconBg: C.blue,
                    label: "Docker Host",
                    sub: "container agent",
                    begin: "1s",
                  },
                  {
                    y: 264,
                    icon: "iot",
                    iconBg: C.yellow,
                    label: "IoT Gateway",
                    sub: "OpenWrt / Yocto",
                    begin: "1.5s",
                  },
                ].map((d, i) => (
                  <g key={i}>
                    <rect
                      x="682"
                      y={d.y}
                      width="160"
                      height="60"
                      rx="10"
                      fill={C.card}
                      stroke={C.border}
                    />
                    {/* icon box */}
                    <rect
                      x="696"
                      y={d.y + 12}
                      width="28"
                      height="20"
                      rx="4"
                      fill={`${d.iconBg}15`}
                      stroke={d.iconBg}
                      strokeWidth=".8"
                    />
                    <text
                      x="710"
                      y={d.y + 26}
                      fontSize="9"
                      fill={d.iconBg}
                      textAnchor="middle"
                      fontFamily={FONT_MONO}
                      fontWeight="600"
                    >
                      {d.icon === "Pi"
                        ? "Pi"
                        : d.icon === "srv"
                          ? ">_"
                          : d.icon === "dk"
                            ? "dk"
                            : "IoT"}
                    </text>
                    {/* labels */}
                    <text
                      x="736"
                      y={d.y + 24}
                      fontFamily={FONT_SANS}
                      fontSize="10"
                      fill={C.text}
                    >
                      {d.label}
                    </text>
                    <text
                      x="736"
                      y={d.y + 38}
                      fontFamily={FONT_MONO}
                      fontSize="7"
                      fill={C.textMuted}
                    >
                      {d.sub}
                    </text>
                    <text
                      x="736"
                      y={d.y + 50}
                      fontFamily={FONT_MONO}
                      fontSize="7"
                      fill={C.textMuted}
                    >
                      agent running
                    </text>
                    {/* animated status dot */}
                    <circle cx="834" cy={d.y + 10} r="3.5" fill={C.green}>
                      <animate
                        attributeName="opacity"
                        values="1;.3;1"
                        dur="2s"
                        repeatCount="indefinite"
                        begin={d.begin}
                      />
                    </circle>
                  </g>
                ))}
              </svg>
            </Card>
          </ShimmerCard>
        </Reveal>
      </Section>

      {/* ── 3-Step Process (Vertical Timeline) ───────────────────── */}
      <Section>
        <SectionHeader
          eyebrow="Getting Started"
          title="Three steps to secure remote access"
          subtitle="From first install to first connection in under five minutes."
          className="mb-16"
        />

        <div className="relative max-w-4xl mx-auto">
          {/* Timeline line */}
          <div className="absolute left-8 lg:left-1/2 top-0 bottom-0 w-px bg-border hidden md:block" />

          {/* Step 1: Install Agent */}
          <Reveal>
            <div className="relative grid md:grid-cols-2 gap-8 mb-16">
              {/* Timeline dot */}
              <div className="absolute left-8 lg:left-1/2 top-8 w-3 h-3 -ml-1.5 rounded-full bg-primary border-2 border-background z-10 hidden md:block" />

              <div className="md:pr-12 md:text-right">
                <div className="flex items-center gap-3 mb-4 md:justify-end">
                  <span className="w-10 h-10 rounded-xl bg-primary/10 border border-primary/20 flex items-center justify-center text-sm font-bold font-mono text-primary">
                    01
                  </span>
                  <span className="text-2xs font-mono font-semibold uppercase tracking-[0.1em] text-primary">
                    Install
                  </span>
                </div>
                <h3 className="text-lg font-bold tracking-[-0.02em] mb-3">
                  Install the ShellHub Agent
                </h3>
                <p className="text-sm text-text-secondary leading-relaxed mb-4">
                  Deploy a lightweight agent on each device you want to manage.
                  One command works on any Linux system, Raspberry Pi,
                  container, or VM.
                </p>
                <ul className="space-y-2 md:ml-auto md:mr-0">
                  {[
                    "Single-line install script",
                    "Under 10 MB footprint",
                    "Runs as a system service",
                    "Auto-starts on boot",
                  ].map((d) => (
                    <li
                      key={d}
                      className="flex items-center gap-2 text-xs text-text-secondary md:justify-end"
                    >
                      <CheckIcon
                        className="w-3.5 h-3.5 text-primary shrink-0 md:order-last"
                        strokeWidth={2.5}
                      />
                      {d}
                    </li>
                  ))}
                </ul>
              </div>

              <div className="md:pl-12">
                <ShimmerCard>
                  <WindowChrome variant="terminal">
                    <div className="space-y-2 overflow-x-auto">
                      <p>
                        <span className="text-accent-green">$</span>{" "}
                        <span className="text-text-secondary">
                          curl -sSf https://cloud.shellhub.io/install.sh | sh
                        </span>
                      </p>
                      <p className="text-text-muted">
                        # Downloading ShellHub agent v0.17.2...
                      </p>
                      <p className="text-text-muted">
                        # Installing to /usr/local/bin/shellhub-agent
                      </p>
                      <p className="text-text-muted">
                        # Registering systemd service...
                      </p>
                      <p className="text-accent-green">
                        # Agent installed and running.
                      </p>
                      <p className="text-accent-green">
                        # Device ID: a1b2c3d4-e5f6-7890
                      </p>
                    </div>
                  </WindowChrome>
                </ShimmerCard>
              </div>
            </div>
          </Reveal>

          {/* Step 2: Agent Connects */}
          <Reveal>
            <div className="relative grid md:grid-cols-2 gap-8 mb-16">
              <div className="absolute left-8 lg:left-1/2 top-8 w-3 h-3 -ml-1.5 rounded-full bg-accent-cyan border-2 border-background z-10 hidden md:block" />

              <div className="md:order-last md:text-left md:pl-12">
                <div className="flex items-center gap-3 mb-4">
                  <span className="w-10 h-10 rounded-xl bg-accent-cyan/10 border border-accent-cyan/20 flex items-center justify-center text-sm font-bold font-mono text-accent-cyan">
                    02
                  </span>
                  <span className="text-2xs font-mono font-semibold uppercase tracking-[0.1em] text-accent-cyan">
                    Connect
                  </span>
                </div>
                <h3 className="text-lg font-bold tracking-[-0.02em] mb-3">
                  Agent Connects to ShellHub
                </h3>
                <p className="text-sm text-text-secondary leading-relaxed mb-4">
                  The agent initiates an outbound connection to the ShellHub
                  gateway. No inbound ports, no public IPs, no firewall changes
                  required on the device side.
                </p>
                <ul className="space-y-2">
                  {[
                    "Outbound-only connection (port 443)",
                    "Works behind NAT, firewalls, CGNAT",
                    "Automatic TLS encryption",
                    "Persistent WebSocket tunnel",
                  ].map((d) => (
                    <li
                      key={d}
                      className="flex items-center gap-2 text-xs text-text-secondary"
                    >
                      <CheckIcon
                        className="w-3.5 h-3.5 text-accent-cyan shrink-0"
                        strokeWidth={2.5}
                      />
                      {d}
                    </li>
                  ))}
                </ul>
              </div>

              <div className="md:pr-12">
                <ShimmerCard>
                  <Card className="overflow-hidden">
                    <div className="p-5">
                      <svg
                        viewBox="0 0 400 200"
                        fill="none"
                        xmlns="http://www.w3.org/2000/svg"
                        className="w-full h-auto"
                      >
                        <defs>
                          <ArrowMarker
                            id="hw-s2-a"
                            fill={C.cyan}
                            markerWidth={7}
                            markerHeight={5}
                            refX={7}
                            refY={2.5}
                          />
                        </defs>

                        {/* Device */}
                        <rect
                          x="20"
                          y="60"
                          width="90"
                          height="80"
                          rx="10"
                          fill={C.card}
                          stroke={C.border}
                        />
                        <rect
                          x="35"
                          y="74"
                          width="24"
                          height="16"
                          rx="3"
                          fill={`${C.green}15`}
                          stroke={C.green}
                          strokeWidth=".8"
                        />
                        <text
                          x="47"
                          y="86"
                          fontSize="8"
                          fill={C.green}
                          textAnchor="middle"
                          fontFamily={FONT_MONO}
                          fontWeight="600"
                        >
                          Pi
                        </text>
                        <text
                          x="65"
                          y="110"
                          fontFamily={FONT_SANS}
                          fontSize="9"
                          fill={C.textSec}
                          textAnchor="middle"
                        >
                          Device
                        </text>
                        <text
                          x="65"
                          y="124"
                          fontFamily={FONT_MONO}
                          fontSize="7"
                          fill={C.textMuted}
                          textAnchor="middle"
                        >
                          agent
                        </text>
                        <circle cx="100" cy="68" r="3" fill={C.green}>
                          <animate
                            attributeName="opacity"
                            values="1;.3;1"
                            dur="2s"
                            repeatCount="indefinite"
                          />
                        </circle>

                        {/* NAT Wall */}
                        <rect
                          x="145"
                          y="40"
                          width="6"
                          height="120"
                          rx="3"
                          fill={C.border}
                        />
                        <rect
                          x="145"
                          y="40"
                          width="6"
                          height="120"
                          rx="3"
                          fill={`${C.red}15`}
                        />
                        <text
                          x="148"
                          y="34"
                          fontFamily={FONT_MONO}
                          fontSize="8"
                          fill={C.textMuted}
                          textAnchor="middle"
                        >
                          NAT
                        </text>

                        {/* Outbound arrow: Device through NAT to Cloud */}
                        <line
                          x1="115"
                          y1="100"
                          x2="142"
                          y2="100"
                          stroke={C.cyan}
                          strokeWidth="1.5"
                          markerEnd="url(#hw-s2-a)"
                        />
                        <line
                          x1="154"
                          y1="100"
                          x2="225"
                          y2="100"
                          stroke={C.cyan}
                          strokeWidth="1.5"
                          strokeDasharray="6 4"
                          markerEnd="url(#hw-s2-a)"
                        />
                        <text
                          x="190"
                          y="92"
                          fontFamily={FONT_MONO}
                          fontSize="7"
                          fill={C.cyan}
                          textAnchor="middle"
                        >
                          outbound :443
                        </text>

                        {/* Cloud */}
                        <rect
                          x="230"
                          y="50"
                          width="150"
                          height="100"
                          rx="14"
                          fill={`${C.primary}08`}
                          stroke={C.primary}
                          strokeWidth="1.2"
                        />
                        <rect
                          x="278"
                          y="68"
                          width="54"
                          height="28"
                          rx="8"
                          fill={C.primaryDim}
                          stroke={C.primary}
                          strokeWidth=".8"
                        />
                        <text
                          x="305"
                          y="87"
                          fontFamily={FONT_MONO}
                          fontSize="12"
                          fill={C.primary}
                          textAnchor="middle"
                          fontWeight="700"
                        >
                          SH
                        </text>
                        <text
                          x="305"
                          y="118"
                          fontFamily={FONT_SANS}
                          fontSize="9"
                          fill={C.textSec}
                          textAnchor="middle"
                        >
                          ShellHub Gateway
                        </text>
                        <text
                          x="305"
                          y="132"
                          fontFamily={FONT_MONO}
                          fontSize="7"
                          fill={C.textMuted}
                          textAnchor="middle"
                        >
                          Auth + Encryption
                        </text>

                        {/* Labels */}
                        <text
                          x="65"
                          y="170"
                          fontFamily={FONT_MONO}
                          fontSize="7"
                          fill={C.textMuted}
                          textAnchor="middle"
                        >
                          PRIVATE NETWORK
                        </text>
                        <text
                          x="305"
                          y="170"
                          fontFamily={FONT_MONO}
                          fontSize="7"
                          fill={C.textMuted}
                          textAnchor="middle"
                        >
                          PUBLIC CLOUD
                        </text>
                      </svg>
                    </div>
                  </Card>
                </ShimmerCard>
              </div>
            </div>
          </Reveal>

          {/* Step 3: SSH from Anywhere */}
          <Reveal>
            <div className="relative grid md:grid-cols-2 gap-8">
              <div className="absolute left-8 lg:left-1/2 top-8 w-3 h-3 -ml-1.5 rounded-full bg-accent-green border-2 border-background z-10 hidden md:block" />

              <div className="md:pr-12 md:text-right">
                <div className="flex items-center gap-3 mb-4 md:justify-end">
                  <span className="w-10 h-10 rounded-xl bg-accent-green/10 border border-accent-green/20 flex items-center justify-center text-sm font-bold font-mono text-accent-green">
                    03
                  </span>
                  <span className="text-2xs font-mono font-semibold uppercase tracking-[0.1em] text-accent-green">
                    Access
                  </span>
                </div>
                <h3 className="text-lg font-bold tracking-[-0.02em] mb-3">
                  SSH from Anywhere
                </h3>
                <p className="text-sm text-text-secondary leading-relaxed mb-4">
                  Use your standard SSH client to connect through ShellHub. The
                  gateway authenticates the user, applies firewall rules, and
                  routes the connection to the right device.
                </p>
                <ul className="space-y-2 md:ml-auto md:mr-0">
                  {[
                    "Standard SSH client, no plugins",
                    "MFA and public key auth",
                    "Firewall rules per device/user",
                    "Session recording and audit trail",
                  ].map((d) => (
                    <li
                      key={d}
                      className="flex items-center gap-2 text-xs text-text-secondary md:justify-end"
                    >
                      <CheckIcon
                        className="w-3.5 h-3.5 text-accent-green shrink-0 md:order-last"
                        strokeWidth={2.5}
                      />
                      {d}
                    </li>
                  ))}
                </ul>
              </div>

              <div className="md:pl-12">
                <ShimmerCard>
                  <WindowChrome variant="terminal">
                    <div className="space-y-2 overflow-x-auto">
                      <p>
                        <span className="text-accent-green">$</span>{" "}
                        <span className="text-text-secondary">
                          ssh pi@a1b2c3d4-e5f6-7890.mycompany.shellhub.io
                        </span>
                      </p>
                      <p className="text-text-muted">
                        # Connecting through ShellHub gateway...
                      </p>
                      <p className="text-text-muted">
                        # Authenticating with public key...
                      </p>
                      <p className="text-text-muted">
                        # Session recording enabled.
                      </p>
                      <p>&nbsp;</p>
                      <p>
                        <span className="text-accent-green">
                          pi@raspberrypi
                        </span>
                        :<span className="text-accent-blue">~</span>
                        <span className="text-text-secondary">$ uname -a</span>
                      </p>
                      <p className="text-text-secondary">
                        Linux raspberrypi 6.1.0-rpi7 armv7l GNU/Linux
                      </p>
                      <p>
                        <span className="text-accent-green">
                          pi@raspberrypi
                        </span>
                        :<span className="text-accent-blue">~</span>
                        <span className="text-text-muted animate-pulse">_</span>
                      </p>
                    </div>
                  </WindowChrome>
                </ShimmerCard>
              </div>
            </div>
          </Reveal>
        </div>
      </Section>

      {/* ── Why Not VPN? (Comparison) ────────────────────────────── */}
      <Section>
        <SectionHeader
          eyebrow="Comparison"
          title="Why ShellHub over a VPN?"
          subtitle="VPNs were designed to link networks, not manage individual devices. ShellHub gives you direct, secure access without the overhead."
        />

        <div className="grid md:grid-cols-2 gap-6">
          {/* ShellHub Side (highlighted) */}
          <Reveal delay={0}>
            <ShimmerCard className="h-full">
              <HighlightCard
                color="primary"
                className="p-8 flex flex-col h-full"
              >
                <div className="relative">
                  <div className="flex items-center gap-3 mb-6">
                    <IconBadge color="primary">
                      <ShieldCheckIcon className="w-5 h-5 text-primary" />
                    </IconBadge>
                    <div>
                      <h3 className="text-sm font-bold">ShellHub</h3>
                      <p className="text-2xs text-primary">
                        Purpose-built for device access
                      </p>
                    </div>
                    <Badge shape="pill" color="green" className="ml-auto">
                      Recommended
                    </Badge>
                  </div>

                  <ul className="space-y-3">
                    {shellhubAdvantages.map((item) => (
                      <li
                        key={item.label}
                        className="flex items-center gap-2.5 text-sm text-text-secondary"
                      >
                        <CheckIcon
                          className="w-4 h-4 text-accent-green shrink-0"
                          strokeWidth={2}
                        />
                        {item.label}
                      </li>
                    ))}
                  </ul>
                </div>
              </HighlightCard>
            </ShimmerCard>
          </Reveal>

          {/* VPN Side */}
          <Reveal delay={0.1}>
            <ShimmerCard className="h-full">
              <Card hover className="p-8 flex flex-col h-full">
                <div className="flex items-center gap-3 mb-6">
                  <div className="w-10 h-10 rounded-lg bg-white/[0.04] border border-border flex items-center justify-center">
                    <ShieldExclamationIcon className="w-5 h-5 text-text-secondary" />
                  </div>
                  <div>
                    <h3 className="text-sm font-bold">Traditional VPN</h3>
                    <p className="text-2xs text-text-muted">
                      Network-level tunneling
                    </p>
                  </div>
                </div>

                <ul className="space-y-3">
                  {vpnLimitations.map((item) => (
                    <li
                      key={item.label}
                      className="flex items-center gap-2.5 text-sm text-text-secondary"
                    >
                      <XMarkIcon
                        className="w-4 h-4 text-accent-red/70 shrink-0"
                        strokeWidth={2}
                      />
                      {item.label}
                    </li>
                  ))}
                </ul>
              </Card>
            </ShimmerCard>
          </Reveal>
        </div>
      </Section>

      <Section>
        <SectionHeader
          eyebrow="Under the Hood"
          title="Built on solid foundations"
          subtitle="Every component of the stack is designed for security, reliability, and minimal overhead."
        />

        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
          {techDetails.map((f, i) => (
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
        eyebrow="Ready to try it?"
        title="See it in action"
        subtitle="Deploy ShellHub in under five minutes and connect to your first device. No credit card required."
        primaryAction={{ label: "Get Started Free", to: "/getting-started" }}
        secondaryAction={{ label: "View Pricing", to: "/pricing" }}
      />
    </SiteLayout>
  );
}
