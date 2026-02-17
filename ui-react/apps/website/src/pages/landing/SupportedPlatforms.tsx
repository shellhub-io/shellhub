import { Reveal } from "./components";

export function SupportedPlatforms() {
  return (
    <section className="py-24 border-t border-border relative overflow-hidden">
      <div className="absolute top-1/2 left-1/2 -translate-x-1/2 -translate-y-1/2 w-[700px] h-[700px] bg-primary/[0.03] rounded-full blur-3xl pointer-events-none" />
      <div className="max-w-7xl mx-auto px-8 relative z-10">
        <Reveal className="text-center mb-14">
          <p className="text-2xs font-mono font-semibold uppercase tracking-[0.15em] text-[#7B8EDB] mb-3">Supported Platforms</p>
          <h2 className="text-[clamp(1.75rem,4vw,3rem)] font-bold tracking-[-0.03em] leading-tight mb-4">One agent. <span className="text-primary">Every platform.</span></h2>
          <p className="text-sm text-text-secondary max-w-xl mx-auto leading-relaxed">The ShellHub agent runs on{" "}
            <span className="inline-block px-1.5 py-0.5 bg-border/40 rounded text-2xs font-mono font-medium text-text-primary/90">x86</span>{" "}and{" "}
            <span className="inline-block px-1.5 py-0.5 bg-border/40 rounded text-2xs font-mono font-medium text-text-primary/90">ARM</span>{" "}
            — from Docker containers to embedded Linux and FreeBSD.
          </p>
        </Reveal>

        {/* Bento Grid — 12-col for precise control */}
        <div className="grid grid-cols-2 lg:grid-cols-12 gap-3 auto-rows-auto">

          {/* Docker — featured (5 cols, 2 rows) */}
          <Reveal className="col-span-2 lg:col-span-5 lg:row-span-2">
            <div className="h-full bg-card border border-border rounded-xl p-6 lg:p-7 hover:border-[#1D63ED]/30 transition-all duration-300 group relative overflow-hidden">
              <div className="absolute -right-6 -bottom-6 w-32 h-32 bg-[#1D63ED]/[0.04] rounded-full blur-2xl pointer-events-none" />
              <div className="relative">
                <div className="flex items-center gap-3 mb-5">
                  <div className="w-10 h-10 rounded-lg flex items-center justify-center bg-[#1D63ED]/10 border border-[#1D63ED]/20 text-[#1D63ED]">
                    <svg className="w-5 h-5" viewBox="0 0 24 24" fill="currentColor">
                      <path d="M13.983 11.078h2.119a.186.186 0 0 0 .186-.185V9.006a.186.186 0 0 0-.186-.186h-2.119a.185.185 0 0 0-.185.185v1.888c0 .102.083.185.185.185m-2.954-5.43h2.118a.186.186 0 0 0 .186-.186V3.574a.186.186 0 0 0-.186-.185h-2.118a.185.185 0 0 0-.185.185v1.888c0 .102.082.185.185.186m0 2.716h2.118a.187.187 0 0 0 .186-.186V6.29a.186.186 0 0 0-.186-.185h-2.118a.185.185 0 0 0-.185.185v1.887c0 .102.082.185.185.186m-2.93 0h2.12a.186.186 0 0 0 .184-.186V6.29a.185.185 0 0 0-.185-.185H8.1a.185.185 0 0 0-.185.185v1.887c0 .102.083.185.185.186m-2.964 0h2.119a.186.186 0 0 0 .185-.186V6.29a.186.186 0 0 0-.185-.185H5.136a.186.186 0 0 0-.186.185v1.887c0 .102.084.185.186.186m5.893 2.715h2.118a.186.186 0 0 0 .186-.185V9.006a.186.186 0 0 0-.186-.186h-2.118a.185.185 0 0 0-.185.185v1.888c0 .102.082.185.185.185m-2.93 0h2.12a.185.185 0 0 0 .184-.185V9.006a.185.185 0 0 0-.184-.186h-2.12a.185.185 0 0 0-.184.185v1.888c0 .102.083.185.185.185m-2.964 0h2.119a.185.185 0 0 0 .185-.185V9.006a.186.186 0 0 0-.185-.186H5.136a.186.186 0 0 0-.186.185v1.888c0 .102.084.185.186.185m-2.92 0h2.12a.185.185 0 0 0 .184-.185V9.006a.185.185 0 0 0-.184-.186h-2.12a.185.185 0 0 0-.184.185v1.888c0 .102.082.185.185.185M23.763 9.89c-.065-.051-.672-.51-1.954-.51-.338.001-.676.03-1.01.087-.248-1.7-1.653-2.53-1.716-2.566l-.344-.199-.226.327c-.284.438-.49.922-.612 1.43-.23.97-.09 1.882.403 2.661-.595.332-1.55.413-1.744.42H.751a.751.751 0 0 0-.75.748 11.376 11.376 0 0 0 .692 4.062c.545 1.428 1.355 2.48 2.41 3.124 1.18.723 3.1 1.137 5.275 1.137.983.003 1.963-.086 2.93-.266a12.248 12.248 0 0 0 3.823-1.389c.98-.567 1.86-1.288 2.61-2.136 1.252-1.418 1.998-2.997 2.553-4.4h.221c1.372 0 2.215-.549 2.68-1.009.309-.293.55-.65.707-1.046l.098-.288Z" />
                    </svg>
                  </div>
                  <span className="text-2xs font-mono font-semibold text-[#1D63ED] uppercase tracking-[0.1em]">Most Popular</span>
                </div>
                <h3 className="text-xl font-bold tracking-[-0.02em] mb-2">Docker</h3>
                <p className="text-sm text-text-secondary leading-relaxed mb-5">Run the agent as a container. The recommended way to get started.</p>
                <div className="bg-surface border border-border rounded-lg px-4 py-3">
                  <code className="font-mono text-xs text-text-secondary"><span className="text-primary">$</span> curl -sSf &lt;server&gt;/install.sh | sh</code>
                </div>
              </div>
            </div>
          </Reveal>

          {/* Snap — wider card (4 cols) */}
          <Reveal delay={0.04} className="col-span-1 lg:col-span-4">
            <div className="h-full bg-card border border-border rounded-xl p-5 hover:border-accent-green/30 transition-all duration-300">
              <div className="flex items-start gap-3">
                <div className="w-9 h-9 rounded-lg flex shrink-0 items-center justify-center bg-accent-green/10 border border-accent-green/20 text-accent-green">
                  <svg className="w-[18px] h-[18px]" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={1.5}>
                    <path strokeLinecap="round" strokeLinejoin="round" d="M20.25 7.5l-.625 10.632a2.25 2.25 0 0 1-2.247 2.118H6.622a2.25 2.25 0 0 1-2.247-2.118L3.75 7.5M10 11.25h4M3.375 7.5h17.25c.621 0 1.125-.504 1.125-1.125v-1.5c0-.621-.504-1.125-1.125-1.125H3.375c-.621 0-1.125.504-1.125 1.125v1.5c0 .621.504 1.125 1.125 1.125Z" />
                  </svg>
                </div>
                <div>
                  <h4 className="text-sm font-bold mb-1">Snap</h4>
                  <p className="text-xs text-text-secondary leading-relaxed">Install from the Snap Store with automatic updates and rollback support.</p>
                </div>
              </div>
            </div>
          </Reveal>

          {/* Podman (3 cols) */}
          <Reveal delay={0.08} className="col-span-1 lg:col-span-3">
            <div className="h-full bg-card border border-border rounded-xl p-5 hover:border-primary/30 transition-all duration-300">
              <div className="w-9 h-9 rounded-lg flex items-center justify-center bg-primary/10 border border-primary/20 text-primary mb-3">
                <svg className="w-[18px] h-[18px]" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={1.5}>
                  <path strokeLinecap="round" strokeLinejoin="round" d="m21 7.5-9-5.25L3 7.5m18 0-9 5.25m9-5.25v9l-9 5.25M3 7.5l9 5.25M3 7.5v9l9 5.25m0-9v9" />
                </svg>
              </div>
              <h4 className="text-sm font-bold mb-1">Podman</h4>
              <p className="text-xs text-text-secondary leading-relaxed">Same container workflow as Docker, but rootless and daemonless.</p>
            </div>
          </Reveal>

          {/* Standalone (4 cols) */}
          <Reveal delay={0.12} className="col-span-1 lg:col-span-4">
            <div className="h-full bg-card border border-border rounded-xl p-5 hover:border-accent-yellow/30 transition-all duration-300">
              <div className="flex items-start gap-3">
                <div className="w-9 h-9 rounded-lg flex shrink-0 items-center justify-center bg-accent-yellow/10 border border-accent-yellow/20 text-accent-yellow">
                  <svg className="w-[18px] h-[18px]" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={1.5}>
                    <path strokeLinecap="round" strokeLinejoin="round" d="M5.25 14.25h13.5m-13.5 0a3 3 0 0 1-3-3m3 3a3 3 0 1 0 0 6h13.5a3 3 0 1 0 0-6m-16.5-3a3 3 0 0 1 3-3h13.5a3 3 0 0 1 3 3m-19.5 0a4.5 4.5 0 0 1 .9-2.7L5.737 5.1a3.375 3.375 0 0 1 2.7-1.35h7.126c1.062 0 2.062.5 2.7 1.35l2.587 3.45a4.5 4.5 0 0 1 .9 2.7" />
                  </svg>
                </div>
                <div>
                  <h4 className="text-sm font-bold mb-1">Standalone</h4>
                  <p className="text-xs text-text-secondary leading-relaxed">Direct binary managed by systemd. No container runtime needed.</p>
                </div>
              </div>
            </div>
          </Reveal>

          {/* WSL (3 cols) */}
          <Reveal delay={0.16} className="col-span-1 lg:col-span-3">
            <div className="h-full bg-card border border-border rounded-xl p-5 hover:border-accent-cyan/30 transition-all duration-300">
              <div className="w-9 h-9 rounded-lg flex items-center justify-center bg-accent-cyan/10 border border-accent-cyan/20 text-accent-cyan mb-3">
                <svg className="w-[18px] h-[18px]" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={1.5}>
                  <path strokeLinecap="round" strokeLinejoin="round" d="M9 17.25v1.007a3 3 0 0 1-.879 2.122L7.5 21h9l-.621-.621A3 3 0 0 1 15 18.257V17.25m6-12V15a2.25 2.25 0 0 1-2.25 2.25H5.25A2.25 2.25 0 0 1 3 15V5.25m18 0A2.25 2.25 0 0 0 18.75 3H5.25A2.25 2.25 0 0 0 3 5.25m18 0V12a2.25 2.25 0 0 1-2.25 2.25H5.25A2.25 2.25 0 0 1 3 12V5.25" />
                </svg>
              </div>
              <h4 className="text-sm font-bold mb-1">WSL</h4>
              <p className="text-xs text-text-secondary leading-relaxed">Run the agent inside WSL to manage Windows dev environments.</p>
            </div>
          </Reveal>

          {/* ─── Embedded row — full width, horizontal strip ─── */}
          <Reveal delay={0.20} className="col-span-2 lg:col-span-12">
            <div className="grid grid-cols-3 gap-3">
              {[
                { name: "Yocto Project", sub: "Integrate via our official meta-layer", color: "accent-cyan", href: "https://github.com/shellhub-io/meta-shellhub",
                  icon: <path strokeLinecap="round" strokeLinejoin="round" d="M8.25 3v1.5M4.5 8.25H3m18 0h-1.5M4.5 12H3m18 0h-1.5m-15 3.75H3m18 0h-1.5M8.25 19.5V21M12 3v1.5m0 15V21m3.75-18v1.5m0 15V21m-9-1.5h10.5a2.25 2.25 0 0 0 2.25-2.25V6.75a2.25 2.25 0 0 0-2.25-2.25H6.75A2.25 2.25 0 0 0 4.5 6.75v10.5a2.25 2.25 0 0 0 2.25 2.25Zm.75-12h9v9h-9v-9Z" /> },
                { name: "Buildroot", sub: "Add via our external tree package", color: "accent-yellow", href: "https://github.com/shellhub-io/buildroot",
                  icon: <path strokeLinecap="round" strokeLinejoin="round" d="M11.42 15.17 17.25 21A2.652 2.652 0 0 0 21 17.25l-5.877-5.877M11.42 15.17l2.496-3.03c.317-.384.74-.626 1.208-.766M11.42 15.17l-4.655 5.653a2.548 2.548 0 1 1-3.586-3.586l6.837-5.63m5.108-.233c.55-.164 1.163-.188 1.743-.14a4.5 4.5 0 0 0 4.486-6.336l-3.276 3.277a3.004 3.004 0 0 1-2.25-2.25l3.276-3.276a4.5 4.5 0 0 0-6.336 4.486c.049.58.025 1.193-.14 1.743" /> },
                { name: "FreeBSD", sub: "Runs natively with a dedicated ports entry", color: "accent-red", href: "https://github.com/shellhub-io/ports",
                  icon: <path strokeLinecap="round" strokeLinejoin="round" d="M6.75 7.5l3 2.25-3 2.25m4.5 0h3m-9 8.25h13.5A2.25 2.25 0 0 0 21 18V6a2.25 2.25 0 0 0-2.25-2.25H5.25A2.25 2.25 0 0 0 3 6v12a2.25 2.25 0 0 0 2.25 2.25Z" /> },
              ].map((p) => (
                <a key={p.name} href={p.href} target="_blank" rel="noopener noreferrer" className="bg-surface border border-border rounded-xl p-4 hover:border-border-light transition-all duration-300 group">
                  <div className="flex items-center gap-3">
                    <div className={`w-8 h-8 rounded-md flex items-center justify-center bg-${p.color}/10 border border-${p.color}/20 text-${p.color}`}>
                      <svg className="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={1.5}>{p.icon}</svg>
                    </div>
                    <div className="flex-1">
                      <h4 className="text-sm font-bold leading-tight group-hover:text-text-primary transition-colors">{p.name}</h4>
                      <span className="text-2xs text-text-secondary">{p.sub}</span>
                    </div>
                    <svg className="w-3.5 h-3.5 text-text-muted group-hover:text-text-secondary transition-colors shrink-0" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}><path strokeLinecap="round" strokeLinejoin="round" d="M13.5 6H5.25A2.25 2.25 0 0 0 3 8.25v10.5A2.25 2.25 0 0 0 5.25 21h10.5A2.25 2.25 0 0 0 18 18.75V10.5m-10.5 6L21 3m0 0h-5.25M21 3v5.25" /></svg>
                  </div>
                </a>
              ))}
            </div>
          </Reveal>
        </div>

        <Reveal className="text-center mt-10">
          <a href="#" className="inline-flex items-center gap-1.5 text-xs font-medium text-[#7B8EDB] hover:gap-2.5 transition-all group">
            View all supported platforms
            <svg className="w-3 h-3 group-hover:translate-x-0.5 transition-transform" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2.5}><path strokeLinecap="round" strokeLinejoin="round" d="M13.5 4.5 21 12m0 0-7.5 7.5M21 12H3" /></svg>
          </a>
        </Reveal>
      </div>
    </section>
  );
}
