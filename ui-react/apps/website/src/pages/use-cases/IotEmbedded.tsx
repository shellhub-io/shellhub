import { useEffect, useState, useCallback } from "react";
import { Navbar } from "../landing/Navbar";
import { Footer } from "@shellhub/design-system/components";
import { Reveal, ShimmerCard, ConnectionGrid } from "../landing/components";
import { C } from "../landing/constants";

/* ═══════ Data ═══════ */

const painPoints = [
  {
    color: C.yellow,
    title: "No public IP",
    desc: "IoT devices behind cellular, CGNAT, or private networks are unreachable with traditional SSH.",
    icon: (
      <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke={C.yellow} strokeWidth="1.5" strokeLinecap="round">
        <path d="M18.364 5.636a9 9 0 0 1 0 12.728M5.636 18.364a9 9 0 0 1 0-12.728" />
        <path d="M15.536 8.464a5 5 0 0 1 0 7.072M8.464 15.536a5 5 0 0 1 0-7.072" />
        <line x1="12" y1="12" x2="12.01" y2="12" strokeWidth="3" />
      </svg>
    ),
  },
  {
    color: C.primary,
    title: "Fleet scale",
    desc: "Managing SSH keys and firewall rules across hundreds or thousands of devices is unsustainable.",
    icon: (
      <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke={C.primary} strokeWidth="1.5" strokeLinecap="round">
        <rect x="2" y="3" width="20" height="14" rx="2" />
        <path d="M8 21h8M12 17v4" />
      </svg>
    ),
  },
  {
    color: C.red,
    title: "Security gaps",
    desc: "Exposing SSH ports on embedded devices creates a massive attack surface you can't easily monitor.",
    icon: (
      <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke={C.red} strokeWidth="1.5" strokeLinecap="round">
        <path d="M12 22s8-4 8-10V5l-8-3-8 3v7c0 6 8 10 8 10z" />
        <line x1="15" y1="9" x2="9" y2="15" />
        <line x1="9" y1="9" x2="15" y2="15" />
      </svg>
    ),
  },
  {
    color: C.cyan,
    title: "No visibility",
    desc: "Without centralized logging, you have no idea who accessed which device or what they did.",
    icon: (
      <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke={C.cyan} strokeWidth="1.5" strokeLinecap="round">
        <path d="M1 12s4-8 11-8 11 8 11 8-4 8-11 8-11-8-11-8z" />
        <circle cx="12" cy="12" r="3" />
      </svg>
    ),
  },
];

