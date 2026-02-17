import { Reveal } from "../landing/components";

export function HeroPricing() {
  return (
    <section className="relative pt-32 pb-16">
      <div className="absolute top-1/3 left-1/2 -translate-x-1/2 w-[600px] h-[600px] bg-primary/5 rounded-full blur-3xl pointer-events-none" />

      <div className="max-w-7xl mx-auto px-8 relative z-10 text-center">
        <Reveal>
          <h1 className="text-[clamp(2rem,5vw,3.5rem)] font-bold tracking-[-0.03em] leading-[1.1] mb-4">
            Simple, transparent pricing
          </h1>
        </Reveal>
        <Reveal>
          <p className="text-sm text-text-secondary max-w-md mx-auto leading-relaxed">
            Start free with the Community edition. Scale up as your team and fleet grow.
          </p>
        </Reveal>
      </div>
    </section>
  );
}
