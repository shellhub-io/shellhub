import { useEffect, useState, useCallback } from "react";
import { Navbar } from "../landing/Navbar";
import { Footer } from "@shellhub/design-system/components";
import { Reveal, ShimmerCard, ConnectionGrid } from "../landing/components";
import { C } from "../landing/constants";

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
    icon: (
      <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke={C.cyan} strokeWidth="1.5" strokeLinecap="round">
        <circle cx="12" cy="12" r="10" />
        <polygon points="10 8 16 12 10 16 10 8" />
      </svg>
    ),
    color: C.cyan,
    title: "Session Recording",
    desc: "Every support session is recorded and can be replayed for quality assurance, training, and compliance audits.",
  },
  {
    icon: (
      <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke={C.green} strokeWidth="1.5" strokeLinecap="round">
        <path d="M12 20h9" />
        <path d="M16.5 3.5a2.121 2.121 0 0 1 3 3L7 19l-4 1 1-4L16.5 3.5z" />
      </svg>
    ),
    color: C.green,
    title: "Audit Logging",
    desc: "Complete audit trail with timestamps, user identity, and session details for every connection made to any device.",
  },
  {
    icon: (
      <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke={C.primary} strokeWidth="1.5" strokeLinecap="round">
        <path d="M17 21v-2a4 4 0 0 0-4-4H5a4 4 0 0 0-4 4v2" />
        <circle cx="9" cy="7" r="4" />
        <path d="M23 21v-2a4 4 0 0 0-3-3.87" />
        <path d="M16 3.13a4 4 0 0 1 0 7.75" />
      </svg>
    ),
    color: C.primary,
    title: "Role-Based Access",
    desc: "Assign support engineers access to specific devices or groups — revoke access instantly when the ticket is closed.",
  },
  {
    icon: (
      <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke={C.green} strokeWidth="1.5" strokeLinecap="round">
        <rect x="2" y="3" width="20" height="14" rx="2" />
        <line x1="8" y1="21" x2="16" y2="21" />
        <line x1="12" y1="17" x2="12" y2="21" />
      </svg>
    ),
    color: C.green,
    title: "Web Terminal",
    desc: "Support engineers access devices straight from the browser — no SSH client setup required on their workstations.",
  },
  {
    icon: (
      <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke={C.yellow} strokeWidth="1.5" strokeLinecap="round">
        <rect x="3" y="11" width="18" height="11" rx="2" />
        <path d="M7 11V7a5 5 0 0 1 10 0v4" />
        <circle cx="12" cy="16" r="1" />
      </svg>
    ),
    color: C.yellow,
    title: "MFA for SSH",
    desc: "Require multi-factor authentication for support sessions — adding a security layer beyond SSH keys alone.",
  },
  {
    icon: (
      <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke={C.primary} strokeWidth="1.5" strokeLinecap="round">
        <path d="M12 22s8-4 8-10V5l-8-3-8 3v7c0 6 8 10 8 10z" />
      </svg>
    ),
    color: C.primary,
    title: "Firewall Rules",
    desc: "Restrict support access by IP, time window, or device group with granular, time-based firewall policies.",
  },
];

/* ─── Audit log mock data ──────────────────────────────────────────── */

