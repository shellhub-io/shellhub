import { Link } from "react-router-dom";
import { Button } from "@shellhub/design-system/primitives";
import { ArrowRight } from "@/components/ArrowRight";
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
                Get a demo, discuss your requirements, and find the right plan
                for your organization. Our team typically responds within one
                business day.
              </p>

              <div className="flex flex-col sm:flex-row items-center justify-center gap-3">
                <Button
                  as="a"
                  href="mailto:sales@shellhub.io"
                  variant="primary"
                  size="xl"
                  glow
                  iconRight={<ArrowRight />}
                >
                  Contact Sales
                </Button>
                <Button as={Link} to="/pricing" variant="outline" size="xl">
                  View Pricing
                </Button>
              </div>
            </div>
          </div>
        </Reveal>
      </div>
    </section>
  );
}
