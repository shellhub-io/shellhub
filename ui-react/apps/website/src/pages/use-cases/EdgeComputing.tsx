import { useEffect, useState, useCallback } from "react";
import { Navbar } from "../landing/Navbar";
import { Footer } from "@shellhub/design-system/components";
import { Reveal, ShimmerCard, ConnectionGrid } from "../landing/components";
import { C } from "../landing/constants";

/* ═══════ Pain-point data ═══════ */
const painPoints = [
  {
    color: C.primary,
    title: "Distributed locations",
    desc: "Edge servers scattered across retail stores, warehouses, cell towers, and data centers — each with unique network topology.",
    icon: (
      <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke={C.primary} strokeWidth="1.5" strokeLinecap="round" strokeLinejoin="round">
        <circle cx="12" cy="12" r="10" />
        <path d="M2 12h20" />
        <path d="M12 2a15.3 15.3 0 0 1 4 10 15.3 15.3 0 0 1-4 10 15.3 15.3 0 0 1-4-10 15.3 15.3 0 0 1 4-10z" />
      </svg>
    ),
  },
  {
    color: C.yellow,
    title: "Unreliable connectivity",
    desc: "Intermittent or low-bandwidth links make traditional VPN tunnels unstable and impossible to maintain reliably.",
    icon: (
      <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke={C.yellow} strokeWidth="1.5" strokeLinecap="round" strokeLinejoin="round">
        <path d="M1 1l22 22" />
        <path d="M16.72 11.06A10.94 10.94 0 0 1 19 12.55" />
        <path d="M5 12.55a10.94 10.94 0 0 1 5.17-2.39" />
        <path d="M10.71 5.05A16 16 0 0 1 22.56 9" />
        <path d="M1.42 9a15.91 15.91 0 0 1 4.7-2.88" />
        <path d="M8.53 16.11a6 6 0 0 1 6.95 0" />
        <line x1="12" y1="20" x2="12.01" y2="20" />
      </svg>
    ),
  },
  {
    color: C.red,
    title: "On-site visits are expensive",
    desc: "Dispatching a technician to a remote cell tower or warehouse for a configuration change burns time and budget.",
    icon: (
      <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke={C.red} strokeWidth="1.5" strokeLinecap="round" strokeLinejoin="round">
        <path d="M12 22s-8-4.5-8-11.8A8 8 0 0 1 12 2a8 8 0 0 1 8 8.2c0 7.3-8 11.8-8 11.8z" />
        <circle cx="12" cy="10" r="3" />
      </svg>
    ),
  },
  {
    color: C.cyan,
    title: "Security at scale",
    desc: "Managing SSH keys, firewall rules, and access policies across hundreds of edge locations is operationally complex.",
    icon: (
      <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke={C.cyan} strokeWidth="1.5" strokeLinecap="round" strokeLinejoin="round">
        <path d="M12 22s8-4 8-10V5l-8-3-8 3v7c0 6 8 10 8 10z" />
      </svg>
    ),
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
        <div style={{ color: C.green }}>● pos-service.service - POS Application</div>
        <div style={{ color: C.green }}>&nbsp;&nbsp;Active: active (running) since Mon</div>
        <div style={{ color: C.textMuted }}>$ pos-cli --update-config --store=042</div>
        <div style={{ color: C.primary }}>Config updated. Restarting service...</div>
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
        <div style={{ color: C.textMuted }}>$ scp firmware-v3.2.1.bin admin@tower-017:/opt/fw/</div>
        <div style={{ color: C.yellow }}>firmware-v3.2.1.bin&nbsp;&nbsp;78%&nbsp;&nbsp;14MB&nbsp;&nbsp;2.1MB/s&nbsp;&nbsp;00:03</div>
        <div style={{ color: C.green }}>firmware-v3.2.1.bin&nbsp;&nbsp;100%&nbsp;&nbsp;18MB&nbsp;&nbsp;2.3MB/s&nbsp;&nbsp;00:00</div>
        <div style={{ color: C.textMuted }}>$ fw-upgrade --apply --verify</div>
        <div style={{ color: C.green }}>Firmware v3.2.1 applied. Radio restarting...</div>
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
          <span className="text-2xs font-mono" style={{ color: C.textSec }}>inv-server-chicago</span>
          <span className="px-2 py-0.5 text-2xs font-mono bg-accent-green/10 text-accent-green border border-accent-green/20 rounded-full">Healthy</span>
        </div>
        <div className="grid grid-cols-3 gap-2">
          {[
            { label: "CPU", value: "23%", color: C.green },
            { label: "Memory", value: "4.2 GB", color: C.primary },
            { label: "Disk", value: "67%", color: C.yellow },
          ].map((m) => (
            <div key={m.label} className="bg-background rounded-md p-2 text-center">
              <p className="text-2xs font-mono" style={{ color: C.textMuted }}>{m.label}</p>
              <p className="text-xs font-semibold" style={{ color: m.color }}>{m.value}</p>
            </div>
          ))}
        </div>
        <div className="mt-3 pt-2 border-t border-border">
          <div className="flex items-center justify-between text-2xs font-mono">
            <span style={{ color: C.textMuted }}>Last scan: 142,847 items indexed</span>
            <span style={{ color: C.green }}>Synced</span>
          </div>
        </div>
      </div>
    ),
  },
];

