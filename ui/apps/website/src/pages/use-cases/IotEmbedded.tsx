import {
  BoltIcon,
  CheckIcon,
  CommandLineIcon,
  ComputerDesktopIcon,
  EyeIcon,
  PencilIcon,
  PlayCircleIcon,
  ShieldCheckIcon,
  ShieldExclamationIcon,
  SignalIcon,
  TagIcon,
} from "@heroicons/react/24/outline";
import {
  Badge,
  IconBadge,
  WindowChrome,
} from "@shellhub/design-system/primitives";
import { GlowOrbs } from "@shellhub/design-system/components";
import { SiteLayout } from "@/components/SiteLayout";
import {
  ActionButtonGroup,
  CTABanner,
  HighlightCard,
  InfoCard,
  Section,
  SectionHeader,
} from "@/components/marketing";
import { ArrowMarker } from "@/components/marketing/ArrowMarker";
import { FeatureListItem } from "@/components/marketing/FeatureListItem";
import { Reveal, ShimmerCard, ConnectionGrid } from "../landing/components";
import { C } from "../landing/constants";

const painPoints = [
  {
    color: C.yellow,
    title: "No public IP",
    desc: "IoT devices behind cellular, CGNAT, or private networks are unreachable with traditional SSH.",
    icon: SignalIcon,
  },
  {
    color: C.primary,
    title: "Fleet scale",
    desc: "Managing SSH keys and firewall rules across hundreds or thousands of devices is unsustainable.",
    icon: ComputerDesktopIcon,
  },
  {
    color: C.red,
    title: "Security gaps",
    desc: "Exposing SSH ports on embedded devices creates a massive attack surface you can't easily monitor.",
    icon: ShieldExclamationIcon,
  },
  {
    color: C.cyan,
    title: "No visibility",
    desc: "Without centralized logging, you have no idea who accessed which device or what they did.",
    icon: EyeIcon,
  },
];

const features = [
  {
    icon: ShieldCheckIcon,
    color: C.green,
    title: "Firewall Rules",
    desc: "Define who can access which devices with granular per-device, per-user, or per-tag firewall policies. Block unauthorized connections before they reach the agent.",
  },
  {
    icon: TagIcon,
    color: C.yellow,
    title: "Device Tags",
    desc: "Organize fleets by location, firmware version, hardware type, or customer. Apply firewall rules and access policies to entire groups at once.",
  },
  {
    icon: PlayCircleIcon,
    color: C.cyan,
    title: "Session Recording",
    desc: "Record every SSH session for compliance, debugging, and knowledge sharing. Replay exactly what happened on a remote device during an incident.",
  },
  {
    icon: PencilIcon,
    color: C.primary,
    title: "Audit Logging",
    desc: "Full audit trail of every connection, command, and configuration change. Know who accessed what, when, and from where for regulatory compliance.",
  },
];

const fleetDevices = [
  {
    name: "rpi-warehouse-01",
    ip: "10.0.12.4",
    status: "online",
    tags: [
      { label: "production", color: C.green },
      { label: "firmware-v2.1", color: C.primary },
    ],
  },
  {
    name: "jetson-cam-03",
    ip: "172.16.0.88",
    status: "online",
    tags: [
      { label: "staging", color: C.yellow },
      { label: "vision-ml", color: C.cyan },
    ],
  },
  {
    name: "bbb-sensor-07",
    ip: "192.168.1.42",
    status: "online",
    tags: [{ label: "production", color: C.green }],
  },
  {
    name: "rpi-gateway-12",
    ip: "10.0.12.19",
    status: "offline",
    tags: [
      { label: "production", color: C.green },
      { label: "firmware-v2.0", color: C.primary },
    ],
  },
];

const platforms = [
  { name: "Raspberry Pi", abbr: "RPi" },
  { name: "BeagleBone", abbr: "BB" },
  { name: "NVIDIA Jetson", abbr: "NJ" },
  { name: "Yocto", abbr: "Yo" },
  { name: "Buildroot", abbr: "BR" },
  { name: "Ubuntu Core", abbr: "UC" },
  { name: "Debian", abbr: "De" },
  { name: "Alpine", abbr: "Al" },
];

