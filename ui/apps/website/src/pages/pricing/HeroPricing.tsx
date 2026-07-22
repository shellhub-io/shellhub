import { GlowOrbs, Reveal } from "@shellhub/design-system/components";

export function HeroPricing() {
  return (
    <section className="relative pt-32 pb-16 overflow-hidden">
      <GlowOrbs preset="section" tone="primary" />

      <div className="max-w-7xl mx-auto px-8 relative z-10 text-center">
        <Reveal>
          <h1 className="text-[clamp(2rem,5vw,3.5rem)] font-bold tracking-[-0.03em] leading-[1.1] mb-4">
            Simple, transparent pricing
          </h1>
        </Reveal>
        <Reveal>
          <p className="text-sm text-text-secondary max-w-md mx-auto leading-relaxed">
            Start free with the Community edition. Scale up as your team and
            fleet grow.
          </p>
        </Reveal>
      </div>
    </section>
  );
}
