import { Reveal, ConnectionGrid } from "../landing/components";

export function EnterpriseCTA() {
  return (
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
                Talk to our team
              </h2>
              <p className="text-sm text-text-secondary max-w-md mx-auto leading-relaxed mb-8">
                Get a demo, discuss your requirements, and find the right plan for your organization. Our team typically responds within one business day.
              </p>

              <div className="flex flex-col sm:flex-row items-center justify-center gap-3">
                <a
                  href="mailto:sales@shellhub.io"
                  className="relative inline-flex items-center gap-2.5 px-8 py-3.5 text-[15px] font-semibold bg-primary text-[#111214] rounded-xl shadow-[0_0_20px_rgba(102,122,204,0.3)] hover:shadow-[0_0_32px_rgba(102,122,204,0.5)] hover:scale-[1.03] active:scale-[0.98] transition-all duration-300 group overflow-hidden"
                >
                  <span className="absolute inset-0 bg-gradient-to-r from-white/0 via-white/25 to-white/0 translate-x-[-200%] group-hover:translate-x-[200%] transition-transform duration-700 ease-in-out" />
                  <svg className="relative w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
                    <path strokeLinecap="round" strokeLinejoin="round" d="M21.75 6.75v10.5a2.25 2.25 0 0 1-2.25 2.25h-15a2.25 2.25 0 0 1-2.25-2.25V6.75m19.5 0A2.25 2.25 0 0 0 19.5 4.5h-15a2.25 2.25 0 0 0-2.25 2.25m19.5 0v.243a2.25 2.25 0 0 1-1.07 1.916l-7.5 4.615a2.25 2.25 0 0 1-2.36 0L3.32 8.91a2.25 2.25 0 0 1-1.07-1.916V6.75" />
                  </svg>
                  <span className="relative">Contact Sales</span>
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
  );
}