const auditRows = [
  { time: "14:32:08", initials: "JD", name: "john@company.com", action: "SSH session", device: "prod-server-03", duration: "12m 15s", color: C.primary },
  { time: "14:18:41", initials: "AL", name: "ana@company.com", action: "File transfer", device: "edge-gw-eu-01", duration: "3m 42s", color: C.cyan },
  { time: "13:55:19", initials: "MS", name: "mike@company.com", action: "SSH session", device: "staging-db-02", duration: "8m 04s", color: C.green },
  { time: "13:40:03", initials: "JD", name: "john@company.com", action: "SSH session", device: "iot-gateway-07", duration: "5m 51s", color: C.yellow },
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
  const [navSolid, setNavSolid] = useState(false);
  const [mobileMenu, setMobileMenu] = useState(false);
  const handleScroll = useCallback(() => { setNavSolid(window.scrollY > 50); }, []);
  useEffect(() => {
    window.addEventListener("scroll", handleScroll, { passive: true });
    return () => window.removeEventListener("scroll", handleScroll);
  }, [handleScroll]);

  return (
    <div className="min-h-screen bg-background text-text-primary font-sans antialiased overflow-x-hidden">
      <Navbar navSolid={navSolid} mobileMenu={mobileMenu} setMobileMenu={setMobileMenu} />

      {/* ── Hero ──────────────────────────────────────────────────── */}
      <section className="relative pt-32 pb-24 overflow-hidden">
        <ConnectionGrid />
        <div className="absolute inset-0 bg-gradient-radial from-primary/8 via-transparent to-transparent pointer-events-none" />
        <div className="absolute top-1/3 right-1/4 w-96 h-96 bg-accent-yellow/6 rounded-full blur-3xl pointer-events-none" />

        <div className="max-w-7xl mx-auto px-8 relative z-10 text-center">
          <Reveal>
            <span className="inline-block px-3 py-1 text-2xs font-mono font-semibold uppercase tracking-[0.15em] bg-accent-yellow/10 text-accent-yellow border border-accent-yellow/20 rounded-full mb-6">
              Use Case
            </span>
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
              Give your support team secure, audited access to customer devices — with session recording, full audit trails, and complete accountability for every connection.
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

      {/* ── Session Recording Mockup ─────────────────────────────── */}
      <section className="py-24 border-t border-border">
        <div className="max-w-7xl mx-auto px-8">
          <div className="grid lg:grid-cols-2 gap-12 items-center">
            <div>
              <Reveal>
                <p className="text-2xs font-mono font-semibold uppercase tracking-[0.15em] text-accent-cyan mb-3">
                  Session Recording
                </p>
                <h2 className="text-[clamp(1.75rem,4vw,2.5rem)] font-bold tracking-[-0.03em] leading-tight mb-4">
                  Record every session. Replay any time.
                </h2>
                <p className="text-sm text-text-secondary leading-relaxed mb-8">
                  Every SSH session is automatically captured and stored. Replay sessions to review what happened during a support interaction, verify that procedures were followed, or train new engineers on troubleshooting workflows.
                </p>
              </Reveal>

              <div className="space-y-3">
                {[
                  { label: "Full terminal replay", desc: "Watch sessions frame by frame, exactly as they happened" },
                  { label: "Searchable archive", desc: "Find sessions by user, device, date range, or keyword" },
                  { label: "Compliance-ready exports", desc: "Export session recordings for external audits and compliance reports" },
                  { label: "Incident investigation", desc: "Pinpoint exactly what command caused an issue, down to the second" },
                ].map((item, i) => (
                  <Reveal key={i} delay={i * 0.04}>
                    <div className="flex items-start gap-3">
                      <svg className="w-4 h-4 text-accent-green shrink-0 mt-0.5" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
                        <path strokeLinecap="round" strokeLinejoin="round" d="m4.5 12.75 6 6 9-13.5" />
                      </svg>
                      <div>
                        <p className="text-sm font-medium text-text-primary">{item.label}</p>
                        <p className="text-xs text-text-secondary leading-relaxed">{item.desc}</p>
                      </div>
                    </div>
                  </Reveal>
                ))}
              </div>
            </div>

            {/* Session player mockup */}
            <Reveal delay={0.1}>
              <ShimmerCard className="bg-card border border-border rounded-xl overflow-hidden">
                <div className="p-6">
                  {/* Window chrome */}
                  <div className="flex items-center gap-2 mb-4">
                    <div className="w-3 h-3 rounded-full bg-accent-red/60" />
                    <div className="w-3 h-3 rounded-full bg-accent-yellow/60" />
                    <div className="w-3 h-3 rounded-full bg-accent-green/60" />
                    <span className="ml-2 text-2xs text-text-muted font-mono">Session Replay</span>
                  </div>

                  {/* Terminal area */}
                  <div className="bg-[#111214] rounded-lg border border-border p-4 font-mono text-2xs leading-relaxed mb-4">
                    <p className="text-accent-green">john@prod-server-03:~$<span className="text-text-primary ml-2">ls -la /var/log/nginx/</span></p>
                    <p className="text-text-muted mt-1">total 2184</p>
                    <p className="text-text-muted">drwxr-xr-x  2 root root    4096 Feb 14 09:00 .</p>
                    <p className="text-text-muted">-rw-r-----  1 root adm   842560 Feb 14 14:31 access.log</p>
                    <p className="text-text-muted">-rw-r-----  1 root adm   194720 Feb 14 14:28 error.log</p>
                    <p className="text-accent-green mt-2">john@prod-server-03:~$<span className="text-text-primary ml-2">systemctl status nginx</span></p>
                    <p className="text-text-muted mt-1">
                      <span className="text-accent-green">●</span> nginx.service - A high performance web server
                    </p>
                    <p className="text-text-muted">
                      {"   "}Loaded: loaded (/lib/systemd/system/nginx.service; enabled)
                    </p>
                    <p className="text-text-muted">
                      {"   "}Active: <span className="text-accent-green">active (running)</span> since Fri 2026-02-14 09:00:12 UTC
                    </p>
                    <p className="text-accent-green mt-2">john@prod-server-03:~$<span className="text-text-primary ml-2 animate-pulse">_</span></p>
                  </div>

                  {/* Playback controls */}
                  <div className="flex items-center gap-3 bg-surface rounded-lg border border-border p-3">
                    <button className="w-8 h-8 rounded-lg bg-accent-cyan/10 border border-accent-cyan/20 flex items-center justify-center shrink-0">
                      <svg width="14" height="14" viewBox="0 0 24 24" fill={C.cyan} stroke="none">
                        <polygon points="6 4 20 12 6 20 6 4" />
                      </svg>
                    </button>

                    {/* Timeline */}
                    <div className="flex-1 relative">
                      <div className="h-1.5 bg-border rounded-full overflow-hidden">
                        <div className="h-full rounded-full bg-gradient-to-r from-accent-cyan to-primary" style={{ width: "30%" }} />
                      </div>
                    </div>

                    <span className="text-2xs text-text-muted font-mono shrink-0">00:03:42 / 00:12:15</span>
                  </div>

                  {/* Metadata */}
                  <div className="mt-4 pt-4 border-t border-border">
                    <p className="text-2xs text-text-muted font-mono">
                      User: <span className="text-text-secondary">john@company.com</span>
                      <span className="mx-2 text-border">|</span>
                      Device: <span className="text-text-secondary">prod-server-03</span>
                      <span className="mx-2 text-border">|</span>
                      Duration: <span className="text-text-secondary">12m 15s</span>
                    </p>
                  </div>
                </div>
              </ShimmerCard>
            </Reveal>
          </div>
        </div>
      </section>

      {/* ── Audit Trail Mockup ───────────────────────────────────── */}
      <section className="py-24 border-t border-border">
        <div className="max-w-7xl mx-auto px-8">
          <div className="grid lg:grid-cols-2 gap-12 items-center">
            {/* Audit log table mockup */}
            <Reveal delay={0.1} className="order-2 lg:order-1">
              <ShimmerCard className="bg-card border border-border rounded-xl overflow-hidden">
                <div className="p-6">
                  {/* Window chrome */}
                  <div className="flex items-center gap-2 mb-5">
                    <div className="w-3 h-3 rounded-full bg-accent-red/60" />
                    <div className="w-3 h-3 rounded-full bg-accent-yellow/60" />
                    <div className="w-3 h-3 rounded-full bg-accent-green/60" />
                    <span className="ml-2 text-2xs text-text-muted font-mono">Audit Log</span>
                  </div>

                  {/* Table header */}
                  <div className="grid grid-cols-[60px_1fr_1fr_1fr_70px] gap-2 px-3 py-2 mb-1">
                    <span className="text-2xs font-mono font-semibold text-text-muted uppercase tracking-wider">Time</span>
                    <span className="text-2xs font-mono font-semibold text-text-muted uppercase tracking-wider">User</span>
                    <span className="text-2xs font-mono font-semibold text-text-muted uppercase tracking-wider">Action</span>
                    <span className="text-2xs font-mono font-semibold text-text-muted uppercase tracking-wider">Device</span>
                    <span className="text-2xs font-mono font-semibold text-text-muted uppercase tracking-wider">Duration</span>
                  </div>

                  {/* Rows */}
                  <div className="space-y-1">
                    {auditRows.map((row, i) => (
                      <div
                        key={i}
                        className={`grid grid-cols-[60px_1fr_1fr_1fr_70px] gap-2 px-3 py-2.5 rounded-lg items-center ${
                          i % 2 === 0 ? "bg-surface" : "bg-transparent"
                        }`}
                      >
                        <span className="text-2xs font-mono text-text-muted">{row.time}</span>
                        <div className="flex items-center gap-2 min-w-0">
                          <div
                            className="w-6 h-6 rounded-full flex items-center justify-center text-[9px] font-semibold shrink-0"
                            style={{ background: `${row.color}20`, color: row.color }}
                          >
                            {row.initials}
                          </div>
                          <span className="text-2xs text-text-secondary truncate">{row.name}</span>
                        </div>
                        <span className="text-2xs text-text-secondary font-mono">{row.action}</span>
                        <span className="text-2xs text-text-secondary font-mono truncate">{row.device}</span>
                        <span className="text-2xs text-text-muted font-mono">{row.duration}</span>
                      </div>
                    ))}
                  </div>

                  {/* Footer */}
                  <div className="mt-4 pt-4 border-t border-border flex items-center justify-between">
                    <span className="text-2xs text-text-muted">Showing 4 of 1,247 entries</span>
                    <span className="text-2xs text-primary font-medium">View all &rarr;</span>
                  </div>
                </div>
              </ShimmerCard>
            </Reveal>

            {/* Text */}
            <div className="order-1 lg:order-2">
              <Reveal>
                <p className="text-2xs font-mono font-semibold uppercase tracking-[0.15em] text-accent-green mb-3">
                  Audit Trail
                </p>
                <h2 className="text-[clamp(1.75rem,4vw,2.5rem)] font-bold tracking-[-0.03em] leading-tight mb-4">
                  Full accountability for every connection
                </h2>
                <p className="text-sm text-text-secondary leading-relaxed mb-8">
                  Every session is logged with the user identity, target device, timestamps, and duration. Build compliance reports, investigate incidents, and prove to auditors exactly who did what, when, and for how long.
                </p>
              </Reveal>

              <div className="space-y-3">
                {[
                  { label: "Immutable log entries", desc: "Audit records cannot be altered or deleted by users" },
                  { label: "Filter and search", desc: "Find events by user, device, date, or action type instantly" },
                  { label: "Compliance reporting", desc: "Generate reports for SOC 2, ISO 27001, and internal audits" },
                  { label: "Real-time monitoring", desc: "View active sessions and who is currently connected to which device" },
                ].map((item, i) => (
                  <Reveal key={i} delay={i * 0.04}>
                    <div className="flex items-start gap-3">
                      <svg className="w-4 h-4 text-accent-green shrink-0 mt-0.5" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
                        <path strokeLinecap="round" strokeLinejoin="round" d="m4.5 12.75 6 6 9-13.5" />
                      </svg>
                      <div>
                        <p className="text-sm font-medium text-text-primary">{item.label}</p>
                        <p className="text-xs text-text-secondary leading-relaxed">{item.desc}</p>
                      </div>
                    </div>
                  </Reveal>
                ))}
              </div>
            </div>
          </div>
        </div>
      </section>

      {/* ── Pain Points ──────────────────────────────────────────── */}
      <section className="py-24 border-t border-border">
        <div className="max-w-7xl mx-auto px-8">
          <Reveal className="text-center mb-14">
            <p className="text-2xs font-mono font-semibold uppercase tracking-[0.15em] text-[#7B8EDB] mb-3">
              The Problem
            </p>
            <h2 className="text-[clamp(1.75rem,4vw,3rem)] font-bold tracking-[-0.03em] leading-tight mb-4">
              Why traditional remote support falls short
            </h2>
            <p className="text-sm text-text-secondary max-w-lg mx-auto leading-relaxed">
              Legacy workflows leave gaps in security, visibility, and compliance that put your organization at risk.
            </p>
          </Reveal>

          <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
            {painPoints.map((p, i) => (
              <Reveal key={i} delay={i * 0.06}>
                <ShimmerCard className="h-full">
                  <div className="bg-card border border-border rounded-xl p-6 hover:border-border-light transition-all duration-300 h-full">
                    <div className="w-2 h-2 rounded-full mb-4" style={{ background: p.color }} />
                    <h4 className="text-sm font-semibold mb-2">{p.title}</h4>
                    <p className="text-xs text-text-secondary leading-relaxed">{p.desc}</p>
                  </div>
                </ShimmerCard>
              </Reveal>
            ))}
          </div>
        </div>
      </section>

      {/* ── Key Features ─────────────────────────────────────────── */}
      <section className="py-24 border-t border-border">
        <div className="max-w-7xl mx-auto px-8">
          <Reveal className="text-center mb-14">
            <p className="text-2xs font-mono font-semibold uppercase tracking-[0.15em] text-[#7B8EDB] mb-3">
              Features
            </p>
            <h2 className="text-[clamp(1.75rem,4vw,3rem)] font-bold tracking-[-0.03em] leading-tight mb-4">
              Built for audited, accountable support
            </h2>
            <p className="text-sm text-text-secondary max-w-lg mx-auto leading-relaxed">
              Everything your support team needs to connect securely, troubleshoot efficiently, and maintain a complete compliance record.
            </p>
          </Reveal>

          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
            {features.map((f, i) => (
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

      {/* ── Support Workflow ──────────────────────────────────────── */}
      <section className="py-24 border-t border-border">
        <div className="max-w-7xl mx-auto px-8">
          <Reveal className="text-center mb-14">
            <p className="text-2xs font-mono font-semibold uppercase tracking-[0.15em] text-[#7B8EDB] mb-3">
              Workflow
            </p>
            <h2 className="text-[clamp(1.75rem,4vw,3rem)] font-bold tracking-[-0.03em] leading-tight mb-4">
              From ticket to resolution — fully audited
            </h2>
            <p className="text-sm text-text-secondary max-w-lg mx-auto leading-relaxed">
              A streamlined support workflow that captures everything for compliance without slowing your team down.
            </p>
          </Reveal>

          <div className="grid grid-cols-1 md:grid-cols-3 gap-6 relative">
            {/* Connecting arrows (desktop only) */}
            <div className="hidden md:block absolute top-1/2 left-[calc(33.33%-12px)] w-[calc(33.33%+24px)] -translate-y-1/2 pointer-events-none z-0">
              <svg className="w-full h-8" viewBox="0 0 400 32" fill="none" preserveAspectRatio="none">
                <line x1="0" y1="16" x2="180" y2="16" stroke={C.border} strokeWidth="1" strokeDasharray="6 4" />
                <polygon points="180,12 188,16 180,20" fill={C.textMuted} />
                <line x1="210" y1="16" x2="390" y2="16" stroke={C.border} strokeWidth="1" strokeDasharray="6 4" />
                <polygon points="390,12 398,16 390,20" fill={C.textMuted} />
              </svg>
            </div>

            {workflowSteps.map((step, i) => (
              <Reveal key={i} delay={i * 0.08}>
                <ShimmerCard className="h-full">
                  <div className="relative bg-card border border-border rounded-xl p-6 hover:border-border-light transition-all duration-300 h-full overflow-hidden">
                    <div className="absolute top-0 left-0 w-full h-0.5" style={{ background: `linear-gradient(90deg, ${step.color}00, ${step.color}, ${step.color}00)` }} />
                    <div className="flex items-center gap-3 mb-4">
                      <span
                        className="w-10 h-10 rounded-lg flex items-center justify-center text-xs font-bold font-mono border"
                        style={{ background: `${step.color}15`, borderColor: `${step.color}25`, color: step.color }}
                      >
                        {step.num}
                      </span>
                      <h4 className="text-sm font-semibold">{step.title}</h4>
                    </div>
                    <p className="text-xs text-text-secondary leading-relaxed">{step.desc}</p>
                  </div>
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
                  Ready to get started?
                </p>
                <h2 className="text-[clamp(1.5rem,3vw,2.25rem)] font-bold tracking-[-0.03em] leading-tight mb-4">
                  Secure support starts here
                </h2>
                <p className="text-sm text-text-secondary max-w-md mx-auto leading-relaxed mb-8">
                  Enable audited, accountable remote support for your team in minutes. No VPN required, no SSH keys to manage, no audit gaps.
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
