import {
  ConnectionGrid,
  GlowOrbs,
  Reveal,
} from "@shellhub/design-system/components";
import { ActionButton, Section } from "@/components";
import { signupUrl } from "@/links";

export function CTA() {
  return (
    <Section
      container={false}
      bordered={false}
      className="text-center grid-bg relative overflow-hidden"
    >
      <ConnectionGrid />
      <GlowOrbs preset="duo" tone="primary" />
      <div className="max-w-7xl mx-auto px-8 relative z-raised">
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
          <ActionButton
            action={{
              label: "Get Started Free",
              href: signupUrl,
              external: true,
            }}
          />
        </Reveal>
      </div>
    </Section>
  );
}
