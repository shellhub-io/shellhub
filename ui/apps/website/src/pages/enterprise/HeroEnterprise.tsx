import { Badge } from "@shellhub/design-system/primitives";
import {
  ConnectionGrid,
  GlowOrbs,
  Reveal,
} from "@shellhub/design-system/components";
import { ActionButtonGroup } from "@/components";

export function HeroEnterprise() {
  return (
    <section className="relative pt-32 pb-24 overflow-hidden">
      <ConnectionGrid />
      <GlowOrbs preset="section" tone="primary" />

      <div className="max-w-7xl mx-auto px-8 relative z-10 text-center">
        <Reveal>
          <Badge shape="pill" color="yellow" className="mb-6">
            Enterprise
          </Badge>
        </Reveal>
        <Reveal>
          <h1 className="text-[clamp(2rem,5vw,3.5rem)] font-bold tracking-[-0.03em] leading-[1.1] mb-6 max-w-3xl mx-auto">
            Device management built for{" "}
            <span className="bg-gradient-to-r from-primary via-accent-blue to-accent-cyan bg-clip-text text-transparent">
              scale and control
            </span>
          </h1>
        </Reveal>
        <Reveal>
          <p className="text-sm text-text-secondary max-w-lg mx-auto leading-relaxed mb-10">
            SSO, admin panel, audit logs, and dedicated support. Everything your
            team needs to manage thousands of devices securely.
          </p>
        </Reveal>
        <Reveal>
          <ActionButtonGroup
            primaryAction={{ label: "Get a Quote", to: "/pricing" }}
            secondaryAction={{ label: "Compare Plans", to: "/pricing" }}
          />
        </Reveal>
      </div>
    </section>
  );
}
