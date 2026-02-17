import { useEffect, useState, useCallback } from "react";
import { Navbar } from "../landing/Navbar";
import { Footer } from "@shellhub/design-system/components";
import { Reveal, ShimmerCard, ConnectionGrid } from "../landing/components";
import { C } from "../landing/constants";

/* ------------------------------------------------------------------ */
/*  Data                                                               */
/* ------------------------------------------------------------------ */

const techDetails = [
  {
    icon: (
      <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke={C.primary} strokeWidth="1.5" strokeLinecap="round">
        <rect x="3" y="11" width="18" height="11" rx="2" />
        <path d="M7 11V7a5 5 0 0 1 10 0v4" />
      </svg>
    ),
    color: C.primary,
    title: "TLS Encryption",
    desc: "All traffic between agents and the gateway is encrypted with TLS 1.3. No plaintext data ever leaves a device.",
  },
  {
    icon: (
      <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke={C.cyan} strokeWidth="1.5" strokeLinecap="round">
        <path d="M4 4h16c1.1 0 2 .9 2 2v12c0 1.1-.9 2-2 2H4c-1.1 0-2-.9-2-2V6c0-1.1.9-2 2-2z" />
        <polyline points="22,6 12,13 2,6" />
      </svg>
    ),
    color: C.cyan,
    title: "WebSocket Tunnels",
    desc: "Persistent WebSocket connections over port 443 ensure reliable communication even through restrictive proxies.",
  },
  {
    icon: (
      <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke={C.green} strokeWidth="1.5" strokeLinecap="round">
        <polyline points="16 3 21 3 21 8" />
        <line x1="4" y1="20" x2="21" y2="3" />
        <polyline points="21 16 21 21 16 21" />
        <line x1="15" y1="15" x2="21" y2="21" />
        <line x1="4" y1="4" x2="9" y2="9" />
      </svg>
    ),
    color: C.green,
    title: "Reverse SSH",
    desc: "The agent initiates the connection outbound, then the gateway multiplexes inbound SSH sessions back through the same tunnel.",
  },
  {
    icon: (
      <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke={C.yellow} strokeWidth="1.5" strokeLinecap="round">
        <path d="M12 22s8-4 8-10V5l-8-3-8 3v7c0 6 8 10 8 10z" />
      </svg>
    ),
    color: C.yellow,
    title: "NAT Traversal",
    desc: "Outbound-only connections mean devices behind any NAT, CGNAT, or carrier-grade firewall are reachable without port forwarding.",
  },
  {
    icon: (
      <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke={C.blue} strokeWidth="1.5" strokeLinecap="round">
        <path d="M21 12a9 9 0 0 1-9 9m9-9a9 9 0 0 0-9-9m9 9H3m9 9a9 9 0 0 1-9-9m9 9c1.66 0 3-4.03 3-9s-1.34-9-3-9m0 18c-1.66 0-3-4.03-3-9s1.34-9 3-9m-9 9a9 9 0 0 1 9-9" />
      </svg>
    ),
    color: C.blue,
    title: "Agent Auto-Update",
    desc: "Agents check for updates automatically and apply them without downtime. Always running the latest secure version.",
  },
  {
    icon: (
      <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke={C.red} strokeWidth="1.5" strokeLinecap="round">
        <path d="M20.24 12.24a6 6 0 0 0-8.49-8.49L5 10.5V19h8.5z" />
        <line x1="16" y1="8" x2="2" y2="22" />
        <line x1="17.5" y1="15" x2="9" y2="15" />
      </svg>
    ),
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
  const [navSolid, setNavSolid] = useState(false);
  const [mobileMenu, setMobileMenu] = useState(false);
  const handleScroll = useCallback(() => { setNavSolid(window.scrollY > 50); }, []);
  useEffect(() => { window.addEventListener("scroll", handleScroll, { passive: true }); return () => window.removeEventListener("scroll", handleScroll); }, [handleScroll]);

  return (
    <div className="min-h-screen bg-background text-text-primary font-sans antialiased overflow-x-hidden">
      <Navbar navSolid={navSolid} mobileMenu={mobileMenu} setMobileMenu={setMobileMenu} />

      {/* ── Hero ─────────────────────────────────────────────────── */}
      <section className="relative pt-32 pb-24 overflow-hidden">
        <ConnectionGrid />
        <div className="absolute inset-0 bg-gradient-radial from-primary/8 via-transparent to-transparent pointer-events-none" />
        <div className="absolute top-1/3 left-1/4 w-96 h-96 bg-accent-cyan/6 rounded-full blur-3xl pointer-events-none" />

        <div className="max-w-7xl mx-auto px-8 relative z-10 text-center">
          <Reveal>
            <span className="inline-block px-3 py-1 text-2xs font-mono font-semibold uppercase tracking-[0.15em] bg-primary/10 text-primary border border-primary/20 rounded-full mb-6">
              How It Works
            </span>
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
              ShellHub replaces VPNs, public IPs, and firewall rules with a single secure gateway.
              Install an agent, and SSH in from anywhere in the world.
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
                href="#architecture"
                className="inline-flex items-center gap-2 px-6 py-3.5 text-[15px] font-medium text-text-secondary border border-border rounded-xl hover:text-text-primary hover:border-border-light hover:bg-white/[0.04] transition-all duration-300"
              >
                See the Architecture
              </a>
            </div>
          </Reveal>
        </div>
      </section>

      {/* ── Architecture Diagram ─────────────────────────────────── */}
      <section id="architecture" className="py-24 border-t border-border">
        <div className="max-w-7xl mx-auto px-8">
          <Reveal className="text-center mb-14">
            <p className="text-2xs font-mono font-semibold uppercase tracking-[0.15em] text-[#7B8EDB] mb-3">Architecture</p>
            <h2 className="text-[clamp(1.75rem,4vw,3rem)] font-bold tracking-[-0.03em] leading-tight mb-4">
              The full picture, from user to device
            </h2>
            <p className="text-sm text-text-secondary max-w-lg mx-auto leading-relaxed">
              Every connection flows through the ShellHub gateway where it is authenticated, encrypted, and logged before reaching the target device.
            </p>
          </Reveal>

          <Reveal>
            <ShimmerCard className="bg-card border border-border rounded-xl p-6 lg:p-10 overflow-x-auto">
              <svg viewBox="0 0 960 340" fill="none" xmlns="http://www.w3.org/2000/svg" className="w-full h-auto min-w-[720px]">
                <defs>
                  <marker id="hw-a-pri" markerWidth="8" markerHeight="6" refX="8" refY="3" orient="auto"><path d="M0,0 L8,3 L0,6" fill={C.primary}/></marker>
                  <marker id="hw-a-grn" markerWidth="8" markerHeight="6" refX="8" refY="3" orient="auto"><path d="M0,0 L8,3 L0,6" fill={C.green}/></marker>
                  <marker id="hw-a-dim" markerWidth="8" markerHeight="6" refX="8" refY="3" orient="auto"><path d="M0,0 L8,3 L0,6" fill={`${C.primary}60`}/></marker>
                </defs>

                {/* ── User ── */}
                <text x="70" y="24" fontFamily="IBM Plex Sans" fontSize="11" fill={C.textMuted} textAnchor="middle" letterSpacing=".1em">YOU</text>
                <rect x="20" y="36" width="100" height="90" rx="12" fill={C.card} stroke={C.border} strokeWidth="1.2"/>
                <circle cx="70" cy="62" r="14" stroke={C.primary} strokeWidth="1.5" fill="none"/>
                <circle cx="70" cy="58" r="4.5" fill={C.primary}/>
                <path d="M70 63 L70 67 M63 65 L77 65" stroke={C.primary} strokeWidth="1.5" strokeLinecap="round"/>
                <text x="70" y="100" fontFamily="IBM Plex Mono" fontSize="9" fill={C.textSec} textAnchor="middle">Any location</text>
                <text x="70" y="114" fontFamily="IBM Plex Mono" fontSize="8" fill={C.textMuted} textAnchor="middle">laptop / browser</text>

                {/* ── Arrow: User to Gateway ── */}
                <line x1="125" y1="80" x2="260" y2="80" stroke={C.primary} strokeWidth="1.5" strokeDasharray="6 4" markerEnd="url(#hw-a-pri)"/>
                <rect x="155" y="58" width="68" height="16" rx="4" fill={C.bg}/>
                <text x="189" y="70" fontFamily="IBM Plex Mono" fontSize="8" fill={C.primary} textAnchor="middle">SSH / HTTPS</text>

                {/* ── ShellHub Cloud ── */}
                <rect x="265" y="20" width="330" height="290" rx="16" fill={`${C.primary}08`} stroke={C.primary} strokeWidth="1.5"/>
                <text x="430" y="14" fontFamily="IBM Plex Sans" fontSize="12" fill={C.primary} textAnchor="middle" fontWeight="600" letterSpacing=".1em">SHELLHUB CLOUD</text>

                {/* Logo badge */}
                <rect x="395" y="38" width="70" height="36" rx="10" fill={C.primaryDim} stroke={C.primary} strokeWidth="1"/>
                <text x="430" y="62" fontFamily="IBM Plex Mono" fontSize="16" fill={C.primary} textAnchor="middle" fontWeight="700">SH</text>

                {/* Internal modules */}
                {[
                  { x: 285, y: 100, label: "Authentication", color: C.primary, icon: "lock" },
                  { x: 285, y: 140, label: "TLS Encryption", color: C.primary, icon: "shield" },
                  { x: 285, y: 180, label: "Session Recording", color: C.cyan, icon: "rec" },
                  { x: 285, y: 220, label: "Connection Router", color: C.green, icon: "route" },
                  { x: 445, y: 100, label: "Firewall Rules", color: C.yellow, icon: "fire" },
                  { x: 445, y: 140, label: "Audit Logging", color: C.green, icon: "log" },
                  { x: 445, y: 180, label: "Team RBAC", color: C.primary, icon: "team" },
                  { x: 445, y: 220, label: "Device Registry", color: C.blue, icon: "device" },
                ].map((m, i) => (
                  <g key={i}>
                    <rect x={m.x} y={m.y} width="130" height="30" rx="6" fill={C.card} stroke={C.border}/>
                    <circle cx={m.x + 16} cy={m.y + 15} r="5" fill={`${m.color}30`} stroke={m.color} strokeWidth=".8"/>
                    <text x={m.x + 32} y={m.y + 19} fontFamily="IBM Plex Sans" fontSize="9" fill={C.textSec}>{m.label}</text>
                  </g>
                ))}

                {/* Gateway label */}
                <text x="430" y="278" fontFamily="IBM Plex Mono" fontSize="8" fill={C.textMuted} textAnchor="middle">CLOUD OR SELF-HOSTED</text>

                {/* ── NAT Wall ── */}
                <rect x="630" y="30" width="8" height="270" rx="4" fill={C.border}/>
                <rect x="630" y="30" width="8" height="270" rx="4" fill={`${C.red}12`}/>
                <text x="634" y="22" fontFamily="IBM Plex Mono" fontSize="9" fill={C.textMuted} textAnchor="middle">NAT</text>

                {/* ── Arrow: Gateway to NAT ── */}
                <line x1="600" y1="115" x2="626" y2="115" stroke={C.green} strokeWidth="1.5" markerEnd="url(#hw-a-grn)"/>

                {/* ── Arrows: NAT to Devices ── */}
                <line x1="642" y1="85" x2="678" y2="65" stroke={`${C.primary}60`} strokeWidth="1.2" strokeDasharray="4 3" markerEnd="url(#hw-a-dim)"/>
                <line x1="642" y1="125" x2="678" y2="140" stroke={`${C.primary}60`} strokeWidth="1.2" strokeDasharray="4 3" markerEnd="url(#hw-a-dim)"/>
                <line x1="642" y1="175" x2="678" y2="215" stroke={`${C.primary}60`} strokeWidth="1.2" strokeDasharray="4 3" markerEnd="url(#hw-a-dim)"/>
                <line x1="642" y1="220" x2="678" y2="280" stroke={`${C.primary}60`} strokeWidth="1.2" strokeDasharray="4 3" markerEnd="url(#hw-a-dim)"/>

                <text x="658" y="170" fontFamily="IBM Plex Mono" fontSize="7" fill={`${C.primary}50`} textAnchor="middle" transform="rotate(-90,658,170)">NAT Traversal</text>

                {/* ── YOUR DEVICES label ── */}
                <text x="790" y="22" fontFamily="IBM Plex Sans" fontSize="11" fill={C.textMuted} textAnchor="middle" letterSpacing=".1em">YOUR DEVICES</text>

                {/* ── Devices ── */}
                {[
                  { y: 36, icon: "Pi", iconBg: C.green, label: "Raspberry Pi", sub: "armv7 / aarch64", begin: "0s" },
                  { y: 112, icon: "srv", iconBg: C.primary, label: "Linux Server", sub: "Ubuntu / Debian / RHEL", begin: ".5s" },
                  { y: 188, icon: "dk", iconBg: C.blue, label: "Docker Host", sub: "container agent", begin: "1s" },
                  { y: 264, icon: "iot", iconBg: C.yellow, label: "IoT Gateway", sub: "OpenWrt / Yocto", begin: "1.5s" },
                ].map((d, i) => (
                  <g key={i}>
                    <rect x="682" y={d.y} width="160" height="60" rx="10" fill={C.card} stroke={C.border}/>
                    {/* icon box */}
                    <rect x="696" y={d.y + 12} width="28" height="20" rx="4" fill={`${d.iconBg}15`} stroke={d.iconBg} strokeWidth=".8"/>
                    <text x="710" y={d.y + 26} fontSize="9" fill={d.iconBg} textAnchor="middle" fontFamily="IBM Plex Mono" fontWeight="600">
                      {d.icon === "Pi" ? "Pi" : d.icon === "srv" ? ">_" : d.icon === "dk" ? "dk" : "IoT"}
                    </text>
                    {/* labels */}
                    <text x="736" y={d.y + 24} fontFamily="IBM Plex Sans" fontSize="10" fill={C.text}>{d.label}</text>
                    <text x="736" y={d.y + 38} fontFamily="IBM Plex Mono" fontSize="7" fill={C.textMuted}>{d.sub}</text>
                    <text x="736" y={d.y + 50} fontFamily="IBM Plex Mono" fontSize="7" fill={C.textMuted}>agent running</text>
                    {/* animated status dot */}
                    <circle cx="834" cy={d.y + 10} r="3.5" fill={C.green}>
                      <animate attributeName="opacity" values="1;.3;1" dur="2s" repeatCount="indefinite" begin={d.begin}/>
                    </circle>
                  </g>
                ))}
              </svg>
            </ShimmerCard>
          </Reveal>
        </div>
      </section>

      {/* ── 3-Step Process (Vertical Timeline) ───────────────────── */}
      <section className="py-24 border-t border-border">
        <div className="max-w-7xl mx-auto px-8">
          <Reveal className="text-center mb-16">
            <p className="text-2xs font-mono font-semibold uppercase tracking-[0.15em] text-[#7B8EDB] mb-3">Getting Started</p>
            <h2 className="text-[clamp(1.75rem,4vw,3rem)] font-bold tracking-[-0.03em] leading-tight mb-4">
              Three steps to secure remote access
            </h2>
            <p className="text-sm text-text-secondary max-w-lg mx-auto leading-relaxed">
              From first install to first connection in under five minutes.
            </p>
          </Reveal>

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
                    <span className="w-10 h-10 rounded-xl bg-primary/10 border border-primary/20 flex items-center justify-center text-sm font-bold font-mono text-primary">01</span>
                    <span className="text-2xs font-mono font-semibold uppercase tracking-[0.1em] text-primary">Install</span>
                  </div>
                  <h3 className="text-lg font-bold tracking-[-0.02em] mb-3">Install the ShellHub Agent</h3>
                  <p className="text-sm text-text-secondary leading-relaxed mb-4">
                    Deploy a lightweight agent on each device you want to manage. One command works on any Linux system, Raspberry Pi, container, or VM.
                  </p>
                  <ul className="space-y-2 md:ml-auto md:mr-0">
                    {["Single-line install script", "Under 10 MB footprint", "Runs as a system service", "Auto-starts on boot"].map((d) => (
                      <li key={d} className="flex items-center gap-2 text-xs text-text-secondary md:justify-end">
                        <svg className="w-3.5 h-3.5 text-primary shrink-0 md:order-last" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2.5}><path strokeLinecap="round" strokeLinejoin="round" d="m4.5 12.75 6 6 9-13.5" /></svg>
                        {d}
                      </li>
                    ))}
                  </ul>
                </div>

                <div className="md:pl-12">
                  <ShimmerCard className="bg-card border border-border rounded-xl overflow-hidden">
                    <div className="p-5">
                      <div className="flex items-center gap-2 mb-4">
                        <div className="w-3 h-3 rounded-full bg-accent-red/60" />
                        <div className="w-3 h-3 rounded-full bg-accent-yellow/60" />
                        <div className="w-3 h-3 rounded-full bg-accent-green/60" />
                        <span className="ml-2 text-2xs text-text-muted font-mono">Terminal</span>
                      </div>
                      <div className="bg-surface rounded-lg p-4 font-mono text-xs leading-relaxed space-y-2 overflow-x-auto">
                        <p><span className="text-accent-green">$</span> <span className="text-text-secondary">curl -sSf https://cloud.shellhub.io/install.sh | sh</span></p>
                        <p className="text-text-muted"># Downloading ShellHub agent v0.17.2...</p>
                        <p className="text-text-muted"># Installing to /usr/local/bin/shellhub-agent</p>
                        <p className="text-text-muted"># Registering systemd service...</p>
                        <p className="text-accent-green"># Agent installed and running.</p>
                        <p className="text-accent-green"># Device ID: a1b2c3d4-e5f6-7890</p>
                      </div>
                    </div>
                  </ShimmerCard>
                </div>
              </div>
            </Reveal>

            {/* Step 2: Agent Connects */}
            <Reveal>
              <div className="relative grid md:grid-cols-2 gap-8 mb-16">
                <div className="absolute left-8 lg:left-1/2 top-8 w-3 h-3 -ml-1.5 rounded-full bg-accent-cyan border-2 border-background z-10 hidden md:block" />

                <div className="md:pr-12 md:text-right md:order-last md:text-left md:pl-12 md:pr-0">
                  <div className="flex items-center gap-3 mb-4">
                    <span className="w-10 h-10 rounded-xl bg-accent-cyan/10 border border-accent-cyan/20 flex items-center justify-center text-sm font-bold font-mono text-accent-cyan">02</span>
                    <span className="text-2xs font-mono font-semibold uppercase tracking-[0.1em] text-accent-cyan">Connect</span>
                  </div>
                  <h3 className="text-lg font-bold tracking-[-0.02em] mb-3">Agent Connects to ShellHub</h3>
                  <p className="text-sm text-text-secondary leading-relaxed mb-4">
                    The agent initiates an outbound connection to the ShellHub gateway. No inbound ports, no public IPs, no firewall changes required on the device side.
                  </p>
                  <ul className="space-y-2">
                    {["Outbound-only connection (port 443)", "Works behind NAT, firewalls, CGNAT", "Automatic TLS encryption", "Persistent WebSocket tunnel"].map((d) => (
                      <li key={d} className="flex items-center gap-2 text-xs text-text-secondary">
                        <svg className="w-3.5 h-3.5 text-accent-cyan shrink-0" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2.5}><path strokeLinecap="round" strokeLinejoin="round" d="m4.5 12.75 6 6 9-13.5" /></svg>
                        {d}
                      </li>
                    ))}
                  </ul>
                </div>

                <div className="md:pr-12">
                  <ShimmerCard className="bg-card border border-border rounded-xl overflow-hidden">
                    <div className="p-5">
                      <svg viewBox="0 0 400 200" fill="none" xmlns="http://www.w3.org/2000/svg" className="w-full h-auto">
                        <defs>
                          <marker id="hw-s2-a" markerWidth="7" markerHeight="5" refX="7" refY="2.5" orient="auto"><path d="M0,0 L7,2.5 L0,5" fill={C.cyan}/></marker>
                        </defs>

                        {/* Device */}
                        <rect x="20" y="60" width="90" height="80" rx="10" fill={C.card} stroke={C.border}/>
                        <rect x="35" y="74" width="24" height="16" rx="3" fill={`${C.green}15`} stroke={C.green} strokeWidth=".8"/>
                        <text x="47" y="86" fontSize="8" fill={C.green} textAnchor="middle" fontFamily="IBM Plex Mono" fontWeight="600">Pi</text>
                        <text x="65" y="110" fontFamily="IBM Plex Sans" fontSize="9" fill={C.textSec} textAnchor="middle">Device</text>
                        <text x="65" y="124" fontFamily="IBM Plex Mono" fontSize="7" fill={C.textMuted} textAnchor="middle">agent</text>
                        <circle cx="100" cy="68" r="3" fill={C.green}>
                          <animate attributeName="opacity" values="1;.3;1" dur="2s" repeatCount="indefinite"/>
                        </circle>

                        {/* NAT Wall */}
                        <rect x="145" y="40" width="6" height="120" rx="3" fill={C.border}/>
                        <rect x="145" y="40" width="6" height="120" rx="3" fill={`${C.red}15`}/>
                        <text x="148" y="34" fontFamily="IBM Plex Mono" fontSize="8" fill={C.textMuted} textAnchor="middle">NAT</text>

                        {/* Outbound arrow: Device through NAT to Cloud */}
                        <line x1="115" y1="100" x2="142" y2="100" stroke={C.cyan} strokeWidth="1.5" markerEnd="url(#hw-s2-a)"/>
                        <line x1="154" y1="100" x2="225" y2="100" stroke={C.cyan} strokeWidth="1.5" strokeDasharray="6 4" markerEnd="url(#hw-s2-a)"/>
                        <text x="190" y="92" fontFamily="IBM Plex Mono" fontSize="7" fill={C.cyan} textAnchor="middle">outbound :443</text>

                        {/* Cloud */}
                        <rect x="230" y="50" width="150" height="100" rx="14" fill={`${C.primary}08`} stroke={C.primary} strokeWidth="1.2"/>
                        <rect x="278" y="68" width="54" height="28" rx="8" fill={C.primaryDim} stroke={C.primary} strokeWidth=".8"/>
                        <text x="305" y="87" fontFamily="IBM Plex Mono" fontSize="12" fill={C.primary} textAnchor="middle" fontWeight="700">SH</text>
                        <text x="305" y="118" fontFamily="IBM Plex Sans" fontSize="9" fill={C.textSec} textAnchor="middle">ShellHub Gateway</text>
                        <text x="305" y="132" fontFamily="IBM Plex Mono" fontSize="7" fill={C.textMuted} textAnchor="middle">Auth + Encryption</text>

                        {/* Labels */}
                        <text x="65" y="170" fontFamily="IBM Plex Mono" fontSize="7" fill={C.textMuted} textAnchor="middle">PRIVATE NETWORK</text>
                        <text x="305" y="170" fontFamily="IBM Plex Mono" fontSize="7" fill={C.textMuted} textAnchor="middle">PUBLIC CLOUD</text>
                      </svg>
                    </div>
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
                    <span className="w-10 h-10 rounded-xl bg-accent-green/10 border border-accent-green/20 flex items-center justify-center text-sm font-bold font-mono text-accent-green">03</span>
                    <span className="text-2xs font-mono font-semibold uppercase tracking-[0.1em] text-accent-green">Access</span>
                  </div>
                  <h3 className="text-lg font-bold tracking-[-0.02em] mb-3">SSH from Anywhere</h3>
                  <p className="text-sm text-text-secondary leading-relaxed mb-4">
                    Use your standard SSH client to connect through ShellHub. The gateway authenticates the user, applies firewall rules, and routes the connection to the right device.
                  </p>
                  <ul className="space-y-2 md:ml-auto md:mr-0">
                    {["Standard SSH client, no plugins", "MFA and public key auth", "Firewall rules per device/user", "Session recording and audit trail"].map((d) => (
                      <li key={d} className="flex items-center gap-2 text-xs text-text-secondary md:justify-end">
                        <svg className="w-3.5 h-3.5 text-accent-green shrink-0 md:order-last" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2.5}><path strokeLinecap="round" strokeLinejoin="round" d="m4.5 12.75 6 6 9-13.5" /></svg>
                        {d}
                      </li>
                    ))}
                  </ul>
                </div>

                <div className="md:pl-12">
                  <ShimmerCard className="bg-card border border-border rounded-xl overflow-hidden">
                    <div className="p-5">
                      <div className="flex items-center gap-2 mb-4">
                        <div className="w-3 h-3 rounded-full bg-accent-red/60" />
                        <div className="w-3 h-3 rounded-full bg-accent-yellow/60" />
                        <div className="w-3 h-3 rounded-full bg-accent-green/60" />
                        <span className="ml-2 text-2xs text-text-muted font-mono">Terminal</span>
                      </div>
                      <div className="bg-surface rounded-lg p-4 font-mono text-xs leading-relaxed space-y-2 overflow-x-auto">
                        <p><span className="text-accent-green">$</span> <span className="text-text-secondary">ssh pi@a1b2c3d4-e5f6-7890.mycompany.shellhub.io</span></p>
                        <p className="text-text-muted"># Connecting through ShellHub gateway...</p>
                        <p className="text-text-muted"># Authenticating with public key...</p>
                        <p className="text-text-muted"># Session recording enabled.</p>
                        <p>&nbsp;</p>
                        <p><span className="text-accent-green">pi@raspberrypi</span>:<span className="text-accent-blue">~</span><span className="text-text-secondary">$ uname -a</span></p>
                        <p className="text-text-secondary">Linux raspberrypi 6.1.0-rpi7 armv7l GNU/Linux</p>
                        <p><span className="text-accent-green">pi@raspberrypi</span>:<span className="text-accent-blue">~</span><span className="text-text-muted animate-pulse">_</span></p>
                      </div>
                    </div>
                  </ShimmerCard>
                </div>
              </div>
            </Reveal>
          </div>
        </div>
      </section>

      {/* ── Why Not VPN? (Comparison) ────────────────────────────── */}
      <section className="py-24 border-t border-border">
        <div className="max-w-7xl mx-auto px-8">
          <Reveal className="text-center mb-14">
            <p className="text-2xs font-mono font-semibold uppercase tracking-[0.15em] text-[#7B8EDB] mb-3">Comparison</p>
            <h2 className="text-[clamp(1.75rem,4vw,3rem)] font-bold tracking-[-0.03em] leading-tight mb-4">
              Why ShellHub over a VPN?
            </h2>
            <p className="text-sm text-text-secondary max-w-lg mx-auto leading-relaxed">
              VPNs were designed to link networks, not manage individual devices. ShellHub gives you direct, secure access without the overhead.
            </p>
          </Reveal>

          <div className="grid md:grid-cols-2 gap-6">
            {/* ShellHub Side (highlighted) */}
            <Reveal delay={0}>
              <ShimmerCard className="h-full">
                <div className="relative bg-card border border-primary/30 rounded-xl p-8 flex flex-col h-full hover:border-primary/50 transition-all duration-300 shadow-[0_0_40px_rgba(102,122,204,0.1)] overflow-hidden">
                  <div className="absolute inset-0 bg-gradient-to-br from-primary/[0.06] via-transparent to-transparent pointer-events-none" />
                  <div className="relative">
                    <div className="flex items-center gap-3 mb-6">
                      <div className="w-10 h-10 rounded-lg bg-primary/10 border border-primary/20 flex items-center justify-center">
                        <svg className="w-5 h-5 text-primary" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={1.5}>
                          <path strokeLinecap="round" strokeLinejoin="round" d="M9 12.75 11.25 15 15 9.75m-3-7.036A11.959 11.959 0 0 1 3.598 6 11.99 11.99 0 0 0 3 9.749c0 5.592 3.824 10.29 9 11.623 5.176-1.332 9-6.03 9-11.622 0-1.31-.21-2.571-.598-3.751h-.152c-3.196 0-6.1-1.248-8.25-3.285z" />
                        </svg>
                      </div>
                      <div>
                        <h3 className="text-sm font-bold">ShellHub</h3>
                        <p className="text-2xs text-primary">Purpose-built for device access</p>
                      </div>
                      <span className="ml-auto px-2 py-0.5 text-2xs font-mono font-semibold uppercase tracking-[0.1em] bg-accent-green/10 text-accent-green border border-accent-green/20 rounded-full">
                        Recommended
                      </span>
                    </div>

                    <ul className="space-y-3">
                      {shellhubAdvantages.map((item) => (
                        <li key={item.label} className="flex items-center gap-2.5 text-sm text-text-secondary">
                          <svg className="w-4 h-4 text-accent-green shrink-0" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
                            <path strokeLinecap="round" strokeLinejoin="round" d="m4.5 12.75 6 6 9-13.5" />
                          </svg>
                          {item.label}
                        </li>
                      ))}
                    </ul>
                  </div>
                </div>
              </ShimmerCard>
            </Reveal>

            {/* VPN Side */}
            <Reveal delay={0.1}>
              <ShimmerCard className="h-full">
                <div className="bg-card border border-border rounded-xl p-8 flex flex-col h-full hover:border-border-light transition-colors duration-300">
                  <div className="flex items-center gap-3 mb-6">
                    <div className="w-10 h-10 rounded-lg bg-white/[0.04] border border-border flex items-center justify-center">
                      <svg className="w-5 h-5 text-text-secondary" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={1.5}>
                        <path strokeLinecap="round" strokeLinejoin="round" d="M12 9v3.75m0-10.036A11.959 11.959 0 0 1 3.598 6 11.99 11.99 0 0 0 3 9.75c0 5.592 3.824 10.29 9 11.622 5.176-1.332 9-6.03 9-11.622 0-1.31-.21-2.57-.598-3.75h-.152c-3.196 0-6.1-1.25-8.25-3.286zM12 15.75h.007v.008H12v-.008z" />
                      </svg>
                    </div>
                    <div>
                      <h3 className="text-sm font-bold">Traditional VPN</h3>
                      <p className="text-2xs text-text-muted">Network-level tunneling</p>
                    </div>
                  </div>

                  <ul className="space-y-3">
                    {vpnLimitations.map((item) => (
                      <li key={item.label} className="flex items-center gap-2.5 text-sm text-text-secondary">
                        <svg className="w-4 h-4 text-accent-red/70 shrink-0" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
                          <path strokeLinecap="round" strokeLinejoin="round" d="M6 18 18 6M6 6l12 12" />
                        </svg>
                        {item.label}
                      </li>
                    ))}
                  </ul>
                </div>
              </ShimmerCard>
            </Reveal>
          </div>
        </div>
      </section>

      {/* ── Technical Details ─────────────────────────────────────── */}
      <section className="py-24 border-t border-border">
        <div className="max-w-7xl mx-auto px-8">
          <Reveal className="text-center mb-14">
            <p className="text-2xs font-mono font-semibold uppercase tracking-[0.15em] text-[#7B8EDB] mb-3">Under the Hood</p>
            <h2 className="text-[clamp(1.75rem,4vw,3rem)] font-bold tracking-[-0.03em] leading-tight mb-4">
              Built on solid foundations
            </h2>
            <p className="text-sm text-text-secondary max-w-lg mx-auto leading-relaxed">
              Every component of the stack is designed for security, reliability, and minimal overhead.
            </p>
          </Reveal>

          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
            {techDetails.map((f, i) => (
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
        </div>
      </section>

      {/* ── CTA ──────────────────────────────────────────────────── */}
      <section className="py-24 border-t border-border">
        <div className="max-w-7xl mx-auto px-8">
          <Reveal>
            <div className="relative bg-card border border-border rounded-2xl p-12 text-center overflow-hidden">
              <ConnectionGrid />
              <div className="absolute inset-0 bg-gradient-to-br from-primary/[0.06] via-transparent to-accent-cyan/[0.04] pointer-events-none" />

              <div className="relative z-10">
                <p className="text-2xs font-mono font-semibold uppercase tracking-[0.15em] text-[#7B8EDB] mb-3">
                  Ready to try it?
                </p>
                <h2 className="text-[clamp(1.5rem,3vw,2.25rem)] font-bold tracking-[-0.03em] leading-tight mb-4">
                  See it in action
                </h2>
                <p className="text-sm text-text-secondary max-w-md mx-auto leading-relaxed mb-8">
                  Deploy ShellHub in under five minutes and connect to your first device. No credit card required.
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
                    href="/v2/pricing"
                    className="inline-flex items-center gap-2 px-6 py-3.5 text-[15px] font-medium text-text-secondary border border-border rounded-xl hover:text-text-primary hover:border-border-light hover:bg-white/[0.04] transition-all duration-300"
                  >
                    View Pricing
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