/* ═══════ Component ═══════ */

export default function IotEmbedded() {
  return (
    <SiteLayout>
      {/* ═══════ Hero ═══════ */}
      <section className="relative pt-32 pb-24 overflow-hidden">
        <ConnectionGrid />
        <GlowOrbs preset="section" tone="green" />

        <div className="max-w-7xl mx-auto px-8 relative z-10 text-center">
          <Reveal>
            <Badge shape="pill" color="green" className="mb-6 tracking-label">
              Use Case
            </Badge>
          </Reveal>
          <Reveal>
            <h1 className="text-[clamp(2rem,5vw,3.5rem)] font-bold tracking-[-0.03em] leading-[1.1] mb-6 max-w-4xl mx-auto">
              IoT &amp; Embedded{" "}
              <span className="bg-gradient-to-r from-accent-green via-accent-cyan to-primary bg-clip-text text-transparent">
                Device Management
              </span>
            </h1>
          </Reveal>
          <Reveal>
            <p className="text-sm text-text-secondary max-w-lg mx-auto leading-relaxed mb-10">
              Remotely access and manage fleets of IoT devices, Raspberry Pis,
              and embedded Linux systems — even behind CGNAT, cellular networks,
              or private VLANs.
            </p>
          </Reveal>
          <Reveal>
            <ActionButtonGroup
              primaryAction={{
                label: "Get Started Free",
                to: "/getting-started",
              }}
              secondaryAction={{ label: "View Pricing", to: "/pricing" }}
            />
          </Reveal>
        </div>
      </section>

      {/* ═══════ Fleet Overview Mockup (2-col) ═══════ */}
      <Section>
        <div className="grid lg:grid-cols-2 gap-12 items-center">
          <div>
            <SectionHeader
              align="left"
              size="sub"
              className="mb-8"
              eyebrowColor="green"
              eyebrow="Fleet Dashboard"
              title="Manage thousands of devices from one place"
              subtitle="See every device in your fleet at a glance. Filter by status, location, firmware version, or custom tags. No more SSH-ing into a jump server to check on a single sensor."
            />

            <div className="space-y-3">
              {[
                {
                  label: "Real-time status",
                  desc: "See which devices are online, offline, or pending across every location",
                },
                {
                  label: "Custom tagging",
                  desc: "Organize by firmware version, deployment site, hardware type, or any custom label",
                },
                {
                  label: "Bulk operations",
                  desc: "Select multiple devices to apply firewall rules, update tags, or revoke access",
                },
                {
                  label: "Search & filter",
                  desc: "Find any device instantly by hostname, IP, tag, or SUID",
                },
              ].map((cap, i) => (
                <Reveal key={i} delay={i * 0.04}>
                  <div className="flex items-start gap-3">
                    <CheckIcon
                      className="w-4 h-4 text-accent-green shrink-0 mt-0.5"
                      strokeWidth={2}
                    />
                    <div>
                      <p className="text-sm font-medium text-text-primary">
                        {cap.label}
                      </p>
                      <p className="text-xs text-text-secondary leading-relaxed">
                        {cap.desc}
                      </p>
                    </div>
                  </div>
                </Reveal>
              ))}
            </div>
          </div>

          {/* Fake dashboard panel */}
          <Reveal delay={0.1}>
            <ShimmerCard>
              <WindowChrome variant="browser" path="/devices">
                {/* Table header */}
                <div className="flex items-center gap-3 px-3 py-2 mb-1">
                  <span className="text-2xs text-text-muted font-mono uppercase tracking-wider w-5">
                    St
                  </span>
                  <span className="text-2xs text-text-muted font-mono uppercase tracking-wider flex-1">
                    Hostname
                  </span>
                  <span className="text-2xs text-text-muted font-mono uppercase tracking-wider w-24 hidden sm:block">
                    IP
                  </span>
                  <span className="text-2xs text-text-muted font-mono uppercase tracking-wider flex-1 text-right">
                    Tags
                  </span>
                </div>

                {/* Device rows */}
                <div className="space-y-2">
                  {fleetDevices.map((d, i) => (
                    <div
                      key={i}
                      className="flex items-center gap-3 p-3 bg-surface rounded-lg border border-border hover:border-border-light transition-colors duration-200"
                    >
                      <div
                        className="w-2 h-2 rounded-full shrink-0"
                        style={{
                          background:
                            d.status === "online" ? C.green : C.textMuted,
                        }}
                      />
                      <div className="flex-1 min-w-0">
                        <p className="text-xs font-mono font-medium truncate">
                          {d.name}
                        </p>
                      </div>
                      <span className="text-2xs text-text-muted font-mono w-24 hidden sm:block">
                        {d.ip}
                      </span>
                      <div className="flex flex-1 items-center gap-1.5 flex-wrap justify-end">
                        {d.tags.map((t, j) => (
                          <span
                            key={j}
                            className="px-1.5 py-0.5 text-2xs font-mono rounded-full border"
                            style={{
                              background: `${t.color}10`,
                              color: t.color,
                              borderColor: `${t.color}20`,
                            }}
                          >
                            {t.label}
                          </span>
                        ))}
                      </div>
                    </div>
                  ))}
                </div>

                {/* Footer */}
                <div className="mt-4 pt-4 border-t border-border flex items-center justify-between">
                  <div className="flex items-center gap-2">
                    <div
                      className="w-2 h-2 rounded-full"
                      style={{ background: C.green }}
                    />
                    <span className="text-2xs text-text-muted">
                      24 devices online
                    </span>
                  </div>
                  <span className="text-2xs text-primary font-medium">
                    View all &rarr;
                  </span>
                </div>
              </WindowChrome>
            </ShimmerCard>
          </Reveal>
        </div>
      </Section>

      <Section>
        <SectionHeader
          eyebrow="The Challenge"
          title="Why managing IoT devices remotely is hard"
          subtitle="Traditional SSH workflows break down when devices are distributed, constrained, and unreachable from the public internet."
        />

        <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
          {painPoints.map((p, i) => (
            <InfoCard
              key={i}
              icon={p.icon}
              color={p.color}
              title={p.title}
              description={p.desc}
              layout="horizontal"
              delay={i * 0.06}
            />
          ))}
        </div>
      </Section>

      {/* ═══════ Key Features — Mixed Layout ═══════ */}
      <Section>
        <SectionHeader
          eyebrow="Built for IoT"
          eyebrowColor="green"
          title="Features designed for embedded fleets"
          subtitle="Every feature in ShellHub was designed with constrained, distributed devices in mind."
        />

        {/* Big card: NAT Traversal with SVG diagram */}
        <Reveal>
          <ShimmerCard className="mb-4">
            <HighlightCard color="primary">
              <div className="relative grid lg:grid-cols-2 gap-8 p-8">
                {/* Left: text */}
                <div className="flex flex-col justify-center">
                  <div className="flex items-center gap-3 mb-4">
                    <IconBadge color="primary">
                      <BoltIcon
                        className="w-5 h-5"
                        style={{ color: C.primary }}
                      />
                    </IconBadge>
                    <Badge shape="pill" color="primary">
                      Core
                    </Badge>
                  </div>
                  <h3 className="text-lg font-bold mb-3">NAT Traversal</h3>
                  <p className="text-sm text-text-secondary leading-relaxed mb-4">
                    The ShellHub agent makes an outbound connection from the
                    device, so it works behind any NAT, CGNAT, or firewall
                    without opening inbound ports.
                  </p>
                  <ul className="space-y-2">
                    {[
                      "No public IP required on the device",
                      "Works behind cellular, satellite, and double-NAT",
                      "No VPN configuration or port forwarding needed",
                      "Persistent connection with auto-reconnect",
                    ].map((item) => (
                      <FeatureListItem key={item} color="green">
                        {item}
                      </FeatureListItem>
                    ))}
                  </ul>
                </div>

                {/* Right: SVG network diagram */}
                <div className="flex items-center justify-center">
                  <div className="w-full max-w-md">
                    <svg
                      viewBox="0 0 400 220"
                      fill="none"
                      className="w-full h-auto"
                    >
                      {/* Device box */}
                      <rect
                        x="10"
                        y="60"
                        width="100"
                        height="100"
                        rx="8"
                        fill={C.card}
                        stroke={C.border}
                        strokeWidth="1"
                      />
                      <text
                        x="60"
                        y="96"
                        textAnchor="middle"
                        className="text-[10px]"
                        fill={C.textSec}
                        fontFamily="monospace"
                      >
                        IoT Device
                      </text>
                      <text
                        x="60"
                        y="114"
                        textAnchor="middle"
                        className="text-[9px]"
                        fill={C.textMuted}
                        fontFamily="monospace"
                      >
                        10.0.0.42
                      </text>
                      {/* Device icon */}
                      <rect
                        x="42"
                        y="124"
                        width="36"
                        height="24"
                        rx="3"
                        fill="none"
                        stroke={C.green}
                        strokeWidth="1"
                        opacity="0.5"
                      />
                      <circle
                        cx="60"
                        cy="136"
                        r="2"
                        fill={C.green}
                        opacity="0.7"
                      />

                      {/* CGNAT wall */}
                      <rect
                        x="134"
                        y="30"
                        width="36"
                        height="160"
                        rx="4"
                        fill={C.red}
                        fillOpacity="0.08"
                        stroke={C.red}
                        strokeWidth="1"
                        strokeDasharray="4 3"
                        opacity="0.6"
                      />
                      <text
                        x="152"
                        y="22"
                        textAnchor="middle"
                        className="text-[9px]"
                        fill={C.red}
                        fontFamily="monospace"
                        opacity="0.8"
                      >
                        CGNAT
                      </text>

                      {/* Outbound arrow: device -> ShellHub (over the wall) */}
                      <path
                        d="M110 100 C 130 100, 130 55, 170 55 L 260 55"
                        stroke={C.green}
                        strokeWidth="1.5"
                        strokeLinecap="round"
                        markerEnd="url(#arrowGreen)"
                      />
                      <text
                        x="210"
                        y="46"
                        textAnchor="middle"
                        className="text-[8px]"
                        fill={C.green}
                        fontFamily="monospace"
                      >
                        outbound
                      </text>

                      {/* ShellHub cloud box */}
                      <rect
                        x="260"
                        y="30"
                        width="130"
                        height="100"
                        rx="8"
                        fill={C.primary}
                        fillOpacity="0.08"
                        stroke={C.primary}
                        strokeWidth="1"
                      />
                      <text
                        x="325"
                        y="62"
                        textAnchor="middle"
                        className="text-[10px]"
                        fill={C.primary}
                        fontFamily="monospace"
                        fontWeight="600"
                      >
                        ShellHub
                      </text>
                      <text
                        x="325"
                        y="78"
                        textAnchor="middle"
                        className="text-[9px]"
                        fill={C.textMuted}
                        fontFamily="monospace"
                      >
                        cloud relay
                      </text>
                      {/* Terminal icon inside */}
                      <rect
                        x="305"
                        y="88"
                        width="40"
                        height="28"
                        rx="3"
                        fill={C.surface}
                        stroke={C.border}
                        strokeWidth="1"
                      />
                      <text
                        x="312"
                        y="106"
                        className="text-[9px]"
                        fill={C.green}
                        fontFamily="monospace"
                      >
                        $_
                      </text>

                      {/* User arrow: user -> ShellHub */}
                      <path
                        d="M325 170 L 325 130"
                        stroke={C.cyan}
                        strokeWidth="1.5"
                        strokeLinecap="round"
                        markerEnd="url(#arrowCyan)"
                      />

                      {/* User box */}
                      <rect
                        x="275"
                        y="170"
                        width="100"
                        height="40"
                        rx="6"
                        fill={C.card}
                        stroke={C.border}
                        strokeWidth="1"
                      />
                      <text
                        x="325"
                        y="194"
                        textAnchor="middle"
                        className="text-[10px]"
                        fill={C.textSec}
                        fontFamily="monospace"
                      >
                        You (SSH)
                      </text>

                      <defs>
                        <ArrowMarker id="arrowGreen" fill={C.green} />
                        <ArrowMarker id="arrowCyan" fill={C.cyan} refX={4} />
                      </defs>
                    </svg>
                  </div>
                </div>
              </div>
            </HighlightCard>
          </ShimmerCard>
        </Reveal>

        <div className="grid grid-cols-1 md:grid-cols-2 gap-4 mb-4">
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

        {/* Highlighted card: Lightweight Agent */}
        <Reveal>
          <ShimmerCard className="">
            <HighlightCard color="accent-green">
              <div className="relative p-8">
                <div className="grid lg:grid-cols-[1fr_auto] gap-8 items-center">
                  <div>
                    <div className="flex items-center gap-3 mb-4">
                      <IconBadge color="green">
                        <CommandLineIcon
                          className="w-5 h-5"
                          style={{ color: C.green }}
                        />
                      </IconBadge>
                      <Badge shape="pill" color="green">
                        Agent
                      </Badge>
                    </div>
                    <h3 className="text-lg font-bold mb-3">
                      Lightweight Agent
                    </h3>
                    <p className="text-sm text-text-secondary leading-relaxed">
                      The ShellHub agent is a single static binary with minimal
                      resource footprint. It runs on anything from a Raspberry
                      Pi Zero to an industrial gateway — no runtime dependencies
                      needed.
                    </p>
                  </div>

                  {/* Specs panel */}
                  <div className="flex flex-row lg:flex-col gap-3 flex-wrap">
                    {[
                      {
                        label: "Binary size",
                        value: "< 10 MB",
                        color: C.green,
                      },
                      {
                        label: "Architectures",
                        value: "ARM / x86",
                        color: C.cyan,
                      },
                      {
                        label: "Reconnect",
                        value: "Automatic",
                        color: C.primary,
                      },
                      {
                        label: "Dependencies",
                        value: "None",
                        color: C.yellow,
                      },
                    ].map((spec, i) => (
                      <div
                        key={i}
                        className="flex items-center gap-3 px-4 py-3 bg-surface rounded-lg border border-border min-w-[160px]"
                      >
                        <div
                          className="w-1.5 h-8 rounded-full"
                          style={{ background: spec.color }}
                        />
                        <div>
                          <p className="text-2xs text-text-muted font-mono uppercase tracking-wider">
                            {spec.label}
                          </p>
                          <p
                            className="text-sm font-semibold font-mono"
                            style={{ color: spec.color }}
                          >
                            {spec.value}
                          </p>
                        </div>
                      </div>
                    ))}
                  </div>
                </div>
              </div>
            </HighlightCard>
          </ShimmerCard>
        </Reveal>
      </Section>

      {/* ═══════ Supported Platforms ═══════ */}
      <Section>
        <SectionHeader
          eyebrow="Compatibility"
          title="Runs on the platforms you use"
          size="sub"
          subtitle="The ShellHub agent supports any Linux-based system with an ARM or x86 processor. Here are some of the most popular platforms."
          className="mb-12"
        />

        <Reveal>
          <div className="flex flex-wrap justify-center gap-4">
            {platforms.map((p, i) => (
              <div
                key={i}
                className="flex items-center gap-3 px-5 py-3.5 bg-card border border-border rounded-xl hover:border-border-light transition-colors duration-300"
              >
                <div className="w-9 h-9 rounded-lg bg-surface border border-border flex items-center justify-center">
                  <span className="text-2xs font-mono font-bold text-primary">
                    {p.abbr}
                  </span>
                </div>
                <span className="text-sm font-medium text-text-primary">
                  {p.name}
                </span>
              </div>
            ))}
          </div>
        </Reveal>
      </Section>

      <CTABanner
        eyebrow="Ready to get started?"
        title="Start managing your IoT fleet today"
        subtitle="Deploy the ShellHub agent on your devices and get instant remote access — no network changes, no VPNs, no exposed ports."
        primaryAction={{ label: "Get Started Free", to: "/getting-started" }}
        secondaryAction={{
          label: "Contact Sales",
          href: "mailto:sales@shellhub.io",
        }}
        eyebrowColor="green"
        gradient={{ from: "accent-green", to: "primary" }}
      />
    </SiteLayout>
  );
}
