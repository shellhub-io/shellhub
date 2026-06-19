import { ReactNode } from "react";
import {
  ExclamationCircleIcon,
  CheckCircleIcon,
  ExclamationTriangleIcon,
  InformationCircleIcon,
  XMarkIcon,
} from "@heroicons/react/24/outline";
import { IconButton } from "@shellhub/design-system/primitives";

export type AlertVariant = "error" | "success" | "warning" | "info";

interface AlertProps {
  /** Visual and semantic variant. Controls color, icon, role, and aria-live. */
  variant: AlertVariant;

  /** Alert body — plain string or rich JSX. */
  children: ReactNode;

  /** When provided, renders a dismiss button (XMark). The parent owns visibility state. */
  onDismiss?: () => void;

  /** Layout overrides only — margins, display mode. Don't override color or typography. */
  className?: string;
}

type IconComponent = (props: {
  className?: string;
  strokeWidth?: number;
}) => React.ReactElement | null;

type VariantConfig = {
  bg: string;
  border: string;
  text: string;
  Icon: IconComponent;
  role: "alert" | "status";
  ariaLive: "assertive" | "polite";
};

const VARIANT: Record<AlertVariant, VariantConfig> = {
  error: {
    bg: "bg-accent-red/8",
    border: "border-accent-red/20",
    text: "text-accent-red",
    Icon: ExclamationCircleIcon as IconComponent,
    role: "alert",
    ariaLive: "assertive",
  },
  success: {
    bg: "bg-accent-green/8",
    border: "border-accent-green/20",
    text: "text-accent-green",
    Icon: CheckCircleIcon as IconComponent,
    role: "status",
    ariaLive: "polite",
  },
  warning: {
    bg: "bg-accent-yellow/8",
    border: "border-accent-yellow/20",
    text: "text-accent-yellow",
    Icon: ExclamationTriangleIcon as IconComponent,
    role: "alert",
    ariaLive: "assertive",
  },
  info: {
    bg: "bg-accent-blue/8",
    border: "border-accent-blue/20",
    text: "text-accent-blue",
    Icon: InformationCircleIcon as IconComponent,
    role: "status",
    ariaLive: "polite",
  },
};

export default function Alert({
  variant,
  children,
  onDismiss,
  className,
}: AlertProps) {
  const { bg, border, text, Icon, role, ariaLive } = VARIANT[variant];

  return (
    <div
      role={role}
      aria-live={ariaLive}
      className={[
        "flex items-center gap-2 border px-3.5 py-2.5 rounded-md text-xs font-mono animate-slide-down",
        bg,
        border,
        text,
        className,
      ]
        .filter(Boolean)
        .join(" ")}
    >
      <Icon className="w-3.5 h-3.5 shrink-0" strokeWidth={2} />
      <span className="flex-1 min-w-0">{children}</span>
      {onDismiss && (
        <IconButton
          size="sm"
          aria-label="Dismiss alert"
          onClick={onDismiss}
          className="ml-1 shrink-0 -mr-0.5"
        >
          <XMarkIcon className="w-3 h-3" />
        </IconButton>
      )}
    </div>
  );
}
