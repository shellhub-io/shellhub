import { ConnectionGrid } from "./components";

export function Hero() {
  return (
    <section className="min-h-screen flex flex-col items-center justify-center text-center px-6 pt-28 pb-20 relative overflow-hidden grid-bg">
      <ConnectionGrid />
      <div className="absolute inset-0 bg-gradient-radial from-primary/10 via-transparent to-transparent pointer-events-none" />
      <div className="absolute top-16 left-1/4 w-[500px] h-[500px] bg-primary/8 rounded-full blur-3xl pointer-events-none" />
      <div className="absolute bottom-0 right-1/4 w-96 h-96 bg-accent-cyan/6 rounded-full blur-3xl pointer-events-none" />
      <div className="absolute top-1/3 right-[10%] w-72 h-72 bg-primary/5 rounded-full blur-3xl pointer-events-none" />

      <div className="relative z-10 max-w-4xl flex flex-col items-center">
        {/* Floating ShellHub cloud */}
        <div className="animate-float mb-8 inline-block">
          <img src="/v2/cloud-icon.svg" alt="" className="h-16 drop-shadow-[0_0_24px_rgba(102,122,204,0.35)]" />
        </div>

        {/* Badge like app's active nav pill */}
        <div className="inline-flex items-center gap-2 px-3 py-1.5 rounded-full bg-primary/[0.06] border border-primary/20 mb-8 animate-fade-in">
          <span className="relative flex h-2 w-2">
            <span className="animate-ping absolute inline-flex h-full w-full rounded-full bg-primary opacity-75" />
            <span className="relative inline-flex rounded-full h-2 w-2 bg-primary" />
          </span>
          <span className="text-2xs font-mono font-semibold uppercase tracking-[0.15em] text-[#7B8EDB]">Open Source SSH Gateway</span>
        </div>

        <h1 className="text-[clamp(2.5rem,6.5vw,5rem)] font-bold leading-[1.08] tracking-[-0.035em] mb-6 animate-fade-in" style={{ animationDelay: "100ms" }}>
          All your devices.{" "}
          <span className="text-primary">One gateway.</span>
        </h1>

        <p className="text-base sm:text-lg text-text-secondary max-w-xl mx-auto mb-10 leading-relaxed animate-fade-in" style={{ animationDelay: "200ms" }}>
          ShellHub is a centralized SSH gateway for remote access to Linux devices. No public IPs, no VPNs, no firewall changes.
        </p>

        <div className="flex gap-3 flex-wrap justify-center animate-fade-in" style={{ animationDelay: "300ms" }}>
          <a href="/v2/getting-started" className="inline-flex items-center gap-2.5 px-8 py-3.5 text-[15px] font-semibold bg-primary border border-primary-400/40 text-white rounded-xl animate-[cta-glow_2.5s_ease-in-out_infinite] hover:brightness-110 hover:border-primary-400/60 hover:scale-[1.03] active:scale-[0.98] transition-all duration-300 group">
            Get Started
            <svg className="w-4 h-4 group-hover:translate-x-1 transition-transform duration-300" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2.5}><path strokeLinecap="round" strokeLinejoin="round" d="M13.5 4.5 21 12m0 0-7.5 7.5M21 12H3" /></svg>
          </a>

        </div>
      </div>
    </section>
  );
}
