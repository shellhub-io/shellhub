import { Link } from "react-router-dom";
import { Button } from "@shellhub/design-system/primitives";
import { ArrowRight } from "@/components/ArrowRight";
import { Section, SectionHeader } from "@/components/marketing";
import { Reveal, ConnectionGrid } from "../landing/components";

export function EnterpriseCTA() {
  return (
    <Section>
      <Reveal>
        <div className="relative bg-card border border-border rounded-2xl p-12 text-center overflow-hidden">
          <ConnectionGrid />
          <div className="absolute inset-0 bg-gradient-to-br from-primary/[0.06] via-transparent to-accent-cyan/[0.04] pointer-events-none" />

          <div className="relative z-10">
            <SectionHeader
              variant="cta"
              eyebrow="Ready to get started?"
              title="Talk to our team"
              subtitle="Get a demo, discuss your requirements, and find the right plan for your organization. Our team typically responds within one business day."
            />

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
    </Section>
  );
}