/* ═══════ Component ═══════ */
export default function EdgeComputing() {
  const [navSolid, setNavSolid] = useState(false);
  const [mobileMenu, setMobileMenu] = useState(false);
  const handleScroll = useCallback(() => { setNavSolid(window.scrollY > 50); }, []);
  useEffect(() => { window.addEventListener("scroll", handleScroll, { passive: true }); return () => window.removeEventListener("scroll", handleScroll); }, [handleScroll]);

  return (
    <div className="min-h-screen bg-background text-text-primary font-sans antialiased overflow-x-hidden">
      <Navbar navSolid={navSolid} mobileMenu={mobileMenu} setMobileMenu={setMobileMenu} />

      {/* ───── Hero ───── */}
      <section className="relative pt-32 pb-24 overflow-hidden">
        <ConnectionGrid />
        <div className="absolute inset-0 bg-gradient-radial from-accent-blue/8 via-transparent to-transparent pointer-events-none" />
        <div className="absolute top-1/3 right-1/4 w-96 h-96 bg-accent-blue/6 rounded-full blur-3xl pointer-events-none" />

        <div className="max-w-7xl mx-auto px-8 relative z-10 text-center">
          <Reveal>
            <span className="inline-block px-3 py-1 text-2xs font-mono font-semibold uppercase tracking-[0.15em] bg-accent-blue/10 text-accent-blue border border-accent-blue/20 rounded-full mb-6">
              Use Case
            </span>
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
              Securely manage edge servers across retail stores, warehouses, cell towers, and remote sites — without VPNs, static IPs, or on-site visits.
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
                Compare Plans
              </a>
            </div>
          </Reveal>
        </div>
      </section>

      {/* ───── Edge Network Map ───── */}
      <section className="py-24 border-t border-border">
        <div className="max-w-7xl mx-auto px-8">
          <Reveal className="text-center mb-14">
            <p className="text-2xs font-mono font-semibold uppercase tracking-[0.15em] text-[#7B8EDB] mb-3">
              Network Topology
            </p>
            <h2 className="text-[clamp(1.75rem,4vw,3rem)] font-bold tracking-[-0.03em] leading-tight mb-4">
              One platform, every edge location
            </h2>
            <p className="text-sm text-text-secondary max-w-lg mx-auto leading-relaxed">
              ShellHub connects to all your distributed edge infrastructure through a single control plane — no matter where your devices are.
            </p>
          </Reveal>

          <Reveal delay={0.1}>
            <ShimmerCard className="bg-card border border-border rounded-2xl overflow-hidden">
              <div className="relative p-8 md:p-12">
                {/* Background grid pattern */}
                <div className="absolute inset-0 opacity-[0.03]" style={{
                  backgroundImage: `radial-gradient(${C.text} 1px, transparent 1px)`,
                  backgroundSize: "24px 24px",
                }} />

                <svg viewBox="0 0 100 100" className="w-full max-w-3xl mx-auto relative" style={{ minHeight: 320 }}>
                  {/* Connection lines from center to each location */}
                  {edgeLocations.map((loc, i) => (
                    <line
                      key={`line-${i}`}
                      x1="50" y1="48"
                      x2={loc.x} y2={loc.y}
                      stroke={loc.color}
                      strokeWidth="0.3"
                      strokeDasharray="1.5 1"
                      opacity="0.5"
                    />
                  ))}

                  {/* Central ShellHub Cloud node */}
                  <g>
                    <circle cx="50" cy="48" r="6" fill={C.card} stroke={C.primary} strokeWidth="0.5" />
                    <circle cx="50" cy="48" r="8" fill="none" stroke={C.primary} strokeWidth="0.15" opacity="0.4" />
                    <circle cx="50" cy="48" r="10.5" fill="none" stroke={C.primary} strokeWidth="0.1" opacity="0.2" />
                    {/* Cloud icon */}
                    <path
                      d="M46.5 49.5a2.5 2.5 0 0 1 2.5-3 3 3 0 0 1 5.5-1 2 2 0 0 1 .5 4h-7.5z"
                      fill="none"
                      stroke={C.primary}
                      strokeWidth="0.4"
                      strokeLinecap="round"
                      strokeLinejoin="round"
                    />
                    <text x="50" y="56" textAnchor="middle" fill={C.primary} fontSize="2.2" fontWeight="600" fontFamily="monospace">ShellHub Cloud</text>
                  </g>

                  {/* Edge location nodes */}
                  {edgeLocations.map((loc, i) => (
                    <g key={`node-${i}`}>
                      {/* Outer glow */}
                      <circle cx={loc.x} cy={loc.y} r="4" fill={`${loc.color}08`} stroke={`${loc.color}30`} strokeWidth="0.2" />
                      {/* Server icon body */}
                      <rect x={loc.x - 2} y={loc.y - 2} width="4" height="4" rx="0.6" fill={C.surface} stroke={loc.color} strokeWidth="0.3" />
                      {/* Server lines */}
                      <line x1={loc.x - 1} y1={loc.y - 0.5} x2={loc.x + 1} y2={loc.y - 0.5} stroke={loc.color} strokeWidth="0.2" opacity="0.5" />
                      <line x1={loc.x - 1} y1={loc.y + 0.5} x2={loc.x + 1} y2={loc.y + 0.5} stroke={loc.color} strokeWidth="0.2" opacity="0.5" />
                      {/* Green status dot */}
                      <circle cx={loc.x + 1.2} cy={loc.y - 1.2} r="0.5" fill={C.green}>
                        <animate attributeName="opacity" values="1;0.4;1" dur="2s" begin={`${i * 0.4}s`} repeatCount="indefinite" />
                      </circle>
                      {/* Label */}
                      <text x={loc.x} y={loc.y + 5.5} textAnchor="middle" fill={C.text} fontSize="1.8" fontWeight="600" fontFamily="sans-serif">{loc.label}</text>
                      <text x={loc.x} y={loc.y + 7.5} textAnchor="middle" fill={C.textMuted} fontSize="1.5" fontFamily="monospace">{loc.city}</text>
                    </g>
                  ))}
                </svg>

                {/* Legend */}
                <div className="flex flex-wrap items-center justify-center gap-4 mt-6">
                  {edgeLocations.map((loc) => (
                    <div key={loc.city} className="flex items-center gap-2">
                      <div className="w-2 h-2 rounded-full" style={{ background: loc.color }} />
                      <span className="text-2xs font-mono text-text-muted">{loc.label} - {loc.city}</span>
                    </div>
                  ))}
                </div>
              </div>
            </ShimmerCard>
          </Reveal>
        </div>
      </section>

      {/* ───── Pain Points ───── */}
      <section className="py-24 border-t border-border">
        <div className="max-w-7xl mx-auto px-8">
          <Reveal className="text-center mb-14">
            <p className="text-2xs font-mono font-semibold uppercase tracking-[0.15em] text-[#7B8EDB] mb-3">
              Challenges
            </p>
            <h2 className="text-[clamp(1.75rem,4vw,3rem)] font-bold tracking-[-0.03em] leading-tight mb-4">
              Managing edge infrastructure is hard
            </h2>
            <p className="text-sm text-text-secondary max-w-lg mx-auto leading-relaxed">
              Traditional remote access tools weren't built for the realities of distributed edge computing.
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
                      <h4 className="text-sm font-semibold mb-2">{p.title}</h4>
                      <p className="text-xs text-text-secondary leading-relaxed">{p.desc}</p>
                    </div>
                  </div>
                </ShimmerCard>
              </Reveal>
            ))}
          </div>
        </div>
      </section>

      {/* ───── Key Features ───── */}
      <section className="py-24 border-t border-border">
        <div className="max-w-7xl mx-auto px-8">
          <Reveal className="text-center mb-14">
            <p className="text-2xs font-mono font-semibold uppercase tracking-[0.15em] text-[#7B8EDB] mb-3">
              Features
            </p>
            <h2 className="text-[clamp(1.75rem,4vw,3rem)] font-bold tracking-[-0.03em] leading-tight mb-4">
              Built for the edge
            </h2>
            <p className="text-sm text-text-secondary max-w-lg mx-auto leading-relaxed">
              ShellHub eliminates the complexity of edge access with NAT traversal, browser-based terminals, and fleet-wide management.
            </p>
          </Reveal>

          {/* Big highlighted card: Instant Remote Access */}
          <Reveal>
            <ShimmerCard className="mb-4">
              <div className="relative bg-card border border-primary/30 rounded-xl p-8 overflow-hidden hover:border-primary/50 transition-all duration-300 shadow-[0_0_40px_rgba(102,122,204,0.1)]">
                <div className="absolute inset-0 bg-gradient-to-br from-primary/[0.06] via-transparent to-transparent pointer-events-none" />
                <div className="relative grid lg:grid-cols-2 gap-8 items-center">
                  <div>
                    <div className="flex items-center gap-2 mb-4">
                      <div className="w-10 h-10 rounded-lg bg-primary/10 border border-primary/20 flex items-center justify-center">
                        <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke={C.primary} strokeWidth="1.5" strokeLinecap="round" strokeLinejoin="round">
                          <polyline points="4 17 10 11 4 5" />
                          <line x1="12" y1="19" x2="20" y2="19" />
                        </svg>
                      </div>
                      <span className="px-2 py-0.5 text-2xs font-mono font-semibold uppercase tracking-[0.1em] bg-accent-green/10 text-accent-green border border-accent-green/20 rounded-full">
                        Core Feature
                      </span>
                    </div>
                    <h3 className="text-lg font-bold mb-2">Instant Remote Access</h3>
                    <p className="text-sm text-text-secondary leading-relaxed">
                      Connect to any edge server in seconds — behind NAT, CGNAT, or restrictive firewalls. No VPN configuration, no port forwarding, no static IPs required. The agent handles everything.
                    </p>
                  </div>

                  {/* Terminal mockup */}
                  <div className="bg-surface rounded-xl border border-border overflow-hidden">
                    <div className="flex items-center gap-2 px-4 py-3 border-b border-border">
                      <div className="w-3 h-3 rounded-full bg-accent-red/60" />
                      <div className="w-3 h-3 rounded-full bg-accent-yellow/60" />
                      <div className="w-3 h-3 rounded-full bg-accent-green/60" />
                      <span className="ml-2 text-2xs text-text-muted font-mono">Terminal</span>
                    </div>
                    <div className="p-4 font-mono text-2xs leading-relaxed space-y-1">
                      <div><span style={{ color: C.green }}>user@ops-laptop</span><span style={{ color: C.textMuted }}>:</span><span style={{ color: C.blue }}>~</span><span style={{ color: C.textMuted }}>$</span> <span style={{ color: C.text }}>ssh admin@edge-nyc-01.production</span></div>
                      <div style={{ color: C.textMuted }}>Connecting via ShellHub tunnel...</div>
                      <div style={{ color: C.green }}>Connection established (NAT traversal)</div>
                      <div>&nbsp;</div>
                      <div><span style={{ color: C.green }}>admin@edge-nyc-01</span><span style={{ color: C.textMuted }}>:</span><span style={{ color: C.blue }}>~</span><span style={{ color: C.textMuted }}>$</span> <span style={{ color: C.text }}>uname -a</span></div>
                      <div style={{ color: C.textSec }}>Linux edge-nyc-01 5.15.0 #1 SMP x86_64</div>
                      <div><span style={{ color: C.green }}>admin@edge-nyc-01</span><span style={{ color: C.textMuted }}>:</span><span style={{ color: C.blue }}>~</span><span style={{ color: C.textMuted }}>$</span> <span style={{ color: C.text }}>systemctl status edge-service</span></div>
                      <div><span style={{ color: C.green }}>●</span> <span style={{ color: C.textSec }}>edge-service.service - Edge Compute Service</span></div>
                      <div style={{ color: C.green }}>&nbsp;&nbsp;Active: active (running) since Mon 2026-02-14</div>
                      <div><span style={{ color: C.green }}>admin@edge-nyc-01</span><span style={{ color: C.textMuted }}>:</span><span style={{ color: C.blue }}>~</span><span style={{ color: C.textMuted }}>$</span> <span className="animate-pulse" style={{ color: C.primary }}>_</span></div>
                    </div>
                  </div>
                </div>
              </div>
            </ShimmerCard>
          </Reveal>

          {/* 2-column: Web Terminal + SCP/SFTP */}
          <div className="grid md:grid-cols-2 gap-4 mb-4">
            <Reveal delay={0.05}>
              <ShimmerCard className="h-full">
                <div className="bg-card border border-border rounded-xl p-6 h-full hover:border-border-light transition-all duration-300">
                  <div className="w-10 h-10 rounded-lg flex items-center justify-center mb-4 border" style={{ background: `${C.green}15`, borderColor: `${C.green}25` }}>
                    <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke={C.green} strokeWidth="1.5" strokeLinecap="round" strokeLinejoin="round">
                      <rect x="2" y="3" width="20" height="14" rx="2" />
                      <line x1="8" y1="21" x2="16" y2="21" />
                      <line x1="12" y1="17" x2="12" y2="21" />
                    </svg>
                  </div>
                  <h4 className="text-sm font-semibold mb-2">Web Terminal</h4>
                  <p className="text-xs text-text-secondary leading-relaxed mb-4">
                    Access edge servers from any browser — no SSH client needed. Perfect for field engineers using tablets or shared workstations.
                  </p>

                  {/* Browser mockup */}
                  <div className="bg-surface rounded-lg border border-border overflow-hidden">
                    <div className="flex items-center gap-2 px-3 py-2 border-b border-border">
                      <div className="w-2 h-2 rounded-full bg-accent-red/50" />
                      <div className="w-2 h-2 rounded-full bg-accent-yellow/50" />
                      <div className="w-2 h-2 rounded-full bg-accent-green/50" />
                      <div className="flex-1 mx-2 px-2 py-0.5 bg-background rounded text-2xs font-mono text-text-muted truncate">
                        shellhub.io/terminal/edge-chi-03
                      </div>
                    </div>
                    <div className="p-3 font-mono text-2xs space-y-0.5">
                      <div><span style={{ color: C.green }}>admin@edge-chi-03</span><span style={{ color: C.textMuted }}>:</span><span style={{ color: C.blue }}>~</span><span style={{ color: C.textMuted }}>$</span> <span style={{ color: C.text }}>df -h /data</span></div>
                      <div style={{ color: C.textSec }}>Filesystem&nbsp;&nbsp;Size&nbsp;&nbsp;Used&nbsp;&nbsp;Avail&nbsp;&nbsp;Use%</div>
                      <div style={{ color: C.textSec }}>/dev/sda1&nbsp;&nbsp;&nbsp;500G&nbsp;&nbsp;312G&nbsp;&nbsp;188G&nbsp;&nbsp;&nbsp;62%</div>
                      <div><span style={{ color: C.green }}>admin@edge-chi-03</span><span style={{ color: C.textMuted }}>:</span><span style={{ color: C.blue }}>~</span><span style={{ color: C.textMuted }}>$</span> <span className="animate-pulse" style={{ color: C.primary }}>_</span></div>
                    </div>
                  </div>
                </div>
              </ShimmerCard>
            </Reveal>

            <Reveal delay={0.1}>
              <ShimmerCard className="h-full">
                <div className="bg-card border border-border rounded-xl p-6 h-full hover:border-border-light transition-all duration-300">
                  <div className="w-10 h-10 rounded-lg flex items-center justify-center mb-4 border" style={{ background: `${C.cyan}15`, borderColor: `${C.cyan}25` }}>
                    <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke={C.cyan} strokeWidth="1.5" strokeLinecap="round" strokeLinejoin="round">
                      <path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z" />
                      <polyline points="14 2 14 8 20 8" />
                      <path d="M12 18v-6" />
                      <path d="M9 15l3-3 3 3" />
                    </svg>
                  </div>
                  <h4 className="text-sm font-semibold mb-2">SCP / SFTP File Transfer</h4>
                  <p className="text-xs text-text-secondary leading-relaxed mb-4">
                    Transfer configuration files, firmware updates, and logs to and from edge servers with progress tracking.
                  </p>

                  {/* File transfer mockup */}
                  <div className="bg-surface rounded-lg border border-border p-3 font-mono text-2xs space-y-1.5">
                    <div style={{ color: C.textMuted }}>$ scp config.yml admin@edge-den-02:/etc/app/</div>
                    <div className="flex items-center gap-2">
                      <span style={{ color: C.text }}>config.yml</span>
                      <div className="flex-1 h-1.5 bg-background rounded-full overflow-hidden">
                        <div className="h-full rounded-full" style={{ width: "100%", background: C.green }} />
                      </div>
                      <span style={{ color: C.green }}>100%</span>
                    </div>
                    <div style={{ color: C.textSec }}>2.4 KB&nbsp;&nbsp;&nbsp;0:00</div>
                    <div className="pt-1 border-t border-border" style={{ borderColor: `${C.border}80` }}>
                      <div style={{ color: C.textMuted }}>$ scp admin@edge-den-02:/var/log/app.log ./</div>
                      <div className="flex items-center gap-2 mt-1">
                        <span style={{ color: C.text }}>app.log</span>
                        <div className="flex-1 h-1.5 bg-background rounded-full overflow-hidden">
                          <div className="h-full rounded-full" style={{ width: "68%", background: C.yellow }} />
                        </div>
                        <span style={{ color: C.yellow }}>68%</span>
                      </div>
                      <div style={{ color: C.textSec }}>45 MB&nbsp;&nbsp;&nbsp;12.3 MB/s&nbsp;&nbsp;&nbsp;ETA 0:02</div>
                    </div>
                  </div>
                </div>
              </ShimmerCard>
            </Reveal>
          </div>

          {/* 3 smaller cards: Tags, RBAC, Audit */}
          <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
            <Reveal delay={0.05}>
              <ShimmerCard className="h-full">
                <div className="bg-card border border-border rounded-xl p-6 h-full hover:border-border-light transition-all duration-300">
                  <div className="w-10 h-10 rounded-lg flex items-center justify-center mb-4 border" style={{ background: `${C.yellow}15`, borderColor: `${C.yellow}25` }}>
                    <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke={C.yellow} strokeWidth="1.5" strokeLinecap="round" strokeLinejoin="round">
                      <path d="M20.59 13.41l-7.17 7.17a2 2 0 0 1-2.83 0L2 12V2h10l8.59 8.59a2 2 0 0 1 0 2.82z" />
                      <line x1="7" y1="7" x2="7.01" y2="7" />
                    </svg>
                  </div>
                  <h4 className="text-sm font-semibold mb-2">Device Tags</h4>
                  <p className="text-xs text-text-secondary leading-relaxed mb-3">
                    Organize edge servers by region, site type, or function for fast filtering and batch operations.
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
                        style={{ color: tag.color, background: `${tag.color}10`, borderColor: `${tag.color}20` }}
                      >
                        {tag.label}
                      </span>
                    ))}
                  </div>
                </div>
              </ShimmerCard>
            </Reveal>

            <Reveal delay={0.1}>
              <ShimmerCard className="h-full">
                <div className="bg-card border border-border rounded-xl p-6 h-full hover:border-border-light transition-all duration-300">
                  <div className="w-10 h-10 rounded-lg flex items-center justify-center mb-4 border" style={{ background: `${C.primary}15`, borderColor: `${C.primary}25` }}>
                    <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke={C.primary} strokeWidth="1.5" strokeLinecap="round" strokeLinejoin="round">
                      <path d="M17 21v-2a4 4 0 0 0-4-4H5a4 4 0 0 0-4 4v2" />
                      <circle cx="9" cy="7" r="4" />
                      <path d="M23 21v-2a4 4 0 0 0-3-3.87" />
                      <path d="M16 3.13a4 4 0 0 1 0 7.75" />
                    </svg>
                  </div>
                  <h4 className="text-sm font-semibold mb-2">RBAC</h4>
                  <p className="text-xs text-text-secondary leading-relaxed mb-3">
                    Give regional teams access to only their edge servers with role-based controls and namespace isolation.
                  </p>
                  <div className="space-y-1.5">
                    {[
                      { role: "NYC Ops Team", access: "northeast-*", color: C.primary },
                      { role: "Field Engineers", access: "tower-*", color: C.yellow },
                      { role: "Warehouse Admins", access: "warehouse-*", color: C.green },
                    ].map((r) => (
                      <div key={r.role} className="flex items-center justify-between text-2xs font-mono">
                        <span style={{ color: C.textSec }}>{r.role}</span>
                        <span style={{ color: r.color }}>{r.access}</span>
                      </div>
                    ))}
                  </div>
                </div>
              </ShimmerCard>
            </Reveal>

            <Reveal delay={0.15}>
              <ShimmerCard className="h-full">
                <div className="bg-card border border-border rounded-xl p-6 h-full hover:border-border-light transition-all duration-300">
                  <div className="w-10 h-10 rounded-lg flex items-center justify-center mb-4 border" style={{ background: `${C.green}15`, borderColor: `${C.green}25` }}>
                    <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke={C.green} strokeWidth="1.5" strokeLinecap="round" strokeLinejoin="round">
                      <path d="M12 20h9" />
                      <path d="M16.5 3.5a2.121 2.121 0 0 1 3 3L7 19l-4 1 1-4L16.5 3.5z" />
                    </svg>
                  </div>
                  <h4 className="text-sm font-semibold mb-2">Audit Trail</h4>
                  <p className="text-xs text-text-secondary leading-relaxed mb-3">
                    Full session recording and command logging across all edge locations for compliance and forensics.
                  </p>
                  <div className="space-y-1.5">
                    {[
                      { time: "14:32", user: "jane", device: "edge-nyc-01", action: "SSH session", color: C.green },
                      { time: "14:28", user: "mike", device: "tower-aus-03", action: "File transfer", color: C.cyan },
                      { time: "14:15", user: "ana", device: "wh-chi-07", action: "SSH session", color: C.green },
                    ].map((log, i) => (
                      <div key={i} className="flex items-center gap-2 text-2xs font-mono">
                        <span style={{ color: C.textMuted }}>{log.time}</span>
                        <span style={{ color: C.primary }}>{log.user}</span>
                        <span style={{ color: C.textMuted }}>&rarr;</span>
                        <span style={{ color: C.textSec }}>{log.device}</span>
                        <span className="ml-auto px-1.5 py-0.5 rounded border" style={{ color: log.color, background: `${log.color}10`, borderColor: `${log.color}20` }}>{log.action}</span>
                      </div>
                    ))}
                  </div>
                </div>
              </ShimmerCard>
            </Reveal>
          </div>
        </div>
      </section>

      {/* ───── Use Case Scenarios ───── */}
      <section className="py-24 border-t border-border">
        <div className="max-w-7xl mx-auto px-8">
          <Reveal className="text-center mb-14">
            <p className="text-2xs font-mono font-semibold uppercase tracking-[0.15em] text-[#7B8EDB] mb-3">
              Real-World Scenarios
            </p>
            <h2 className="text-[clamp(1.75rem,4vw,3rem)] font-bold tracking-[-0.03em] leading-tight mb-4">
              How teams use ShellHub at the edge
            </h2>
            <p className="text-sm text-text-secondary max-w-lg mx-auto leading-relaxed">
              From retail to telecom to logistics, ShellHub powers remote access for edge infrastructure across industries.
            </p>
          </Reveal>

          <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
            {scenarios.map((s, i) => (
              <Reveal key={i} delay={i * 0.08}>
                <ShimmerCard className="h-full">
                  <div className="bg-card border border-border rounded-xl p-6 h-full hover:border-border-light transition-all duration-300">
                    <div className="w-2 h-2 rounded-full mb-4" style={{ background: s.color }} />
                    <h4 className="text-sm font-semibold mb-2">{s.title}</h4>
                    <p className="text-xs text-text-secondary leading-relaxed">{s.desc}</p>
                    {s.mockup}
                  </div>
                </ShimmerCard>
              </Reveal>
            ))}
          </div>
        </div>
      </section>

      {/* ───── CTA ───── */}
      <section className="py-24 border-t border-border">
        <div className="max-w-7xl mx-auto px-8">
          <Reveal>
            <div className="relative bg-card border border-border rounded-2xl p-12 text-center overflow-hidden">
              <ConnectionGrid />
              <div className="absolute inset-0 bg-gradient-to-br from-accent-blue/[0.06] via-transparent to-primary/[0.04] pointer-events-none" />

              <div className="relative z-10">
                <p className="text-2xs font-mono font-semibold uppercase tracking-[0.15em] text-[#7B8EDB] mb-3">
                  Ready to connect your edge?
                </p>
                <h2 className="text-[clamp(1.5rem,3vw,2.25rem)] font-bold tracking-[-0.03em] leading-tight mb-4">
                  Your edge servers, instantly accessible
                </h2>
                <p className="text-sm text-text-secondary max-w-md mx-auto leading-relaxed mb-8">
                  Install the lightweight agent on your edge servers and start managing them remotely in minutes — no infrastructure changes needed.
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
                    Compare Plans
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
