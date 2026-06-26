import type { ReactElement, ReactNode } from "react";
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

function ErrorIcon({ className }: { className?: string }) {
  return (
    <svg
      xmlns="http://www.w3.org/2000/svg"
      fill="none"
      viewBox="0 0 24 24"
      strokeWidth={2}
      stroke="currentColor"
      className={className}
      aria-hidden="true"
    >
      <path
        strokeLinecap="round"
        strokeLinejoin="round"
        d="M12 9v3.75m9-.75a9 9 0 1 1-18 0 9 9 0 0 1 18 0Zm-9 3.75h.008v.008H12v-.008Z"
      />
    </svg>
  );
}

function SuccessIcon({ className }: { className?: string }) {
  return (
    <svg
      xmlns="http://www.w3.org/2000/svg"
      fill="none"
      viewBox="0 0 24 24"
      strokeWidth={2}
      stroke="currentColor"
      className={className}
      aria-hidden="true"
    >
      <path
        strokeLinecap="round"
        strokeLinejoin="round"
        d="M9 12.75 11.25 15 15 9.75M21 12a9 9 0 1 1-18 0 9 9 0 0 1 18 0Z"
      />
    </svg>
  );
}

function WarningIcon({ className }: { className?: string }) {
  return (
    <svg
      xmlns="http://www.w3.org/2000/svg"
      fill="none"
      viewBox="0 0 24 24"
      strokeWidth={2}
      stroke="currentColor"
      className={className}
      aria-hidden="true"
    >
      <path
        strokeLinecap="round"
        strokeLinejoin="round"
        d="M12 9v3.75m-9.303 3.376c-.866 1.5.217 3.374 1.948 3.374h14.71c1.73 0 2.813-1.874 1.948-3.374L13.949 3.378c-.866-1.5-3.032-1.5-3.898 0L2.697 16.126ZM12 15.75h.007v.008H12v-.008Z"
      />
    </svg>
  );
}

function InfoIcon({ className }: { className?: string }) {
  return (
    <svg
      xmlns="http://www.w3.org/2000/svg"
      fill="none"
      viewBox="0 0 24 24"
      strokeWidth={2}
      stroke="currentColor"
      className={className}
      aria-hidden="true"
    >
      <path
        strokeLinecap="round"
        strokeLinejoin="round"
        d="m11.25 11.25.041-.02a.75.75 0 0 1 1.063.852l-.708 2.836a.75.75 0 0 0 1.063.853l.041-.021M21 12a9 9 0 1 1-18 0 9 9 0 0 1 18 0Zm-9-3.75h.008v.008H12V8.25Z"
      />
    </svg>
  );
}

function SparkleIcon({ className }: { className?: string }) {
  return (
    <svg
      xmlns="http://www.w3.org/2000/svg"
      fill="none"
      viewBox="0 0 24 24"
      strokeWidth={1.5}
      stroke="currentColor"
      className={className}
      aria-hidden="true"
    >
      <path
        strokeLinecap="round"
        strokeLinejoin="round"
        d="M9.813 15.904 9 18.75l-.813-2.846a4.5 4.5 0 0 0-3.09-3.09L2.25 12l2.846-.813a4.5 4.5 0 0 0 3.09-3.09L9 5.25l.813 2.846a4.5 4.5 0 0 0 3.09 3.09L15.75 12l-2.846.813a4.5 4.5 0 0 0-3.09 3.09Z"
      />
      <path
        strokeLinecap="round"
        strokeLinejoin="round"
        d="M18 2.25l.259 1.035a3.375 3.375 0 0 0 2.456 2.456L21.75 6l-1.035.259a3.375 3.375 0 0 0-2.456 2.456L18 9.75l-.259-1.035a3.375 3.375 0 0 0-2.456-2.456L14.25 6l1.035-.259a3.375 3.375 0 0 0 2.456-2.456L18 2.25Z"
      />
    </svg>
  );
}

function ArrowIcon({ className }: { className?: string }) {
  return (
    <svg
      xmlns="http://www.w3.org/2000/svg"
      fill="none"
      viewBox="0 0 24 24"
      strokeWidth={2}
      stroke="currentColor"
      className={className}
      aria-hidden="true"
    >
      <path strokeLinecap="round" strokeLinejoin="round" d="M5 12h14" />
      <path strokeLinecap="round" strokeLinejoin="round" d="m12 5 7 7-7 7" />
    </svg>
  );
}

function XMarkIcon({ className }: { className?: string }) {
  return (
    <svg
      xmlns="http://www.w3.org/2000/svg"
      fill="none"
      viewBox="0 0 24 24"
      strokeWidth={2}
      stroke="currentColor"
      className={className}
      aria-hidden="true"
    >
      <path
        strokeLinecap="round"
        strokeLinejoin="round"
        d="M6 18 18 6M6 6l12 12"
      />
    </svg>
  );
}

type SemanticConfig = {
  bg: string;
  border: string;
  text: string;
  Icon: (props: { className?: string }) => ReactElement;
  role: "alert" | "status";
  ariaLive: "assertive" | "polite";
};

const SEMANTIC: Record<SemanticVariant, SemanticConfig> = {
  error: {
    bg: "bg-accent-red/8",
    border: "border-accent-red/20",
    text: "text-accent-red",
    Icon: ErrorIcon,
    role: "alert",
    ariaLive: "assertive",
  },
  success: {
    bg: "bg-accent-green/8",
    border: "border-accent-green/20",
    text: "text-accent-green",
    Icon: SuccessIcon,
    role: "status",
    ariaLive: "polite",
  },
  warning: {
    bg: "bg-accent-yellow/8",
    border: "border-accent-yellow/20",
    text: "text-accent-yellow",
    Icon: WarningIcon,
    role: "alert",
    ariaLive: "assertive",
  },
  info: {
    bg: "bg-accent-blue/8",
    border: "border-accent-blue/20",
    text: "text-accent-blue",
    Icon: InfoIcon,
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
          <SparkleIcon className="w-4 h-4 text-primary shrink-0" />
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
            <ArrowIcon className="w-3 h-3 transition-transform duration-150 group-hover/cta:translate-x-0.5" />
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
      <Icon className="w-3.5 h-3.5 shrink-0" />
      <span className="flex-1 min-w-0">{children}</span>
      {onDismiss && (
        <IconButton
          size="sm"
          aria-label="Dismiss"
          onClick={onDismiss}
          className="ml-1 shrink-0 -mr-0.5"
        >
          <XMarkIcon className="w-3 h-3" />
        </IconButton>
      )}
    </div>
  );
}