const fleetDevices = [
  { name: "rpi-warehouse-01", ip: "10.0.12.4", status: "online", tags: [{ label: "production", color: C.green }, { label: "firmware-v2.1", color: C.primary }] },
  { name: "jetson-cam-03", ip: "172.16.0.88", status: "online", tags: [{ label: "staging", color: C.yellow }, { label: "vision-ml", color: C.cyan }] },
  { name: "bbb-sensor-07", ip: "192.168.1.42", status: "online", tags: [{ label: "production", color: C.green }] },
  { name: "rpi-gateway-12", ip: "10.0.12.19", status: "offline", tags: [{ label: "production", color: C.green }, { label: "firmware-v2.0", color: C.primary }] },
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
  const [navSolid, setNavSolid] = useState(false);
  const [mobileMenu, setMobileMenu] = useState(false);
  const handleScroll = useCallback(() => { setNavSolid(window.scrollY > 50); }, []);
  useEffect(() => { window.addEventListener("scroll", handleScroll, { passive: true }); return () => window.removeEventListener("scroll", handleScroll); }, [handleScroll]);

  return (
    <div className="min-h-screen bg-background text-text-primary font-sans antialiased overflow-x-hidden">
      <Navbar navSolid={navSolid} mobileMenu={mobileMenu} setMobileMenu={setMobileMenu} />

      {/* ═══════ Hero ═══════ */}
      <section className="relative pt-32 pb-24 overflow-hidden">
        <ConnectionGrid />
        <div className="absolute inset-0 bg-gradient-radial from-accent-green/8 via-transparent to-transparent pointer-events-none" />
        <div className="absolute top-1/3 left-1/4 w-96 h-96 bg-accent-green/6 rounded-full blur-3xl pointer-events-none" />

        <div className="max-w-7xl mx-auto px-8 relative z-10 text-center">
          <Reveal>
            <span className="inline-block px-3 py-1 text-2xs font-mono font-semibold uppercase tracking-[0.15em] bg-accent-green/10 text-accent-green border border-accent-green/20 rounded-full mb-6">
              Use Case
            </span>
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
              Remotely access and manage fleets of IoT devices, Raspberry Pis, and embedded Linux systems — even behind CGNAT, cellular networks, or private VLANs.
            </p>
          </Reveal>
          <Reveal>
            <div className="flex flex-col sm:flex-row items-center justify-center gap-3">
              <a
                href="/v2/getting-started"
                className="relative inline-flex items-center gap-2.5 px-8 py-3.5 text-[15px] font-semibold bg-primary text-[#111214] rounded-xl shadow-[0_0_20px_rgba(102,122,204,0.3)] hover:shadow-[0_0_32px_rgba(102,122,204,0.5)] hover:scale-[1.03] active:scale-[0.98] transition-all duration-300 group overflow-hidden"
              >
                <span className="absolute inset-0 bg-gradient-to-r from-white/0 via-white/25 to-white/0 translate-x-[-200%] group-hover:translate-x-[200%] transition-transform duration-700 ease-in-out" />
                <span className="relative">Get Started Free</span>
                <svg className="relative w-4 h-4 group-hover:translate-x-1 transition-transform duration-300" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2.5}>
                  <path strokeLinecap="round" strokeLinejoin="round" d="M13.5 4.5 21 12m0 0-7.5 7.5M21 12H3" />
                </svg>
              </a>
              <a
                href="/v2/pricing"
                className="inline-flex items-center gap-2 px-6 py-3.5 text-[15px] font-medium text-text-secondary border border-border rounded-xl hover:text-text-primary hover:border-border-light hover:bg-white/[0.04] transition-all duration-300"
              >
                View Pricing
              </a>
            </div>
          </Reveal>
        </div>
      </section>

      {/* ═══════ Fleet Overview Mockup (2-col) ═══════ */}
      <section className="py-24 border-t border-border">
        <div className="max-w-7xl mx-auto px-8">
          <div className="grid lg:grid-cols-2 gap-12 items-center">
            <div>
              <Reveal>
                <p className="text-2xs font-mono font-semibold uppercase tracking-[0.15em] text-accent-green mb-3">
                  Fleet Dashboard
                </p>
                <h2 className="text-[clamp(1.75rem,4vw,2.5rem)] font-bold tracking-[-0.03em] leading-tight mb-4">
                  Manage thousands of devices from one place
                </h2>
                <p className="text-sm text-text-secondary leading-relaxed mb-8">
                  See every device in your fleet at a glance. Filter by status, location, firmware version, or custom tags. No more SSH-ing into a jump server to check on a single sensor.
                </p>
              </Reveal>

              <div className="space-y-3">
                {[
                  { label: "Real-time status", desc: "See which devices are online, offline, or pending across every location" },
                  { label: "Custom tagging", desc: "Organize by firmware version, deployment site, hardware type, or any custom label" },
                  { label: "Bulk operations", desc: "Select multiple devices to apply firewall rules, update tags, or revoke access" },
                  { label: "Search & filter", desc: "Find any device instantly by hostname, IP, tag, or SUID" },
                ].map((cap, i) => (
                  <Reveal key={i} delay={i * 0.04}>
                    <div className="flex items-start gap-3">
                      <svg className="w-4 h-4 text-accent-green shrink-0 mt-0.5" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
                        <path strokeLinecap="round" strokeLinejoin="round" d="m4.5 12.75 6 6 9-13.5" />
                      </svg>
                      <div>
                        <p className="text-sm font-medium text-text-primary">{cap.label}</p>
                        <p className="text-xs text-text-secondary leading-relaxed">{cap.desc}</p>
                      </div>
                    </div>
                  </Reveal>
                ))}
              </div>
            </div>

            {/* Fake dashboard panel */}
            <Reveal delay={0.1}>
              <ShimmerCard className="bg-card border border-border rounded-xl overflow-hidden">
                <div className="p-6">
                  {/* Window chrome */}
                  <div className="flex items-center gap-2 mb-6">
                    <div className="w-3 h-3 rounded-full bg-accent-red/60" />
                    <div className="w-3 h-3 rounded-full bg-accent-yellow/60" />
                    <div className="w-3 h-3 rounded-full bg-accent-green/60" />
                    <span className="ml-2 text-2xs text-text-muted font-mono">Device Fleet</span>
                  </div>

                  {/* Table header */}
                  <div className="flex items-center gap-3 px-3 py-2 mb-1">
                    <span className="text-2xs text-text-muted font-mono uppercase tracking-wider w-5">St</span>
                    <span className="text-2xs text-text-muted font-mono uppercase tracking-wider flex-1">Hostname</span>
                    <span className="text-2xs text-text-muted font-mono uppercase tracking-wider w-24 hidden sm:block">IP</span>
                    <span className="text-2xs text-text-muted font-mono uppercase tracking-wider flex-1 text-right">Tags</span>
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
                          style={{ background: d.status === "online" ? C.green : C.textMuted }}
                        />
                        <div className="flex-1 min-w-0">
                          <p className="text-xs font-mono font-medium truncate">{d.name}</p>
                        </div>
                        <span className="text-2xs text-text-muted font-mono w-24 hidden sm:block">{d.ip}</span>
                        <div className="flex items-center gap-1.5 flex-wrap justify-end">
                          {d.tags.map((t, j) => (
                            <span
                              key={j}
                              className="px-1.5 py-0.5 text-2xs font-mono rounded-full border"
                              style={{ background: `${t.color}10`, color: t.color, borderColor: `${t.color}20` }}
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
                      <div className="w-2 h-2 rounded-full" style={{ background: C.green }} />
                      <span className="text-2xs text-text-muted">24 devices online</span>
                    </div>
                    <span className="text-2xs text-primary font-medium">View all &rarr;</span>
                  </div>
                </div>
              </ShimmerCard>
            </Reveal>
          </div>
        </div>
      </section>

      {/* ═══════ Pain Points (2x2) ═══════ */}
      <section className="py-24 border-t border-border">
        <div className="max-w-7xl mx-auto px-8">
          <Reveal className="text-center mb-14">
            <p className="text-2xs font-mono font-semibold uppercase tracking-[0.15em] text-[#7B8EDB] mb-3">
              The Challenge
            </p>
            <h2 className="text-[clamp(1.75rem,4vw,3rem)] font-bold tracking-[-0.03em] leading-tight mb-4">
              Why managing IoT devices remotely is hard
            </h2>
            <p className="text-sm text-text-secondary max-w-lg mx-auto leading-relaxed">
              Traditional SSH workflows break down when devices are distributed, constrained, and unreachable from the public internet.
            </p>
          </Reveal>

          <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
            {painPoints.map((p, i) => (
              <Reveal key={i} delay={i * 0.06}>
                <ShimmerCard className="bg-card border border-border rounded-xl p-6 hover:border-border-light transition-all duration-300 h-full">
                  <div className="flex items-start gap-4">
                    <div
                      className="w-10 h-10 rounded-lg flex items-center justify-center shrink-0 border"
                      style={{ background: `${p.color}15`, borderColor: `${p.color}25` }}
                    >
                      {p.icon}
                    </div>
                    <div>
                      <h4 className="text-sm font-semibold mb-1.5">{p.title}</h4>
                      <p className="text-xs text-text-secondary leading-relaxed">{p.desc}</p>
                    </div>
                  </div>
                </ShimmerCard>
              </Reveal>
            ))}
          </div>
        </div>
      </section>

      {/* ═══════ Key Features — Mixed Layout ═══════ */}
      <section className="py-24 border-t border-border">
        <div className="max-w-7xl mx-auto px-8">
          <Reveal className="text-center mb-14">
            <p className="text-2xs font-mono font-semibold uppercase tracking-[0.15em] text-accent-green mb-3">
              Built for IoT
            </p>
            <h2 className="text-[clamp(1.75rem,4vw,3rem)] font-bold tracking-[-0.03em] leading-tight mb-4">
              Features designed for embedded fleets
            </h2>
            <p className="text-sm text-text-secondary max-w-lg mx-auto leading-relaxed">
              Every feature in ShellHub was designed with constrained, distributed devices in mind.
            </p>
          </Reveal>

          {/* Big card: NAT Traversal with SVG diagram */}
          <Reveal>
            <ShimmerCard className="mb-4">
              <div className="relative bg-card border border-primary/30 rounded-xl overflow-hidden shadow-[0_0_40px_rgba(102,122,204,0.1)] hover:border-primary/50 transition-all duration-300">
                <div className="absolute inset-0 bg-gradient-to-br from-primary/[0.06] via-transparent to-transparent pointer-events-none" />
                <div className="relative grid lg:grid-cols-2 gap-8 p-8">
                  {/* Left: text */}
                  <div className="flex flex-col justify-center">
                    <div className="flex items-center gap-3 mb-4">
                      <div className="w-10 h-10 rounded-lg bg-primary/10 border border-primary/20 flex items-center justify-center">
                        <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke={C.primary} strokeWidth="1.5" strokeLinecap="round">
                          <path d="M13 2L3 14h9l-1 8 10-12h-9l1-8z" />
                        </svg>
                      </div>
                      <span className="px-2 py-0.5 text-2xs font-mono font-semibold uppercase tracking-[0.1em] bg-primary/10 text-primary border border-primary/20 rounded-full">
                        Core
                      </span>
                    </div>
                    <h3 className="text-lg font-bold mb-3">NAT Traversal</h3>
                    <p className="text-sm text-text-secondary leading-relaxed mb-4">
                      The ShellHub agent makes an outbound connection from the device, so it works behind any NAT, CGNAT, or firewall without opening inbound ports.
                    </p>
                    <ul className="space-y-2">
                      {[
                        "No public IP required on the device",
                        "Works behind cellular, satellite, and double-NAT",
                        "No VPN configuration or port forwarding needed",
                        "Persistent connection with auto-reconnect",
                      ].map((item) => (
                        <li key={item} className="flex items-center gap-2.5 text-sm text-text-secondary">
                          <svg className="w-4 h-4 text-accent-green shrink-0" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
                            <path strokeLinecap="round" strokeLinejoin="round" d="m4.5 12.75 6 6 9-13.5" />
                          </svg>
                          {item}
                        </li>
                      ))}
                    </ul>
                  </div>

                  {/* Right: SVG network diagram */}
                  <div className="flex items-center justify-center">
                    <div className="w-full max-w-md">
                      <svg viewBox="0 0 400 220" fill="none" className="w-full h-auto">
                        {/* Device box */}
                        <rect x="10" y="60" width="100" height="100" rx="8" fill={C.card} stroke={C.border} strokeWidth="1" />
                        <text x="60" y="96" textAnchor="middle" className="text-[10px]" fill={C.textSec} fontFamily="monospace">IoT Device</text>
                        <text x="60" y="114" textAnchor="middle" className="text-[9px]" fill={C.textMuted} fontFamily="monospace">10.0.0.42</text>
                        {/* Device icon */}
                        <rect x="42" y="124" width="36" height="24" rx="3" fill="none" stroke={C.green} strokeWidth="1" opacity="0.5" />
                        <circle cx="60" cy="136" r="2" fill={C.green} opacity="0.7" />

                        {/* CGNAT wall */}
                        <rect x="134" y="30" width="36" height="160" rx="4" fill={C.red} fillOpacity="0.08" stroke={C.red} strokeWidth="1" strokeDasharray="4 3" opacity="0.6" />
                        <text x="152" y="22" textAnchor="middle" className="text-[9px]" fill={C.red} fontFamily="monospace" opacity="0.8">CGNAT</text>

                        {/* Outbound arrow: device -> ShellHub (over the wall) */}
                        <path d="M110 100 C 130 100, 130 55, 170 55 L 260 55" stroke={C.green} strokeWidth="1.5" strokeLinecap="round" markerEnd="url(#arrowGreen)" />
                        <text x="210" y="46" textAnchor="middle" className="text-[8px]" fill={C.green} fontFamily="monospace">outbound</text>

                        {/* ShellHub cloud box */}
                        <rect x="260" y="30" width="130" height="100" rx="8" fill={C.primary} fillOpacity="0.08" stroke={C.primary} strokeWidth="1" />
                        <text x="325" y="62" textAnchor="middle" className="text-[10px]" fill={C.primary} fontFamily="monospace" fontWeight="600">ShellHub</text>
                        <text x="325" y="78" textAnchor="middle" className="text-[9px]" fill={C.textMuted} fontFamily="monospace">cloud relay</text>
                        {/* Terminal icon inside */}
                        <rect x="305" y="88" width="40" height="28" rx="3" fill={C.surface} stroke={C.border} strokeWidth="1" />
                        <text x="312" y="106" className="text-[9px]" fill={C.green} fontFamily="monospace">$_</text>

                        {/* User arrow: user -> ShellHub */}
                        <path d="M325 170 L 325 130" stroke={C.cyan} strokeWidth="1.5" strokeLinecap="round" markerEnd="url(#arrowCyan)" />

                        {/* User box */}
                        <rect x="275" y="170" width="100" height="40" rx="6" fill={C.card} stroke={C.border} strokeWidth="1" />
                        <text x="325" y="194" textAnchor="middle" className="text-[10px]" fill={C.textSec} fontFamily="monospace">You (SSH)</text>

                        {/* Arrow markers */}
                        <defs>
                          <marker id="arrowGreen" markerWidth="8" markerHeight="6" refX="8" refY="3" orient="auto">
                            <polygon points="0 0, 8 3, 0 6" fill={C.green} />
                          </marker>
                          <marker id="arrowCyan" markerWidth="8" markerHeight="6" refX="4" refY="3" orient="auto">
                            <polygon points="0 0, 8 3, 0 6" fill={C.cyan} />
                          </marker>
                        </defs>
                      </svg>
                    </div>
                  </div>
                </div>
              </div>
            </ShimmerCard>
          </Reveal>

          {/* 2x2 grid of feature cards */}
          <div className="grid grid-cols-1 md:grid-cols-2 gap-4 mb-4">
            {[
              {
                icon: (
                  <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke={C.green} strokeWidth="1.5" strokeLinecap="round">
                    <path d="M12 22s8-4 8-10V5l-8-3-8 3v7c0 6 8 10 8 10z" />
                  </svg>
                ),
                color: C.green,
                title: "Firewall Rules",
                desc: "Define who can access which devices with granular per-device, per-user, or per-tag firewall policies. Block unauthorized connections before they reach the agent.",
              },
              {
                icon: (
                  <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke={C.yellow} strokeWidth="1.5" strokeLinecap="round">
                    <path d="M20.59 13.41l-7.17 7.17a2 2 0 01-2.83 0L2 12V2h10l8.59 8.59a2 2 0 010 2.82z" />
                    <line x1="7" y1="7" x2="7.01" y2="7" />
                  </svg>
                ),
                color: C.yellow,
                title: "Device Tags",
                desc: "Organize fleets by location, firmware version, hardware type, or customer. Apply firewall rules and access policies to entire groups at once.",
              },
              {
                icon: (
                  <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke={C.cyan} strokeWidth="1.5" strokeLinecap="round">
                    <circle cx="12" cy="12" r="10" />
                    <polygon points="10 8 16 12 10 16 10 8" />
                  </svg>
                ),
                color: C.cyan,
                title: "Session Recording",
                desc: "Record every SSH session for compliance, debugging, and knowledge sharing. Replay exactly what happened on a remote device during an incident.",
              },
              {
                icon: (
                  <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke={C.primary} strokeWidth="1.5" strokeLinecap="round">
                    <path d="M12 20h9" />
                    <path d="M16.5 3.5a2.121 2.121 0 013 3L7 19l-4 1 1-4L16.5 3.5z" />
                  </svg>
                ),
                color: C.primary,
                title: "Audit Logging",
                desc: "Full audit trail of every connection, command, and configuration change. Know who accessed what, when, and from where for regulatory compliance.",
              },
            ].map((f, i) => (
              <Reveal key={i} delay={i * 0.04}>
                <ShimmerCard className="bg-card border border-border rounded-xl p-6 hover:border-border-light transition-all duration-300 h-full">
                  <div
                    className="w-10 h-10 rounded-lg flex items-center justify-center mb-4 border"
                    style={{ background: `${f.color}15`, borderColor: `${f.color}25` }}
                  >
                    {f.icon}
                  </div>
                  <h4 className="text-sm font-semibold mb-2">{f.title}</h4>
                  <p className="text-xs text-text-secondary leading-relaxed">{f.desc}</p>
                </ShimmerCard>
              </Reveal>
            ))}
          </div>

          {/* Highlighted card: Lightweight Agent */}
          <Reveal>
            <ShimmerCard className="">
              <div className="relative bg-card border border-accent-green/30 rounded-xl overflow-hidden shadow-[0_0_40px_rgba(130,165,104,0.08)] hover:border-accent-green/50 transition-all duration-300">
                <div className="absolute inset-0 bg-gradient-to-br from-accent-green/[0.05] via-transparent to-transparent pointer-events-none" />
                <div className="relative p-8">
                  <div className="grid lg:grid-cols-[1fr_auto] gap-8 items-center">
                    <div>
                      <div className="flex items-center gap-3 mb-4">
                        <div className="w-10 h-10 rounded-lg bg-accent-green/10 border border-accent-green/20 flex items-center justify-center">
                          <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke={C.green} strokeWidth="1.5" strokeLinecap="round">
                            <polyline points="4 17 10 11 4 5" />
                            <line x1="12" y1="19" x2="20" y2="19" />
                          </svg>
                        </div>
                        <span className="px-2 py-0.5 text-2xs font-mono font-semibold uppercase tracking-[0.1em] bg-accent-green/10 text-accent-green border border-accent-green/20 rounded-full">
                          Agent
                        </span>
                      </div>
                      <h3 className="text-lg font-bold mb-3">Lightweight Agent</h3>
                      <p className="text-sm text-text-secondary leading-relaxed">
                        The ShellHub agent is a single static binary with minimal resource footprint. It runs on anything from a Raspberry Pi Zero to an industrial gateway — no runtime dependencies needed.
                      </p>
                    </div>

                    {/* Specs panel */}
                    <div className="flex flex-row lg:flex-col gap-3 flex-wrap">
                      {[
                        { label: "Binary size", value: "< 10 MB", color: C.green },
                        { label: "Architectures", value: "ARM / x86", color: C.cyan },
                        { label: "Reconnect", value: "Automatic", color: C.primary },
                        { label: "Dependencies", value: "None", color: C.yellow },
                      ].map((spec, i) => (
                        <div
                          key={i}
                          className="flex items-center gap-3 px-4 py-3 bg-surface rounded-lg border border-border min-w-[160px]"
                        >
                          <div className="w-1.5 h-8 rounded-full" style={{ background: spec.color }} />
                          <div>
                            <p className="text-2xs text-text-muted font-mono uppercase tracking-wider">{spec.label}</p>
                            <p className="text-sm font-semibold font-mono" style={{ color: spec.color }}>{spec.value}</p>
                          </div>
                        </div>
                      ))}
                    </div>
                  </div>
                </div>
              </div>
            </ShimmerCard>
          </Reveal>
        </div>
      </section>

      {/* ═══════ Supported Platforms ═══════ */}
      <section className="py-24 border-t border-border">
        <div className="max-w-7xl mx-auto px-8">
          <Reveal className="text-center mb-12">
            <p className="text-2xs font-mono font-semibold uppercase tracking-[0.15em] text-[#7B8EDB] mb-3">
              Compatibility
            </p>
            <h2 className="text-[clamp(1.75rem,4vw,2.5rem)] font-bold tracking-[-0.03em] leading-tight mb-4">
              Runs on the platforms you use
            </h2>
            <p className="text-sm text-text-secondary max-w-lg mx-auto leading-relaxed">
              The ShellHub agent supports any Linux-based system with an ARM or x86 processor. Here are some of the most popular platforms.
            </p>
          </Reveal>

          <Reveal>
            <div className="flex flex-wrap justify-center gap-4">
              {platforms.map((p, i) => (
                <div
                  key={i}
                  className="flex items-center gap-3 px-5 py-3.5 bg-card border border-border rounded-xl hover:border-border-light transition-colors duration-300"
                >
                  <div className="w-9 h-9 rounded-lg bg-surface border border-border flex items-center justify-center">
                    <span className="text-2xs font-mono font-bold text-primary">{p.abbr}</span>
                  </div>
                  <span className="text-sm font-medium text-text-primary">{p.name}</span>
                </div>
              ))}
            </div>
          </Reveal>
        </div>
      </section>

      {/* ═══════ CTA with ConnectionGrid ═══════ */}
      <section className="py-24 border-t border-border">
        <div className="max-w-7xl mx-auto px-8">
          <Reveal>
            <div className="relative bg-card border border-border rounded-2xl p-12 text-center overflow-hidden">
              <ConnectionGrid />
              <div className="absolute inset-0 bg-gradient-to-br from-accent-green/[0.06] via-transparent to-primary/[0.04] pointer-events-none" />

              <div className="relative z-10">
                <p className="text-2xs font-mono font-semibold uppercase tracking-[0.15em] text-accent-green mb-3">
                  Ready to get started?
                </p>
                <h2 className="text-[clamp(1.5rem,3vw,2.25rem)] font-bold tracking-[-0.03em] leading-tight mb-4">
                  Start managing your IoT fleet today
                </h2>
                <p className="text-sm text-text-secondary max-w-md mx-auto leading-relaxed mb-8">
                  Deploy the ShellHub agent on your devices and get instant remote access — no network changes, no VPNs, no exposed ports.
                </p>

                <div className="flex flex-col sm:flex-row items-center justify-center gap-3">
                  <a
                    href="/v2/getting-started"
                    className="relative inline-flex items-center gap-2.5 px-8 py-3.5 text-[15px] font-semibold bg-primary text-[#111214] rounded-xl shadow-[0_0_20px_rgba(102,122,204,0.3)] hover:shadow-[0_0_32px_rgba(102,122,204,0.5)] hover:scale-[1.03] active:scale-[0.98] transition-all duration-300 group overflow-hidden"
                  >
                    <span className="absolute inset-0 bg-gradient-to-r from-white/0 via-white/25 to-white/0 translate-x-[-200%] group-hover:translate-x-[200%] transition-transform duration-700 ease-in-out" />
                    <span className="relative">Get Started Free</span>
                    <svg className="relative w-4 h-4 group-hover:translate-x-1 transition-transform duration-300" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2.5}>
                      <path strokeLinecap="round" strokeLinejoin="round" d="M13.5 4.5 21 12m0 0-7.5 7.5M21 12H3" />
                    </svg>
                  </a>
                  <a
                    href="mailto:sales@shellhub.io"
                    className="inline-flex items-center gap-2 px-6 py-3.5 text-[15px] font-medium text-text-secondary border border-border rounded-xl hover:text-text-primary hover:border-border-light hover:bg-white/[0.04] transition-all duration-300"
                  >
                    Contact Sales
                  </a>
                </div>
              </div>
            </div>
          </Reveal>
        </div>
      </section>

      <Footer />
    </div>
  );
}
