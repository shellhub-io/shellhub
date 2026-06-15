import { Link } from "react-router-dom";
import { Badge, Button } from "@shellhub/design-system/primitives";
import { ArrowRight } from "@/components/ArrowRight";
import { Reveal, ConnectionGrid } from "../landing/components";

export function HeroEnterprise() {
  return (
    <section className="relative pt-32 pb-24 overflow-hidden">
      <ConnectionGrid />
      <div className="absolute inset-0 bg-gradient-radial from-primary/8 via-transparent to-transparent pointer-events-none" />
      <div className="absolute top-1/3 right-1/4 w-96 h-96 bg-primary/6 rounded-full blur-3xl pointer-events-none" />

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
          <div className="flex flex-col sm:flex-row items-center justify-center gap-3">
            <Button
              as={Link}
              to="/pricing"
              variant="primary"
              size="xl"
              glow
              iconRight={<ArrowRight />}
            >
              Get a Quote
            </Button>
            <Button as={Link} to="/pricing" variant="outline" size="xl">
              Compare Plans
            </Button>
          </div>
        </Reveal>
      </div>
    </section>
  );
}
