import { Reveal, ShimmerCard } from "../landing/components";

export function DeploymentOptions() {
  return (
    <section className="py-24 border-t border-border">
      <div className="max-w-7xl mx-auto px-8">
        <Reveal className="text-center mb-14">
          <p className="text-2xs font-mono font-semibold uppercase tracking-[0.15em] text-[#7B8EDB] mb-3">
            Deployment
          </p>
          <h2 className="text-[clamp(1.75rem,4vw,3rem)] font-bold tracking-[-0.03em] leading-tight mb-4">
            Your infrastructure, your rules
          </h2>
          <p className="text-sm text-text-secondary max-w-lg mx-auto leading-relaxed">
            Choose the deployment model that fits your organization. Fully managed or on your own infrastructure.
          </p>
        </Reveal>

        <div className="grid md:grid-cols-2 gap-6">
          <Reveal delay={0}>
            <ShimmerCard className="h-full">
              <div className="relative bg-card border border-primary/30 rounded-xl p-8 flex flex-col h-full hover:border-primary/50 transition-all duration-300 shadow-[0_0_40px_rgba(102,122,204,0.1)] overflow-hidden">
                <div className="absolute inset-0 bg-gradient-to-br from-primary/[0.06] via-transparent to-transparent pointer-events-none" />
                <div className="relative">
                  <div className="flex items-center gap-3 mb-4">
                    <div className="w-10 h-10 rounded-lg bg-primary/10 border border-primary/20 flex items-center justify-center">
                      <svg className="w-5 h-5 text-primary" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={1.5}>
                        <path strokeLinecap="round" strokeLinejoin="round" d="M2.25 15a4.5 4.5 0 0 0 4.5 4.5H18a3.75 3.75 0 0 0 1.332-7.257 3 3 0 0 0-3.758-3.848 5.25 5.25 0 0 0-10.233 2.33A4.502 4.502 0 0 0 2.25 15z" />
                      </svg>
                    </div>
                    <span className="px-2 py-0.5 text-2xs font-mono font-semibold uppercase tracking-[0.1em] bg-accent-green/10 text-accent-green border border-accent-green/20 rounded-full">
                      Recommended
                    </span>
                  </div>

                  <h3 className="text-lg font-bold mb-2">Managed Cloud</h3>
                  <p className="text-sm text-text-secondary leading-relaxed mb-6">
                    We handle the infrastructure. Dedicated servers, automatic updates, and guaranteed uptime.
                  </p>

                  <ul className="space-y-2.5">
                    {[
                      "Dedicated servers for your organization",
                      "Automatic updates and patches",
                      "99.9% uptime SLA",
                      "Daily backups with point-in-time recovery",
                      "Global edge network for low latency",
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
              </div>
            </ShimmerCard>
          </Reveal>

          <Reveal delay={0.1}>
            <ShimmerCard className="h-full">
              <div className="bg-card border border-border rounded-xl p-8 flex flex-col h-full hover:border-border-light transition-colors duration-300">
                <div className="flex items-center gap-3 mb-4">
                  <div className="w-10 h-10 rounded-lg bg-white/[0.04] border border-border flex items-center justify-center">
                    <svg className="w-5 h-5 text-text-secondary" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={1.5}>
                      <path strokeLinecap="round" strokeLinejoin="round" d="M21.75 17.25v-.228a4.5 4.5 0 0 0-.12-1.03l-2.268-9.64a3.375 3.375 0 0 0-3.285-2.602H7.923a3.375 3.375 0 0 0-3.285 2.602l-2.268 9.64a4.5 4.5 0 0 0-.12 1.03v.228m19.5 0a3 3 0 0 1-3 3H5.25a3 3 0 0 1-3-3m19.5 0a3 3 0 0 0-3-3H5.25a3 3 0 0 0-3 3m16.5 0h.008v.008h-.008v-.008Zm-3 0h.008v.008h-.008v-.008Z" />
                    </svg>
                  </div>
                </div>

                <h3 className="text-lg font-bold mb-2">On-Premises</h3>
                <p className="text-sm text-text-secondary leading-relaxed mb-6">
                  Run ShellHub on your own infrastructure. Full data sovereignty and compliance control.
                </p>

                <ul className="space-y-2.5">
                  {[
                    "Complete data sovereignty",
                    "Deploy on Kubernetes with Helm charts",
                    "Docker Compose for simpler setups",
                    "Air-gapped environment support",
                    "Custom integration with your toolchain",
                  ].map((item) => (
                    <li key={item} className="flex items-center gap-2.5 text-sm text-text-secondary">
                      <svg className="w-4 h-4 text-text-muted shrink-0" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
                        <path strokeLinecap="round" strokeLinejoin="round" d="m4.5 12.75 6 6 9-13.5" />
                      </svg>
                      {item}
                    </li>
                  ))}
                </ul>
              </div>
            </ShimmerCard>
          </Reveal>
        </div>
      </div>
    </section>
  );
}
