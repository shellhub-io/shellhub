import { Button } from "@shellhub/design-system/primitives";
import { Section } from "@/components/marketing";
import { ArrowRight } from "@/components/ArrowRight";
import { Reveal, ConnectionGrid } from "./components";
import { signupUrl } from "@/links";

export function CTA() {
  return (
    <Section
      container={false}
      bordered={false}
      className="text-center grid-bg relative overflow-hidden"
    >
      <ConnectionGrid />
      <div className="absolute inset-0 bg-gradient-radial from-primary/8 via-transparent to-transparent pointer-events-none" />
      <div className="absolute top-1/2 left-1/3 -translate-y-1/2 w-80 h-80 bg-primary/6 rounded-full blur-3xl pointer-events-none" />
      <div className="absolute bottom-0 right-1/4 w-64 h-64 bg-accent-cyan/5 rounded-full blur-3xl pointer-events-none" />
      <div className="max-w-7xl mx-auto px-8 relative z-10">
        <Reveal>
          <h2 className="text-[clamp(1.75rem,4vw,3rem)] font-bold tracking-[-0.03em] leading-tight mb-4">
            Ready to connect to your devices?
          </h2>
        </Reveal>
        <Reveal>
          <p className="text-sm text-text-secondary max-w-md mx-auto mb-10 leading-relaxed">
            Get started with ShellHub in minutes. Free forever for small teams.
          </p>
        </Reveal>
        <Reveal>
          <Button
            as="a"
            href={signupUrl}
            target="_blank"
            rel="noopener noreferrer"
            variant="primary"
            size="xl"
            glow
            iconRight={<ArrowRight />}
          >
            Get Started Free
          </Button>
        </Reveal>
      </div>
    </Section>
  );
}
