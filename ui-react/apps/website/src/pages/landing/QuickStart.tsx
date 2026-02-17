import { Reveal, CopyBtn } from "./components";

export function QuickStart() {
  return (
    <section className="py-24">
      <div className="max-w-7xl mx-auto px-8">
        <Reveal className="text-center mb-10">
          <p className="text-2xs font-mono font-semibold uppercase tracking-[0.15em] text-[#7B8EDB] mb-3">Quick Start</p>
          <h2 className="text-[clamp(1.75rem,4vw,3rem)] font-bold tracking-[-0.03em] leading-tight mb-3">Try ShellHub in seconds.</h2>
          <p className="text-sm text-text-secondary max-w-md mx-auto leading-relaxed">Get up and running with a single command.</p>
        </Reveal>

        <Reveal>
          <div className="max-w-xl mx-auto bg-card border border-border rounded-xl overflow-hidden">
            <div className="flex items-center gap-1.5 px-4 py-2.5 border-b border-border bg-surface">
              <span className="w-2.5 h-2.5 rounded-full bg-accent-red/70" />
              <span className="w-2.5 h-2.5 rounded-full bg-accent-yellow/70" />
              <span className="w-2.5 h-2.5 rounded-full bg-accent-green/70" />
              <span className="flex-1 text-center text-2xs font-mono text-text-secondary">terminal</span>
            </div>
            <div className="px-5 py-4 flex items-center justify-between gap-4">
              <code className="font-mono text-sm text-text-secondary leading-relaxed">
                <span className="text-primary">$ </span>
                <span className="text-text-primary">docker run -d -p 80:80 shellhubio/shellhub</span>
              </code>
              <CopyBtn text="docker run -d -p 80:80 shellhubio/shellhub" />
            </div>
          </div>
        </Reveal>

        <Reveal className="text-center mt-5">
          <a href="#" className="inline-flex items-center gap-1.5 text-xs font-medium text-[#7B8EDB] hover:gap-2.5 transition-all group">
            Full installation guide
            <svg className="w-3 h-3 group-hover:translate-x-0.5 transition-transform" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2.5}><path strokeLinecap="round" strokeLinejoin="round" d="M13.5 4.5 21 12m0 0-7.5 7.5M21 12H3" /></svg>
          </a>
        </Reveal>
      </div>
    </section>
  );
}
