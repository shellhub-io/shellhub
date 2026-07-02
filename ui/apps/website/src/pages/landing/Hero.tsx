import { Link } from "react-router-dom";
import { Button, ShellHubCloudIcon } from "@shellhub/design-system/primitives";
import { GlowOrbs } from "@shellhub/design-system/components";
import { ArrowRight } from "@/components/ArrowRight";
import { ConnectionGrid } from "./components";

export function Hero() {
  return (
    <section className="min-h-screen flex flex-col items-center justify-center text-center px-6 pt-28 pb-20 relative overflow-hidden grid-bg">
      <ConnectionGrid />
      <GlowOrbs preset="hero" />

      <div className="relative z-10 max-w-4xl flex flex-col items-center">
        {/* Floating ShellHub cloud */}
        <div className="animate-float mb-8 inline-block">
          <ShellHubCloudIcon className="h-16 drop-shadow-[0_0_24px_rgba(102,122,204,0.35)]" />
        </div>

        {/* Badge like app's active nav pill */}
        <div className="inline-flex items-center gap-2 px-3 py-1.5 rounded-full bg-primary/[0.06] border border-primary/20 mb-8 animate-fade-in">
          <span className="relative flex h-2 w-2">
            <span className="animate-ping absolute inline-flex h-full w-full rounded-full bg-primary opacity-75" />
            <span className="relative inline-flex rounded-full h-2 w-2 bg-primary" />
          </span>
          <span className="text-2xs font-mono font-semibold uppercase tracking-label text-primary">
            Open Source SSH Gateway
          </span>
        </div>

        <h1
          className="text-[clamp(2.5rem,6.5vw,5rem)] font-bold leading-[1.08] tracking-[-0.035em] mb-6 animate-fade-in"
          style={{ animationDelay: "100ms" }}
        >
          All your devices. <span className="text-primary">One gateway.</span>
        </h1>

        <p
          className="text-base sm:text-lg text-text-secondary max-w-xl mx-auto mb-10 leading-relaxed animate-fade-in"
          style={{ animationDelay: "200ms" }}
        >
          ShellHub is a centralized SSH gateway for remote access to Linux
          devices. No public IPs, no VPNs, no firewall changes.
        </p>

        <div
          className="flex gap-3 flex-wrap justify-center animate-fade-in"
          style={{ animationDelay: "300ms" }}
        >
          <Button
            as={Link}
            to="/getting-started"
            variant="primary"
            size="xl"
            glow
            iconRight={<ArrowRight />}
          >
            Get Started
          </Button>
        </div>
      </div>
    </section>
  );
}
