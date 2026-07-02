import { Link } from "react-router-dom";
import {
  Badge,
  Button,
  Card,
  WindowChrome,
} from "@shellhub/design-system/primitives";
import { GlowOrbs } from "@shellhub/design-system/components";
import {
  CheckIcon,
  PlayCircleIcon,
  BackwardIcon,
  PauseIcon,
  ForwardIcon,
  LockClosedIcon,
  ShieldCheckIcon,
  UsersIcon,
  PencilSquareIcon,
  KeyIcon,
  FolderIcon,
  TagIcon,
  ChevronRightIcon,
} from "@heroicons/react/24/outline";
import { ArrowRight } from "@/components/ArrowRight";
import { SiteLayout } from "@/components/SiteLayout";
import { Section, SectionHeader } from "@/components/marketing";
import { Reveal, ShimmerCard, ConnectionGrid } from "../landing/components";
import { C } from "../landing/constants";

/* ─── Typed line helper ────────────────────────────────────────── */
function Line({
  prompt = "$",
  cmd,
  output,
  dimOutput = false,
}: {
  prompt?: string;
  cmd: string;
  output?: string;
  dimOutput?: boolean;
}) {
  return (
    <div className="mb-1.5 last:mb-0">
      <div>
        <span className="text-accent-green">{prompt}</span>
        <span className="text-text-primary ml-2">{cmd}</span>
      </div>
      {output && (
        <div
          className={
            dimOutput ? "text-text-muted mt-0.5" : "text-text-secondary mt-0.5"
          }
        >
          {output}
        </div>
      )}
    </div>
  );
}

/* ═══════════════════════════════════════════════════════════════ */
/*  HERO                                                          */
/* ═══════════════════════════════════════════════════════════════ */
function Hero() {
  return (
    <section className="relative pt-32 pb-24 overflow-hidden">
      <ConnectionGrid />
      <GlowOrbs preset="section" tone="cyan" />

      <div className="max-w-7xl mx-auto px-8 relative z-10 text-center">
        <Reveal>
          <Badge shape="pill" color="primary" className="mb-6 tracking-label">
            Features
          </Badge>
        </Reveal>
        <Reveal>
          <h1 className="text-[clamp(2rem,5vw,3.5rem)] font-bold tracking-[-0.03em] leading-[1.1] mb-6 max-w-3xl mx-auto">
            Everything you need to{" "}
            <span className="bg-gradient-to-r from-primary via-accent-blue to-accent-cyan bg-clip-text text-transparent">
              manage remote devices
            </span>
          </h1>
        </Reveal>
        <Reveal>
          <p className="text-sm text-text-secondary max-w-lg mx-auto leading-relaxed mb-10">
            Native SSH access, session recording, web terminal, file transfer,
            and granular access control. A complete platform for secure remote
            device management.
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
              View Pricing
            </Button>
          </div>
        </Reveal>
      </div>
    </section>
  );
}

