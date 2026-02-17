import { Reveal, ConnectionGrid } from "../landing/components";

export function HeroEnterprise() {
  return (
    <section className="relative pt-32 pb-24 overflow-hidden">
      <ConnectionGrid />
      <div className="absolute inset-0 bg-gradient-radial from-primary/8 via-transparent to-transparent pointer-events-none" />
      <div className="absolute top-1/3 right-1/4 w-96 h-96 bg-primary/6 rounded-full blur-3xl pointer-events-none" />

      <div className="max-w-7xl mx-auto px-8 relative z-10 text-center">
        <Reveal>
          <span className="inline-block px-3 py-1 text-2xs font-mono font-semibold uppercase tracking-[0.15em] bg-accent-yellow/10 text-accent-yellow border border-accent-yellow/20 rounded-full mb-6">
            Enterprise
          </span>
        </Reveal>
        <Reveal>
          <h1 className="text-[clamp(2rem,5vw,3.5rem)] font-bold tracking-[-0.03em] leading-[1.1] mb-6 max-w-3xl mx-auto">
            Device management built for{" "}
            <span className="bg-gradient-to-r from-primary via-accent-blue to-accent-cyan bg-clip-text text-transparent">
              scale and control
            </span>
          </h1>
        </Reveal>
        <Reveal>
          <p className="text-sm text-text-secondary max-w-lg mx-auto leading-relaxed mb-10">
            SSO, admin panel, audit logs, and dedicated support. Everything your team needs to manage thousands of devices securely.
          </p>
        </Reveal>
        <Reveal>
          <div className="flex flex-col sm:flex-row items-center justify-center gap-3">
            <a
              href="/v2/pricing"
              className="relative inline-flex items-center gap-2.5 px-8 py-3.5 text-[15px] font-semibold bg-primary text-[#111214] rounded-xl shadow-[0_0_20px_rgba(102,122,204,0.3)] hover:shadow-[0_0_32px_rgba(102,122,204,0.5)] hover:scale-[1.03] active:scale-[0.98] transition-all duration-300 group overflow-hidden"
            >
              <span className="absolute inset-0 bg-gradient-to-r from-white/0 via-white/25 to-white/0 translate-x-[-200%] group-hover:translate-x-[200%] transition-transform duration-700 ease-in-out" />
              <span className="relative">Get a Quote</span>
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
  );
}
