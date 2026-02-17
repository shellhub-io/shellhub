import { ReactNode } from "react";
import {
  LockClosedIcon,
  ArrowTopRightOnSquareIcon,
} from "@heroicons/react/24/outline";
import { getConfig } from "../../env";

interface Highlight {
  icon: ReactNode;
  title: string;
  description: string;
}

interface FeatureGateProps {
  children: ReactNode;
  feature: string;
  description: string;
  highlights?: Highlight[];
}

export default function FeatureGate({
  children,
  feature,
  description,
  highlights,
}: FeatureGateProps) {
  if (getConfig().cloud || getConfig().enterprise) {
    return <>{children}</>;
  }

  return (
    <div className="relative -mx-8 -mt-8 min-h-[calc(100vh-3.5rem)] flex flex-col">
      {/* Background */}
      <div className="absolute inset-0 overflow-hidden pointer-events-none">
        <div className="absolute -top-32 left-1/3 w-[500px] h-[500px] bg-accent-yellow/5 rounded-full blur-[120px] animate-pulse-subtle" />
        <div
          className="absolute bottom-0 right-1/4 w-[400px] h-[400px] bg-primary/5 rounded-full blur-[100px] animate-pulse-subtle"
          style={{ animationDelay: "1s" }}
        />
        <div className="absolute inset-0 grid-bg opacity-30" />
      </div>

      <div className="relative z-10 flex-1 flex items-center justify-center px-8 py-12">
        <div className="w-full max-w-2xl animate-fade-in">
          {/* Header */}
          <div className="text-center mb-10">
            <div className="w-16 h-16 rounded-2xl bg-accent-yellow/10 border border-accent-yellow/20 flex items-center justify-center mx-auto mb-6 shadow-lg shadow-accent-yellow/5">
              <LockClosedIcon className="w-8 h-8 text-accent-yellow" />
            </div>

            <span className="inline-block text-2xs font-mono font-semibold uppercase tracking-wide text-accent-yellow/80 mb-2">
              Premium Feature
            </span>
            <h1 className="text-3xl font-bold text-text-primary mb-3">
              {feature}
            </h1>
            <p className="text-sm text-text-muted max-w-md mx-auto leading-relaxed">
              {description}
            </p>
          </div>

          {/* Highlights */}
          {highlights && highlights.length > 0 && (
            <div className="grid grid-cols-1 md:grid-cols-3 gap-4 mb-10">
              {highlights.map((h, idx) => (
                <div
                  key={h.title}
                  className="bg-card/60 border border-border rounded-xl p-5 text-center animate-slide-up"
                  style={{ animationDelay: `${150 + idx * 100}ms` }}
                >
                  <div className="w-10 h-10 rounded-lg bg-primary/10 border border-primary/20 flex items-center justify-center mx-auto mb-3 text-primary">
                    {h.icon}
                  </div>
                  <h3 className="text-sm font-semibold text-text-primary mb-1">
                    {h.title}
                  </h3>
                  <p className="text-xs text-text-muted leading-relaxed">
                    {h.description}
                  </p>
                </div>
              ))}
            </div>
          )}

          {/* CTA */}
          <div
            className="text-center animate-slide-up"
            style={{ animationDelay: `${highlights ? 450 : 200}ms` }}
          >
            <a
              href="https://www.shellhub.io/pricing"
              target="_blank"
              rel="noopener noreferrer"
              className="inline-flex items-center gap-2 px-6 py-3 bg-primary hover:bg-primary-600 text-white rounded-lg text-sm font-semibold transition-all duration-200 shadow-lg shadow-primary/20"
            >
              Pricing
              <ArrowTopRightOnSquareIcon className="w-4 h-4" strokeWidth={2} />
            </a>
            <p className="mt-4 text-2xs text-text-muted">
              Available on ShellHub Cloud and Enterprise editions.
            </p>
          </div>
        </div>
      </div>
    </div>
  );
}
