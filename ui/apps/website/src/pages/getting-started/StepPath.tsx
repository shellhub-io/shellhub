import { Link } from "react-router-dom";
import {
  ArrowRightIcon,
  BuildingOffice2Icon,
  ServerIcon,
} from "@heroicons/react/24/outline";
import {
  Badge,
  Button,
  Card,
  IconBadge,
} from "@shellhub/design-system/primitives";
import { FeatureListItem } from "@/components/marketing/FeatureListItem";
import { Reveal, ShimmerCard } from "../landing/components";

interface StepPathProps {
  onSelectCloud: () => void;
  onSelectSelfHosted: () => void;
}

export function StepPath({ onSelectCloud, onSelectSelfHosted }: StepPathProps) {
  return (
    <div className="flex flex-col gap-6 max-w-3xl mx-auto w-full">
      <div className="grid md:grid-cols-2 gap-6">
        {/* Cloud card */}
        <Reveal delay={0}>
          <ShimmerCard className="h-full">
            <div className="relative bg-card border border-primary/30 rounded-xl p-8 flex flex-col h-full hover:border-primary/50 transition-all duration-300 shadow-[0_0_40px_rgba(102,122,204,0.15)] overflow-hidden">
              <div className="absolute inset-0 bg-gradient-to-br from-primary/[0.08] via-primary/[0.02] to-transparent pointer-events-none" />
              <div className="absolute top-0 right-0 w-40 h-40 bg-primary/[0.08] rounded-full -translate-y-1/2 translate-x-1/2 blur-3xl pointer-events-none" />
              <div className="relative flex items-center gap-3 mb-4">
                <IconBadge
                  color="primary"
                  className="shadow-[0_0_12px_rgba(102,122,204,0.15)]"
                >
                  <img src="/cloud-icon.svg" alt="" className="h-5" />
                </IconBadge>
                <Badge shape="pill" color="green">
                  Recommended
                </Badge>
              </div>

              <h3 className="text-lg font-bold mb-2">ShellHub Cloud</h3>
              <p className="text-sm text-text-secondary leading-relaxed mb-6">
                Start in seconds. No infrastructure to manage.
              </p>

              <ul className="space-y-2.5 mb-8 flex-1">
                {[
                  "Free tier available",
                  "Managed updates",
                  "No maintenance",
                ].map((item) => (
                  <FeatureListItem key={item} color="green">
                    {item}
                  </FeatureListItem>
                ))}
              </ul>

              <Button
                variant="primary"
                size="lg"
                glow
                fullWidth
                onClick={onSelectCloud}
              >
                Sign Up Free
              </Button>
            </div>
          </ShimmerCard>
        </Reveal>

        {/* Self-hosted card */}
        <Reveal delay={0.1}>
          <ShimmerCard className="h-full">
            <div className="bg-card/60 border border-border rounded-xl p-8 flex flex-col h-full hover:border-border-light transition-colors duration-300">
              <div className="flex items-center gap-3 mb-4">
                <div className="w-10 h-10 rounded-lg bg-white/[0.04] border border-border flex items-center justify-center">
                  <ServerIcon className="w-5 h-5 text-text-secondary" />
                </div>
                <span className="px-2 py-0.5 text-2xs font-mono font-semibold uppercase tracking-[0.1em] bg-white/[0.03] text-text-muted border border-border rounded-full">
                  Open Source
                </span>
              </div>

              <h3 className="text-lg font-bold mb-2">Self-hosted</h3>
              <p className="text-sm text-text-secondary leading-relaxed mb-6">
                Run on your own infrastructure. Full control.
              </p>

              <ul className="space-y-2.5 mb-8 flex-1">
                {["Open source", "Your data stays yours", "Docker Compose"].map(
                  (item) => (
                    <FeatureListItem key={item} color="muted">
                      {item}
                    </FeatureListItem>
                  ),
                )}
              </ul>

              <Button
                variant="surface"
                size="lg"
                fullWidth
                className="hover:scale-[1.02] active:scale-[0.98]"
                onClick={onSelectSelfHosted}
              >
                Continue
              </Button>
            </div>
          </ShimmerCard>
        </Reveal>
      </div>

      {/* Enterprise card — full width below */}
      <Reveal delay={0.2}>
        <Card className="p-6 flex flex-col sm:flex-row items-start sm:items-center gap-5 hover:border-primary/30 transition-colors duration-300">
          <IconBadge color="yellow">
            <BuildingOffice2Icon className="w-5 h-5 text-accent-yellow" />
          </IconBadge>
          <div className="flex-1 min-w-0">
            <h3 className="text-sm font-bold mb-1">Enterprise</h3>
            <p className="text-xs text-text-secondary leading-relaxed">
              Need SSO, audit logs, or dedicated support? Talk to our team about
              a plan that fits your organization.
            </p>
          </div>
          <Button
            as={Link}
            to="/enterprise"
            variant="outline"
            size="sm"
            className="px-4 py-2 text-xs gap-1.5 shrink-0"
            iconRight={
              <ArrowRightIcon
                className="w-3.5 h-3.5"
                strokeWidth={2.5}
                aria-hidden="true"
              />
            }
          >
            Contact Sales
          </Button>
        </Card>
      </Reveal>
    </div>
  );
}