/* ═══════════════════════════════════════════════════════════════ */
/*  HEADLINE FEATURE: Native SSH Access                           */
/* ═══════════════════════════════════════════════════════════════ */
function NativeSSH() {
  return (
    <Section>
      <div className="grid lg:grid-cols-2 gap-12 items-center">
        {/* Text */}
        <div>
          <SectionHeader
            align="left"
            size="sub"
            className="mb-8"
            eyebrow="Core Feature"
            title="Native SSH access, no agents required"
            subtitle="Connect to any device using your existing SSH client. ShellHub works transparently with OpenSSH, PuTTY, and any SSH-compatible client. No proprietary plugins, no VPNs, no port forwarding."
          />

          <div className="space-y-3">
            {[
              {
                label: "Standard SSH protocol",
                desc: "Use ssh, scp, sftp with zero modifications to your workflow",
              },
              {
                label: "SSHID addressing",
                desc: "Connect via user@device.namespace format for clear device targeting",
              },
              {
                label: "No agent installation",
                desc: "Devices run a lightweight ShellHub agent — nothing on your workstation",
              },
              {
                label: "Works behind NAT",
                desc: "Reach devices behind firewalls, NAT, or CGNAT without port forwarding",
              },
            ].map((item, i) => (
              <Reveal key={i} delay={i * 0.04}>
                <div className="flex items-start gap-3">
                  <CheckIcon
                    strokeWidth={2}
                    className="w-4 h-4 text-accent-green shrink-0 mt-0.5"
                    aria-hidden="true"
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

        {/* Terminal Mockup */}
        <Reveal delay={0.1}>
          <ShimmerCard>
            <WindowChrome
              variant="terminal"
              title="Terminal — ssh"
              accent="green"
            >
              <Line cmd="ssh admin@rpi-gateway.production.shellhub" />
              <div className="my-3 px-3 py-2 bg-surface rounded border border-border">
                <span className="text-accent-green">Connected to</span>{" "}
                <span className="text-primary">rpi-gateway</span>
                <span className="text-text-muted"> (production)</span>
              </div>
              <Line cmd="ssh deploy@sensor-node-04.staging.shellhub" />
              <div className="my-3 px-3 py-2 bg-surface rounded border border-border">
                <span className="text-accent-green">Connected to</span>{" "}
                <span className="text-primary">sensor-node-04</span>
                <span className="text-text-muted"> (staging)</span>
              </div>
              <Line cmd="ssh root@edge-server.iot-fleet.shellhub" />
              <div className="mt-2 flex items-center gap-2">
                <div className="w-1.5 h-1.5 rounded-full bg-accent-green animate-pulse" />
                <span className="text-text-muted text-2xs">
                  Connecting via ShellHub gateway...
                </span>
              </div>
            </WindowChrome>
          </ShimmerCard>
        </Reveal>
      </div>
    </Section>
  );
}

/* ═══════════════════════════════════════════════════════════════ */
/*  HEADLINE FEATURE: Session Recording                           */
/* ═══════════════════════════════════════════════════════════════ */
function SessionRecording() {
  return (
    <Section>
      <div className="grid lg:grid-cols-2 gap-12 items-center">
        {/* Playback Mockup (left on this one for visual variety) */}
        <Reveal delay={0.1}>
          <ShimmerCard>
            <Card className="overflow-hidden">
              <div className="p-6">
                {/* Header bar */}
                <div className="flex items-center justify-between mb-5">
                  <div className="flex items-center gap-3">
                    <div className="w-8 h-8 rounded-lg bg-accent-cyan/15 border border-accent-cyan/20 flex items-center justify-center">
                      <PlayCircleIcon
                        className="w-4 h-4 text-accent-cyan"
                        aria-hidden="true"
                      />
                    </div>
                    <div>
                      <p className="text-xs font-semibold">Session #a7f3c2</p>
                      <p className="text-2xs text-text-muted font-mono">
                        admin@rpi-gateway &middot; production
                      </p>
                    </div>
                  </div>
                  <span className="px-2 py-0.5 text-2xs font-mono bg-accent-green/10 text-accent-green border border-accent-green/20 rounded-full">
                    Recorded
                  </span>
                </div>

                {/* Fake terminal playback area */}
                <div className="bg-[#15161A] rounded-lg border border-border p-4 font-mono text-xs mb-4">
                  <div className="text-text-muted mb-1">
                    <span className="text-accent-green">admin@rpi-gateway</span>
                    :<span className="text-accent-blue">~</span>$ systemctl
                    status nginx
                  </div>
                  <div className="text-text-secondary mb-1">
                    ● nginx.service - A high performance web server
                  </div>
                  <div className="text-text-secondary mb-1">
                    &nbsp;&nbsp;&nbsp;Active:{" "}
                    <span className="text-accent-green">active (running)</span>{" "}
                    since Mon 2026-02-14 09:31:04 UTC
                  </div>
                  <div className="text-text-secondary mb-1">
                    &nbsp;&nbsp;Process: 1247 ExecStartPre=/usr/sbin/nginx -t
                    (code=exited, status=0/SUCCESS)
                  </div>
                  <div className="text-text-muted mt-2">
                    <span className="text-accent-green">admin@rpi-gateway</span>
                    :<span className="text-accent-blue">~</span>${" "}
                    <span className="inline-block w-2 h-3.5 bg-text-primary/60 animate-pulse" />
                  </div>
                </div>

                {/* Playback controls */}
                <div className="flex items-center gap-3">
                  <button
                    type="button"
                    aria-label="Rewind"
                    className="w-8 h-8 rounded-lg bg-surface border border-border flex items-center justify-center hover:border-border-light transition-colors"
                  >
                    <BackwardIcon
                      className="w-4 h-4 text-text-secondary"
                      strokeWidth={2}
                      aria-hidden="true"
                    />
                  </button>
                  <button
                    type="button"
                    aria-label="Pause"
                    className="w-10 h-10 rounded-lg bg-accent-cyan/15 border border-accent-cyan/25 flex items-center justify-center hover:bg-accent-cyan/25 transition-colors"
                  >
                    <PauseIcon
                      className="w-5 h-5 text-accent-cyan"
                      strokeWidth={2}
                      aria-hidden="true"
                    />
                  </button>
                  <button
                    type="button"
                    aria-label="Fast forward"
                    className="w-8 h-8 rounded-lg bg-surface border border-border flex items-center justify-center hover:border-border-light transition-colors"
                  >
                    <ForwardIcon
                      className="w-4 h-4 text-text-secondary"
                      strokeWidth={2}
                      aria-hidden="true"
                    />
                  </button>

                  {/* Timeline */}
                  <div className="flex-1 mx-2">
                    <div className="h-1.5 bg-surface rounded-full overflow-hidden border border-border">
                      <div className="h-full w-[62%] bg-gradient-to-r from-accent-cyan to-primary rounded-full relative">
                        <div className="absolute right-0 top-1/2 -translate-y-1/2 w-3 h-3 bg-white rounded-full shadow-[0_0_8px_rgba(78,154,163,0.6)]" />
                      </div>
                    </div>
                  </div>

                  <span className="text-2xs text-text-muted font-mono whitespace-nowrap">
                    04:32 / 07:15
                  </span>
                </div>

                {/* Session metadata */}
                <div className="mt-4 pt-4 border-t border-border grid grid-cols-3 gap-4">
                  <div>
                    <p className="text-2xs text-text-muted mb-0.5">Duration</p>
                    <p className="text-xs font-mono font-medium">7m 15s</p>
                  </div>
                  <div>
                    <p className="text-2xs text-text-muted mb-0.5">Commands</p>
                    <p className="text-xs font-mono font-medium">23</p>
                  </div>
                  <div>
                    <p className="text-2xs text-text-muted mb-0.5">Size</p>
                    <p className="text-xs font-mono font-medium">1.2 MB</p>
                  </div>
                </div>
              </div>
            </Card>
          </ShimmerCard>
        </Reveal>

        {/* Text */}
        <div>
          <SectionHeader
            align="left"
            size="sub"
            className="mb-8"
            eyebrow="Compliance & Audit"
            title="Record and replay every session"
            subtitle="Automatically capture SSH sessions in real-time. Replay them later for security review, incident investigation, training, or compliance auditing. Every keystroke and output is preserved."
          />

          <div className="space-y-3">
            {[
              {
                label: "Real-time capture",
                desc: "Sessions are recorded as they happen with zero performance impact",
              },
              {
                label: "Full-fidelity playback",
                desc: "Replay sessions exactly as they occurred, including timing and output",
              },
              {
                label: "Searchable history",
                desc: "Find sessions by user, device, date, or namespace",
              },
              {
                label: "Compliance ready",
                desc: "Meet SOC 2, HIPAA, and PCI DSS audit requirements",
              },
            ].map((item, i) => (
              <Reveal key={i} delay={i * 0.04}>
                <div className="flex items-start gap-3">
                  <CheckIcon
                    strokeWidth={2}
                    className="w-4 h-4 text-accent-green shrink-0 mt-0.5"
                    aria-hidden="true"
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
  );
}

/* ═══════════════════════════════════════════════════════════════ */
/*  HEADLINE FEATURE: Web Terminal                                */
/* ═══════════════════════════════════════════════════════════════ */
function WebTerminal() {
  return (
    <Section>
      <div className="grid lg:grid-cols-2 gap-12 items-center">
        {/* Text */}
        <div>
          <SectionHeader
            align="left"
            size="sub"
            className="mb-8"
            eyebrow="Zero Install"
            title="Terminal in your browser"
            subtitle="Access any device directly from the ShellHub web UI. No SSH client needed. Perfect for quick troubleshooting from a tablet, a shared workstation, or when your usual tools are not available."
          />

          <div className="space-y-3">
            {[
              {
                label: "Full terminal emulation",
                desc: "Powered by xterm.js with WebGL rendering for native-like performance",
              },
              {
                label: "Works from any browser",
                desc: "Chrome, Firefox, Safari, Edge — desktop and mobile",
              },
              {
                label: "Secure WebSocket connection",
                desc: "End-to-end encrypted channel between browser and device",
              },
              {
                label: "Copy and paste support",
                desc: "Seamless clipboard integration for efficient workflows",
              },
            ].map((item, i) => (
              <Reveal key={i} delay={i * 0.04}>
                <div className="flex items-start gap-3">
                  <CheckIcon
                    strokeWidth={2}
                    className="w-4 h-4 text-accent-green shrink-0 mt-0.5"
                    aria-hidden="true"
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

        {/* Browser Mockup */}
        <Reveal delay={0.1}>
          <ShimmerCard>
            <WindowChrome
              variant="browser"
              path="/devices/rpi-gateway/terminal"
            >
              {/* Fake tabs row */}
              <div className="flex items-center gap-1 mb-4">
                <div className="px-3 py-1.5 bg-surface border border-border rounded-t-lg text-2xs font-mono text-text-primary flex items-center gap-2">
                  <div className="w-1.5 h-1.5 rounded-full bg-accent-green" />
                  rpi-gateway
                </div>
                <div className="px-3 py-1.5 bg-card border border-border/50 rounded-t-lg text-2xs font-mono text-text-muted flex items-center gap-2">
                  <div className="w-1.5 h-1.5 rounded-full bg-text-muted" />
                  sensor-node-04
                </div>
              </div>

              {/* Terminal area inside browser */}
              <div className="bg-[#15161A] rounded-lg border border-border p-4">
                <div className="text-text-muted mb-1.5">
                  <span className="text-accent-green">admin@rpi-gateway</span>:
                  <span className="text-accent-blue">~</span>$ docker ps
                  --format &quot;table {"{{.Names}}"}\t{"{{.Status}}"}&quot;
                </div>
                <div className="text-text-secondary mb-0.5">
                  NAMES&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;STATUS
                </div>
                <div className="text-text-secondary mb-0.5">
                  nginx-proxy&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;Up 3 days
                </div>
                <div className="text-text-secondary mb-0.5">
                  app-backend&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;Up 3 days
                </div>
                <div className="text-text-secondary mb-0.5">
                  redis-cache&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;Up 3 days
                </div>
                <div className="text-text-secondary mb-1.5">
                  postgres-db&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;Up 3 days
                </div>
                <div className="text-text-muted">
                  <span className="text-accent-green">admin@rpi-gateway</span>:
                  <span className="text-accent-blue">~</span>${" "}
                  <span className="inline-block w-2 h-3.5 bg-text-primary/60 animate-pulse" />
                </div>
              </div>

              {/* Status bar */}
              <div className="mt-3 flex items-center justify-between text-2xs text-text-muted">
                <div className="flex items-center gap-2">
                  <div className="w-1.5 h-1.5 rounded-full bg-accent-green" />
                  <span>Connected &middot; WebSocket</span>
                </div>
                <span className="font-mono">80x24</span>
              </div>
            </WindowChrome>
          </ShimmerCard>
        </Reveal>
      </div>
    </Section>
  );
}

/* ═══════════════════════════════════════════════════════════════ */
/*  SECURITY & ACCESS CONTROL GRID                                */
/* ═══════════════════════════════════════════════════════════════ */
const securityFeatures = [
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
    title: "Multi-Factor Authentication",
    desc: "Add TOTP-based MFA to SSH connections. Works with Google Authenticator, Authy, and any standards-compliant app.",
  },
  {
    icon: (
      <ShieldCheckIcon
        width="20"
        height="20"
        stroke={C.red}
        strokeWidth="1.5"
        aria-hidden="true"
      />
    ),
    color: C.red,
    title: "Firewall Rules",
    desc: "Control access with flexible rules. Allow or deny connections based on IP address, hostname, or user identity.",
  },
  {
    icon: (
      <UsersIcon
        width="20"
        height="20"
        stroke={C.primary}
        strokeWidth="1.5"
        aria-hidden="true"
      />
    ),
    color: C.primary,
    title: "Role-Based Access Control",
    desc: "Assign roles and permissions at namespace and device level. Owners, operators, and viewers with granular control.",
  },
  {
    icon: (
      <PencilSquareIcon
        width="20"
        height="20"
        stroke={C.green}
        strokeWidth="1.5"
        aria-hidden="true"
      />
    ),
    color: C.green,
    title: "Audit Logging",
    desc: "Complete audit trail of every connection, command, and configuration change. Export logs for compliance reporting.",
  },
  {
    icon: (
      <KeyIcon
        width="20"
        height="20"
        stroke={C.cyan}
        strokeWidth="1.5"
        aria-hidden="true"
      />
    ),
    color: C.cyan,
    title: "Public Key Authentication",
    desc: "Use SSH public keys alongside or instead of passwords. Manage authorized keys centrally through the dashboard.",
  },
  {
    icon: (
      <FolderIcon
        width="20"
        height="20"
        stroke={C.blue}
        strokeWidth="1.5"
        aria-hidden="true"
      />
    ),
    color: C.blue,
    title: "Namespaces",
    desc: "Isolate devices and teams into separate namespaces. Each with its own members, devices, and security policies.",
  },
];

function SecurityGrid() {
  return (
    <Section>
      <SectionHeader
        eyebrow="Security & Access Control"
        title="Defense in depth, built into every layer"
        subtitle="Multiple layers of authentication, authorization, and auditing to secure your fleet from the network edge to the terminal."
      />

      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
        {securityFeatures.map((f, i) => (
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

/* ═══════════════════════════════════════════════════════════════ */
/*  FILE TRANSFER                                                 */
/* ═══════════════════════════════════════════════════════════════ */
function FileTransfer() {
  return (
    <Section>
      <div className="grid lg:grid-cols-2 gap-12 items-center">
        {/* Terminal Mockup */}
        <Reveal delay={0.1}>
          <ShimmerCard>
            <WindowChrome
              variant="terminal"
              title="Terminal — scp / sftp"
              accent="cyan"
            >
              <div className="mb-4">
                <p className="text-text-muted text-2xs uppercase tracking-wider mb-2">
                  SCP &mdash; Copy files to device
                </p>
                <Line
                  prompt="$"
                  cmd="scp firmware-v2.4.bin admin@rpi-gateway.production.shellhub:/opt/firmware/"
                />
                <div className="mt-1 flex items-center gap-3">
                  <div className="flex-1 h-1.5 bg-surface rounded-full overflow-hidden border border-border">
                    <div className="h-full w-full bg-gradient-to-r from-accent-cyan to-primary rounded-full" />
                  </div>
                  <span className="text-accent-green text-2xs">100%</span>
                </div>
                <p className="text-text-muted text-2xs mt-1">
                  firmware-v2.4.bin &nbsp; 24.3 MB &nbsp; 12.1MB/s &nbsp; 00:02
                </p>
              </div>

              <div className="pt-4 border-t border-border">
                <p className="text-text-muted text-2xs uppercase tracking-wider mb-2">
                  SFTP &mdash; Interactive session
                </p>
                <Line prompt="sftp>" cmd="ls /var/log/" />
                <div className="text-text-secondary ml-0">
                  <p>
                    syslog &nbsp;&nbsp; auth.log &nbsp;&nbsp; kern.log
                    &nbsp;&nbsp; nginx/
                  </p>
                </div>
                <Line prompt="sftp>" cmd="get /var/log/syslog ./diagnostics/" />
                <p className="text-text-muted text-2xs mt-1">
                  Fetching /var/log/syslog to ./diagnostics/syslog
                </p>
                <p className="text-accent-green text-2xs">
                  /var/log/syslog &nbsp; 100% &nbsp; 847KB &nbsp; 4.2MB/s &nbsp;
                  00:00
                </p>
              </div>
            </WindowChrome>
          </ShimmerCard>
        </Reveal>

        {/* Text */}
        <div>
          <SectionHeader
            align="left"
            size="sub"
            className="mb-8"
            eyebrow="File Transfer"
            title="SCP and SFTP, built right in"
            subtitle="Transfer files to and from remote devices using the same SCP and SFTP commands you already know. Push firmware updates, pull log files, manage configurations — all through the ShellHub gateway with no extra setup."
          />

          <div className="space-y-3">
            {[
              {
                label: "Standard protocols",
                desc: "Works with any SCP/SFTP client, including WinSCP and FileZilla",
              },
              {
                label: "Bidirectional transfers",
                desc: "Push files to devices or pull files from them using familiar commands",
              },
              {
                label: "Large file support",
                desc: "Transfer firmware images, database dumps, and large log archives",
              },
              {
                label: "Secure by default",
                desc: "Files are encrypted in transit through the ShellHub tunnel",
              },
            ].map((item, i) => (
              <Reveal key={i} delay={i * 0.04}>
                <div className="flex items-start gap-3">
                  <CheckIcon
                    strokeWidth={2}
                    className="w-4 h-4 text-accent-green shrink-0 mt-0.5"
                    aria-hidden="true"
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
  );
}

/* ═══════════════════════════════════════════════════════════════ */
/*  DOCKER CONTAINER ACCESS                                       */
/* ═══════════════════════════════════════════════════════════════ */
function DockerAccess() {
  return (
    <Section>
      <Reveal>
        <ShimmerCard className="h-full">
          <div className="relative bg-card border border-primary/30 rounded-2xl overflow-hidden shadow-[0_0_60px_rgba(102,122,204,0.08)]">
            <div className="absolute inset-0 bg-gradient-to-br from-primary/[0.06] via-transparent to-accent-cyan/[0.03] pointer-events-none" />
            <div className="relative grid lg:grid-cols-2 gap-0">
              {/* Diagram side */}
              <div className="p-8 lg:p-10 border-b lg:border-b-0 lg:border-r border-border/50">
                <div className="mb-6">
                  <p className="text-2xs font-mono font-semibold uppercase tracking-label text-primary mb-3">
                    Docker Integration
                  </p>
                  <h2 className="text-[clamp(1.5rem,3vw,2rem)] font-bold tracking-[-0.03em] leading-tight mb-3">
                    SSH into containers directly
                  </h2>
                  <p className="text-sm text-text-secondary leading-relaxed">
                    Access Docker containers on remote hosts with the same SSH
                    workflow you use for VMs. No need for docker exec or SSH
                    into the host first.
                  </p>
                </div>

                {/* Architecture diagram */}
                <div className="mt-6">
                  <svg viewBox="0 0 400 220" className="w-full" fill="none">
                    {/* Your workstation */}
                    <rect
                      x="10"
                      y="10"
                      width="120"
                      height="50"
                      rx="8"
                      fill={`${C.surface}`}
                      stroke={C.border}
                      strokeWidth="1"
                    />
                    <text
                      x="70"
                      y="31"
                      textAnchor="middle"
                      fill={C.textSec}
                      fontSize="9"
                      fontFamily="monospace"
                    >
                      Your Workstation
                    </text>
                    <text
                      x="70"
                      y="46"
                      textAnchor="middle"
                      fill={C.primary}
                      fontSize="10"
                      fontWeight="600"
                      fontFamily="monospace"
                    >
                      ssh client
                    </text>

                    {/* Arrow */}
                    <line
                      x1="130"
                      y1="35"
                      x2="165"
                      y2="35"
                      stroke={C.primary}
                      strokeWidth="1.5"
                      strokeDasharray="4 3"
                    />
                    <polygon points="165,31 173,35 165,39" fill={C.primary} />

                    {/* ShellHub Gateway */}
                    <rect
                      x="175"
                      y="10"
                      width="110"
                      height="50"
                      rx="8"
                      fill={`${C.primaryDim}`}
                      stroke={C.primary}
                      strokeWidth="1"
                    />
                    <text
                      x="230"
                      y="31"
                      textAnchor="middle"
                      fill={C.textSec}
                      fontSize="9"
                      fontFamily="monospace"
                    >
                      ShellHub
                    </text>
                    <text
                      x="230"
                      y="46"
                      textAnchor="middle"
                      fill={C.primary}
                      fontSize="10"
                      fontWeight="600"
                      fontFamily="monospace"
                    >
                      Gateway
                    </text>

                    {/* Arrow down */}
                    <line
                      x1="230"
                      y1="60"
                      x2="230"
                      y2="85"
                      stroke={C.primary}
                      strokeWidth="1.5"
                      strokeDasharray="4 3"
                    />
                    <polygon points="226,85 230,93 234,85" fill={C.primary} />

                    {/* Remote Host */}
                    <rect
                      x="50"
                      y="95"
                      width="340"
                      height="115"
                      rx="10"
                      fill={`${C.surface}`}
                      stroke={C.border}
                      strokeWidth="1"
                    />
                    <text
                      x="70"
                      y="115"
                      fill={C.textMuted}
                      fontSize="9"
                      fontFamily="monospace"
                    >
                      Remote Host
                    </text>

                    {/* ShellHub Agent */}
                    <rect
                      x="160"
                      y="100"
                      width="140"
                      height="28"
                      rx="6"
                      fill={`${C.primaryDim}`}
                      stroke={`${C.primary}40`}
                      strokeWidth="1"
                    />
                    <text
                      x="230"
                      y="118"
                      textAnchor="middle"
                      fill={C.primary}
                      fontSize="9"
                      fontWeight="600"
                      fontFamily="monospace"
                    >
                      ShellHub Agent
                    </text>

                    {/* Container 1 */}
                    <rect
                      x="70"
                      y="140"
                      width="100"
                      height="55"
                      rx="6"
                      fill={`${C.cyanDim}`}
                      stroke={`${C.cyan}50`}
                      strokeWidth="1"
                    />
                    <text
                      x="120"
                      y="158"
                      textAnchor="middle"
                      fill={C.cyan}
                      fontSize="9"
                      fontWeight="600"
                      fontFamily="monospace"
                    >
                      nginx-proxy
                    </text>
                    <text
                      x="120"
                      y="176"
                      textAnchor="middle"
                      fill={C.textMuted}
                      fontSize="8"
                      fontFamily="monospace"
                    >
                      container
                    </text>
                    <text
                      x="120"
                      y="188"
                      textAnchor="middle"
                      fill={C.green}
                      fontSize="8"
                      fontFamily="monospace"
                    >
                      running
                    </text>

                    {/* Container 2 */}
                    <rect
                      x="185"
                      y="140"
                      width="100"
                      height="55"
                      rx="6"
                      fill={`${C.cyanDim}`}
                      stroke={`${C.cyan}50`}
                      strokeWidth="1"
                    />
                    <text
                      x="235"
                      y="158"
                      textAnchor="middle"
                      fill={C.cyan}
                      fontSize="9"
                      fontWeight="600"
                      fontFamily="monospace"
                    >
                      app-backend
                    </text>
                    <text
                      x="235"
                      y="176"
                      textAnchor="middle"
                      fill={C.textMuted}
                      fontSize="8"
                      fontFamily="monospace"
                    >
                      container
                    </text>
                    <text
                      x="235"
                      y="188"
                      textAnchor="middle"
                      fill={C.green}
                      fontSize="8"
                      fontFamily="monospace"
                    >
                      running
                    </text>

                    {/* Container 3 */}
                    <rect
                      x="300"
                      y="140"
                      width="75"
                      height="55"
                      rx="6"
                      fill={`${C.cyanDim}`}
                      stroke={`${C.cyan}50`}
                      strokeWidth="1"
                    />
                    <text
                      x="337"
                      y="158"
                      textAnchor="middle"
                      fill={C.cyan}
                      fontSize="9"
                      fontWeight="600"
                      fontFamily="monospace"
                    >
                      redis
                    </text>
                    <text
                      x="337"
                      y="176"
                      textAnchor="middle"
                      fill={C.textMuted}
                      fontSize="8"
                      fontFamily="monospace"
                    >
                      container
                    </text>
                    <text
                      x="337"
                      y="188"
                      textAnchor="middle"
                      fill={C.green}
                      fontSize="8"
                      fontFamily="monospace"
                    >
                      running
                    </text>

                    {/* Connection lines from agent to containers */}
                    <line
                      x1="195"
                      y1="128"
                      x2="120"
                      y2="140"
                      stroke={`${C.cyan}60`}
                      strokeWidth="1"
                      strokeDasharray="3 2"
                    />
                    <line
                      x1="230"
                      y1="128"
                      x2="235"
                      y2="140"
                      stroke={`${C.cyan}60`}
                      strokeWidth="1"
                      strokeDasharray="3 2"
                    />
                    <line
                      x1="265"
                      y1="128"
                      x2="337"
                      y2="140"
                      stroke={`${C.cyan}60`}
                      strokeWidth="1"
                      strokeDasharray="3 2"
                    />
                  </svg>
                </div>
              </div>

              {/* Terminal side */}
              <div className="p-8 lg:p-10 flex flex-col justify-center">
                <div className="bg-[#15161A] rounded-lg border border-border p-4 font-mono text-xs mb-6">
                  <p className="text-text-muted mb-2">
                    # Connect to a container on a remote host
                  </p>
                  <Line
                    prompt="$"
                    cmd="ssh admin@nginx-proxy.production.shellhub"
                  />
                  <div className="my-2 px-2 py-1.5 bg-surface/50 rounded border border-border/50">
                    <span className="text-accent-green text-2xs">
                      Connected to container
                    </span>{" "}
                    <span className="text-accent-cyan text-2xs font-semibold">
                      nginx-proxy
                    </span>
                  </div>
                  <div className="text-text-muted mt-2 mb-1">
                    <span className="text-accent-green">root@nginx-proxy</span>:
                    <span className="text-accent-blue">/</span># nginx -t
                  </div>
                  <div className="text-text-secondary text-2xs">
                    nginx: configuration file /etc/nginx/nginx.conf test is
                    successful
                  </div>
                </div>

                <div className="space-y-3">
                  {[
                    "Same SSH workflow — no docker exec needed",
                    "Access containers behind NAT and firewalls",
                    "Full SCP/SFTP support for containers",
                    "Works with Docker, Podman, and containerd",
                  ].map((item, i) => (
                    <div
                      key={i}
                      className="flex items-center gap-2.5 text-sm text-text-secondary"
                    >
                      <CheckIcon
                        strokeWidth={2}
                        className="w-4 h-4 text-accent-green shrink-0"
                        aria-hidden="true"
                      />
                      {item}
                    </div>
                  ))}
                </div>
              </div>
            </div>
          </div>
        </ShimmerCard>
      </Reveal>
    </Section>
  );
}

/* ═══════════════════════════════════════════════════════════════ */
/*  DEVICE ORGANIZATION                                           */
/* ═══════════════════════════════════════════════════════════════ */
function DeviceOrganization() {
  return (
    <Section>
      <SectionHeader
        eyebrow="Device Organization"
        title="Tags and namespaces for fleet control"
        subtitle="Organize thousands of devices with tags for flexible grouping and namespaces for complete isolation between teams and environments."
      />

      <div className="grid md:grid-cols-2 gap-6">
        {/* Tags card */}
        <Reveal delay={0}>
          <ShimmerCard className="h-full">
            <Card hover className="p-8 h-full">
              <div className="flex items-center gap-3 mb-6">
                <div className="w-10 h-10 rounded-lg bg-primary/10 border border-primary/20 flex items-center justify-center">
                  <TagIcon
                    className="w-5 h-5 text-primary"
                    strokeWidth={1.5}
                    aria-hidden="true"
                  />
                </div>
                <div>
                  <h3 className="text-sm font-bold">Device Tags</h3>
                  <p className="text-2xs text-text-muted">
                    Flexible grouping & filtering
                  </p>
                </div>
              </div>

              {/* Tag mockup */}
              <div className="space-y-2.5 mb-6">
                {[
                  {
                    name: "rpi-gateway-01",
                    tags: [
                      { label: "production", color: C.green },
                      { label: "gateway", color: C.primary },
                      { label: "eu-west", color: C.blue },
                    ],
                  },
                  {
                    name: "sensor-node-04",
                    tags: [
                      { label: "staging", color: C.yellow },
                      { label: "sensor", color: C.cyan },
                    ],
                  },
                  {
                    name: "edge-server-12",
                    tags: [
                      { label: "production", color: C.green },
                      { label: "compute", color: C.primary },
                      { label: "us-east", color: C.blue },
                    ],
                  },
                ].map((device) => (
                  <div
                    key={device.name}
                    className="flex items-center justify-between p-3 bg-surface rounded-lg border border-border"
                  >
                    <div className="flex items-center gap-2.5">
                      <div className="w-1.5 h-1.5 rounded-full bg-accent-green" />
                      <span className="text-xs font-mono font-medium">
                        {device.name}
                      </span>
                    </div>
                    <div className="flex items-center gap-1.5">
                      {device.tags.map((tag) => (
                        <span
                          key={tag.label}
                          className="px-2 py-0.5 text-2xs font-mono rounded-full border"
                          style={{
                            background: `${tag.color}12`,
                            color: tag.color,
                            borderColor: `${tag.color}25`,
                          }}
                        >
                          {tag.label}
                        </span>
                      ))}
                    </div>
                  </div>
                ))}
              </div>

              <ul className="space-y-2">
                {[
                  "Filter and group devices by custom tags",
                  "Apply firewall rules based on tags",
                  "Bulk operations on tagged groups",
                ].map((item) => (
                  <li
                    key={item}
                    className="flex items-center gap-2.5 text-xs text-text-secondary"
                  >
                    <CheckIcon
                      strokeWidth={2}
                      className="w-3.5 h-3.5 text-accent-green shrink-0"
                      aria-hidden="true"
                    />
                    {item}
                  </li>
                ))}
              </ul>
            </Card>
          </ShimmerCard>
        </Reveal>

        {/* Namespaces card */}
        <Reveal delay={0.1}>
          <ShimmerCard className="h-full">
            <div className="relative bg-card border border-accent-cyan/30 rounded-xl p-8 h-full hover:border-accent-cyan/50 transition-all duration-300 shadow-[0_0_40px_rgba(78,154,163,0.08)] overflow-hidden">
              <div className="absolute inset-0 bg-gradient-to-br from-accent-cyan/[0.04] via-transparent to-transparent pointer-events-none" />
              <div className="relative">
                <div className="flex items-center gap-3 mb-6">
                  <div className="w-10 h-10 rounded-lg bg-accent-cyan/10 border border-accent-cyan/20 flex items-center justify-center">
                    <FolderIcon
                      className="w-5 h-5 text-accent-cyan"
                      strokeWidth={1.5}
                      aria-hidden="true"
                    />
                  </div>
                  <div>
                    <h3 className="text-sm font-bold">Namespaces</h3>
                    <p className="text-2xs text-accent-cyan">
                      Full multi-tenancy
                    </p>
                  </div>
                </div>

                {/* Namespace mockup */}
                <div className="space-y-2.5 mb-6">
                  {[
                    {
                      name: "production",
                      devices: 142,
                      members: 8,
                      color: C.green,
                    },
                    {
                      name: "staging",
                      devices: 36,
                      members: 12,
                      color: C.yellow,
                    },
                    {
                      name: "development",
                      devices: 18,
                      members: 5,
                      color: C.blue,
                    },
                  ].map((ns) => (
                    <div
                      key={ns.name}
                      className="flex items-center justify-between p-3 bg-surface rounded-lg border border-border"
                    >
                      <div className="flex items-center gap-2.5">
                        <div
                          className="w-2 h-6 rounded-sm"
                          style={{ background: ns.color }}
                        />
                        <div>
                          <p className="text-xs font-mono font-medium">
                            {ns.name}
                          </p>
                          <p className="text-2xs text-text-muted">
                            {ns.devices} devices &middot; {ns.members} members
                          </p>
                        </div>
                      </div>
                      <ChevronRightIcon
                        className="w-4 h-4 text-text-muted"
                        strokeWidth={1.5}
                        aria-hidden="true"
                      />
                    </div>
                  ))}
                </div>

                <ul className="space-y-2">
                  {[
                    "Complete isolation between environments",
                    "Separate members, devices, and policies",
                    "Independent firewall rules per namespace",
                  ].map((item) => (
                    <li
                      key={item}
                      className="flex items-center gap-2.5 text-xs text-text-secondary"
                    >
                      <CheckIcon
                        strokeWidth={2}
                        className="w-3.5 h-3.5 text-accent-green shrink-0"
                        aria-hidden="true"
                      />
                      {item}
                    </li>
                  ))}
                </ul>
              </div>
            </div>
          </ShimmerCard>
        </Reveal>
      </div>
    </Section>
  );
}

/* ═══════════════════════════════════════════════════════════════ */
/*  CTA                                                           */
/* ═══════════════════════════════════════════════════════════════ */
function FeaturesCTA() {
  return (
    <Section>
      <Reveal>
        <div className="relative bg-card border border-border rounded-2xl p-12 text-center overflow-hidden">
          <ConnectionGrid />
          <div className="absolute inset-0 bg-gradient-to-br from-primary/[0.06] via-transparent to-accent-cyan/[0.04] pointer-events-none" />

          <div className="relative z-10">
            <SectionHeader
              variant="cta"
              eyebrow="Ready to get started?"
              title="Deploy ShellHub in minutes"
              subtitle="Open-source and free to start. Run a single command and manage your first device in under five minutes."
            />

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
                View Pricing
              </Button>
            </div>
          </div>
        </div>
      </Reveal>
    </Section>
  );
}

/* ═══════════════════════════════════════════════════════════════ */
/*  PAGE                                                          */
/* ═══════════════════════════════════════════════════════════════ */
export default function Features() {
  return (
    <SiteLayout>
      <Hero />
      <NativeSSH />
      <SessionRecording />
      <WebTerminal />
      <SecurityGrid />
      <FileTransfer />
      <DockerAccess />
      <DeviceOrganization />
      <FeaturesCTA />
    </SiteLayout>
  );
}
