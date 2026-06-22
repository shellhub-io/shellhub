import { Link } from "react-router-dom";
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
                  <svg
                    className="w-5 h-5 text-text-secondary"
                    fill="none"
                    viewBox="0 0 24 24"
                    stroke="currentColor"
                    strokeWidth={1.5}
                  >
                    <path
                      strokeLinecap="round"
                      strokeLinejoin="round"
                      d="M21.75 17.25v-.228a4.5 4.5 0 0 0-.12-1.03l-2.268-9.64a3.375 3.375 0 0 0-3.285-2.602H7.923a3.375 3.375 0 0 0-3.285 2.602l-2.268 9.64a4.5 4.5 0 0 0-.12 1.03v.228m19.5 0a3 3 0 0 1-3 3H5.25a3 3 0 0 1-3-3m19.5 0a3 3 0 0 0-3-3H5.25a3 3 0 0 0-3 3m16.5 0h.008v.008h-.008v-.008Zm-3 0h.008v.008h-.008v-.008Z"
                    />
                  </svg>
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
            <svg
              className="w-5 h-5 text-accent-yellow"
              fill="none"
              viewBox="0 0 24 24"
              stroke="currentColor"
              strokeWidth={1.5}
            >
              <path
                strokeLinecap="round"
                strokeLinejoin="round"
                d="M2.25 21h19.5m-18-18v18m10.5-18v18m6-13.5V21M6.75 6.75h.75m-.75 3h.75m-.75 3h.75m3-6h.75m-.75 3h.75m-.75 3h.75M6.75 21v-3.375c0-.621.504-1.125 1.125-1.125h2.25c.621 0 1.125.504 1.125 1.125V21M3 3h12m-.75 4.5H21m-3.75 3H21m-3.75 3H21"
              />
            </svg>
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
              <svg
                className="w-3.5 h-3.5"
                fill="none"
                viewBox="0 0 24 24"
                stroke="currentColor"
                strokeWidth={2.5}
              >
                <path
                  strokeLinecap="round"
                  strokeLinejoin="round"
                  d="M13.5 4.5 21 12m0 0-7.5 7.5M21 12H3"
                />
              </svg>
            }
          >
            Contact Sales
          </Button>
        </Card>
      </Reveal>
    </div>
  );
}
