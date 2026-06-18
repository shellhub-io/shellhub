import { ReactNode, useId } from "react";
import { IconBadge } from "@shellhub/design-system/primitives";

export type EmptyStateAccent = "primary" | "yellow";

export interface EmptyStateFeature {
  /** Sized but uncolored heroicon, e.g. `<LinkIcon className="w-5 h-5" />`. */
  icon: ReactNode;
  title: string;
  description: string;
}

export interface EmptyStateProps {
  /** Sized but uncolored heroicon, e.g. `<GlobeAltIcon className="w-8 h-8" />`. */
  icon: ReactNode;
  overline: string;
  title: string;
  description: string;
  accent?: EmptyStateAccent;
  features?: EmptyStateFeature[];
  /** Small muted text rendered under the call-to-action. */
  footnote?: ReactNode;
  /** Call-to-action slot — button(s), links, RestrictedAction, etc. */
  children?: ReactNode;
}

interface AccentStyles {
  badge: string;
  icon: string;
  overline: string;
  orbPrimary: string;
  orbSecondary: string;
}

/**
 * Accent styles. Full literal class strings (never interpolated fragments) so
 * the Tailwind JIT keeps them. The hero icon inherits the badge's text color
 * via `currentColor`; feature-card icons stay primary-accented in all variants.
 * Typed as `Record<EmptyStateAccent, …>` so adding an accent without a matching
 * entry is a compile error rather than a runtime `undefined`.
 */
const ACCENT = {
  primary: {
    badge: "bg-primary/10 border-primary/20 shadow-primary/5",
    icon: "text-primary",
    overline: "text-primary/80",
    orbPrimary: "bg-primary/5",
    orbSecondary: "bg-accent-blue/5",
  },
  yellow: {
    badge: "bg-accent-yellow/10 border-accent-yellow/20 shadow-accent-yellow/5",
    icon: "text-accent-yellow",
    overline: "text-accent-yellow/80",
    orbPrimary: "bg-accent-yellow/5",
    orbSecondary: "bg-primary/5",
  },
} satisfies Record<EmptyStateAccent, AccentStyles>;

/**
 * Full-page onboarding / empty / gated-feature splash: a centered card over a
 * full-bleed decorative background. Owns the full-bleed layout so call sites
 * only declare content. Render it as the sole content of a page (inside the
 * AppLayout/AdminLayout `<main>`).
 */
export default function EmptyState({
  icon,
  overline,
  title,
  description,
  accent = "primary",
  features,
  footnote,
  children,
}: EmptyStateProps) {
  const headingId = useId();
  const styles = ACCENT[accent];
  const hasFeatures = !!features?.length;

  return (
    <section
      aria-labelledby={headingId}
      className="relative min-h-full flex items-center justify-center"
    >
      {/* Decorative background — bleeds past the main padding (p-8 pb-4) */}
      <div
        aria-hidden="true"
        className="absolute inset-0 overflow-hidden pointer-events-none -mx-8 -mt-8 -mb-4"
      >
        <div
          className={`absolute -top-32 left-1/3 w-[500px] h-[500px] rounded-full blur-[120px] animate-pulse-subtle ${styles.orbPrimary}`}
        />
        <div
          className={`absolute bottom-0 right-1/4 w-[400px] h-[400px] rounded-full blur-[100px] animate-pulse-subtle ${styles.orbSecondary}`}
          style={{ animationDelay: "1s" }}
        />
        <div className="absolute inset-0 grid-bg opacity-30" />
      </div>

      <div className="w-full max-w-3xl px-4 py-6 animate-fade-in">
        {/* Header */}
        <div className="text-center mb-10">
          <div
            aria-hidden="true"
            className={`w-16 h-16 rounded-2xl border flex items-center justify-center mx-auto mb-6 shadow-lg ${styles.badge} ${styles.icon}`}
          >
            {icon}
          </div>
          <span
            className={`inline-block text-2xs font-mono font-semibold uppercase tracking-wide mb-2 ${styles.overline}`}
          >
            {overline}
          </span>
          <h1
            id={headingId}
            className="text-3xl font-bold text-text-primary mb-3"
          >
            {title}
          </h1>
          <p className="text-sm text-text-muted max-w-md mx-auto leading-relaxed">
            {description}
          </p>
        </div>

        {/* Feature highlights */}
        {features?.length ? (
          <ul
            className="grid grid-cols-1 md:grid-cols-3 gap-4 mb-10"
          >
            {features.map((feature, idx) => (
              <li
                key={feature.title}
                className="bg-card/60 border border-border rounded-xl p-5 text-center animate-slide-up"
                style={{ animationDelay: `${150 + idx * 100}ms` }}
              >
                <IconBadge
                  aria-hidden="true"
                  size="md"
                  color="primary"
                  className="mx-auto mb-3"
                >
                  {feature.icon}
                </IconBadge>
                <h2 className="text-sm font-semibold text-text-primary mb-1">
                  {feature.title}
                </h2>
                <p className="text-xs text-text-muted leading-relaxed text-balance">
                  {feature.description}
                </p>
              </li>
            ))}
          </ul>
        ) : null}

        {/* Call to action */}
        <div
          className="text-center animate-slide-up"
          style={{ animationDelay: hasFeatures ? "450ms" : "200ms" }}
        >
          {children}
          {footnote != null ? (
            <p className="mt-4 text-2xs text-text-muted">{footnote}</p>
          ) : null}
        </div>
      </div>
    </section>
  );
}
