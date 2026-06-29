import type { ComponentType, SVGProps, ReactNode } from "react";
import {
  ExclamationCircleIcon,
  CheckCircleIcon,
  ExclamationTriangleIcon,
  InformationCircleIcon,
  SparklesIcon,
  ArrowRightIcon,
  XMarkIcon,
} from "@heroicons/react/24/outline";
import { IconButton } from "./IconButton";
import { cn } from "./cn";

// ---------------------------------------------------------------------------
// Public types
// ---------------------------------------------------------------------------

export type CalloutVariant =
  | "error"
  | "success"
  | "warning"
  | "info"
  | "feature";

/** The semantic (non-feature) variants, which render through the SEMANTIC map. */
type SemanticVariant = Exclude<CalloutVariant, "feature">;

type SemanticConfig = {
  bg: string;
  border: string;
  text: string;
  Icon: ComponentType<SVGProps<SVGSVGElement>>;
  role: "alert" | "status";
  ariaLive: "assertive" | "polite";
};

const SEMANTIC: Record<SemanticVariant, SemanticConfig> = {
  error: {
    bg: "bg-accent-red/8",
    border: "border-accent-red/20",
    text: "text-accent-red",
    Icon: ExclamationCircleIcon,
    role: "alert",
    ariaLive: "assertive",
  },
  success: {
    bg: "bg-accent-green/8",
    border: "border-accent-green/20",
    text: "text-accent-green",
    Icon: CheckCircleIcon,
    role: "status",
    ariaLive: "polite",
  },
  warning: {
    bg: "bg-accent-yellow/8",
    border: "border-accent-yellow/20",
    text: "text-accent-yellow",
    Icon: ExclamationTriangleIcon,
    role: "alert",
    ariaLive: "assertive",
  },
  info: {
    bg: "bg-accent-blue/8",
    border: "border-accent-blue/20",
    text: "text-accent-blue",
    Icon: InformationCircleIcon,
    role: "status",
    ariaLive: "polite",
  },
};

export interface CalloutProps {
  /** Visual and semantic variant. */
  variant: CalloutVariant;

  /** Callout body — plain string or rich JSX. */
  children: ReactNode;

  /** When provided, renders a dismiss button. The parent owns visibility state. */
  onDismiss?: () => void;

  /** CTA link, only meaningful for the `feature` variant. */
  action?: { href: string; label: string };

  /** Layout overrides only — margins, display mode. Don't override color or typography. */
  className?: string;
}

export function Callout({
  variant,
  children,
  onDismiss,
  action,
  className,
}: CalloutProps) {
  if (variant === "feature") {
    return (
      <div
        role="note"
        className={cn(
          "relative overflow-hidden flex items-center justify-between gap-4 px-4 py-3 rounded-lg border border-primary/15 bg-gradient-to-br from-primary/5 to-primary/[0.01] font-sans before:content-[''] before:absolute before:top-0 before:inset-x-0 before:h-px before:bg-gradient-to-r before:from-transparent before:via-primary/25 before:to-transparent",
          className,
        )}
      >
        <div className="flex items-center gap-2 min-w-0">
          <SparklesIcon className="w-4 h-4 text-primary shrink-0" />
          <span className="text-sm leading-snug text-text-secondary [&_strong]:text-text-primary [&_strong]:font-medium">
            {children}
          </span>
        </div>
        {action && (
          <a
            href={action.href}
            className="group/cta inline-flex items-center gap-1 text-xs text-primary no-underline whitespace-nowrap shrink-0 transition-colors duration-150 !border-b-0 hover:text-text-primary hover:!border-b-0"
          >
            {action.label}
            <ArrowRightIcon
              className="w-3 h-3 transition-transform duration-150 group-hover/cta:translate-x-0.5"
              strokeWidth={2}
            />
          </a>
        )}
      </div>
    );
  }

  const { bg, border, text, Icon, role, ariaLive } =
    SEMANTIC[variant as SemanticVariant];

  return (
    <div
      role={role}
      aria-live={ariaLive}
      className={cn(
        "flex items-center gap-2 border px-3.5 py-2.5 rounded-md text-xs font-mono animate-slide-down",
        bg,
        border,
        text,
        className,
      )}
    >
      <Icon className="w-3.5 h-3.5 shrink-0" strokeWidth={2} />
      <span className="flex-1 min-w-0">{children}</span>
      {onDismiss && (
        <IconButton
          size="sm"
          aria-label="Dismiss"
          onClick={onDismiss}
          className="ml-1 shrink-0 -mr-0.5"
        >
          <XMarkIcon className="w-3 h-3" strokeWidth={2} />
        </IconButton>
      )}
    </div>
  );
}
