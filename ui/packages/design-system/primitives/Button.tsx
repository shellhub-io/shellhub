import type { ElementType, ComponentPropsWithoutRef, ReactNode } from "react";
import { Spinner } from "./Spinner";
import { cn } from "./cn";

// ---------------------------------------------------------------------------
// Public types
// ---------------------------------------------------------------------------

export type ButtonVariant =
  | "primary"
  | "secondary"
  | "surface"
  | "ghost"
  | "destructive"
  | "dangerSoft"
  | "warningSoft"
  | "successSoft"
  | "success"
  | "warning"
  | "outline";
export type ButtonSize = "sm" | "md" | "lg" | "xl";

// ---------------------------------------------------------------------------
// Static maps (no runtime allocations per render)
// ---------------------------------------------------------------------------

const VARIANT: Record<ButtonVariant, string> = {
  primary:
    "bg-primary text-white hover:bg-primary/90 focus-visible:ring-primary",
  secondary:
    "bg-surface border border-border text-text-primary hover:border-border-light hover:bg-hover-subtle focus-visible:ring-primary",
  surface:
    "bg-surface border border-border text-text-primary hover:border-border-light hover:bg-white/[0.04] focus-visible:ring-primary",
  ghost:
    "bg-transparent text-text-secondary hover:text-text-primary hover:bg-hover-subtle focus-visible:ring-primary",
  destructive:
    "bg-accent-red text-white hover:bg-accent-red/90 focus-visible:ring-accent-red",
  dangerSoft:
    "bg-accent-red/10 hover:bg-accent-red/20 text-accent-red border border-accent-red/20 focus-visible:ring-accent-red",
  warningSoft:
    "bg-accent-yellow/10 hover:bg-accent-yellow/20 text-accent-yellow border border-accent-yellow/20 focus-visible:ring-accent-yellow",
  successSoft:
    "bg-accent-green/10 hover:bg-accent-green/20 text-accent-green border border-accent-green/20 focus-visible:ring-accent-green",
  success:
    "bg-accent-green/90 hover:bg-accent-green text-white focus-visible:ring-accent-green",
  warning:
    "bg-accent-yellow hover:bg-accent-yellow/80 text-background focus-visible:ring-accent-yellow",
  outline:
    "bg-surface border border-border text-text-secondary hover:text-text-primary hover:border-border-light hover:bg-white/[0.04] focus-visible:ring-primary",
};

const SIZE: Record<ButtonSize, string> = {
  sm: "px-3 py-1.5 text-xs rounded-md",
  md: "px-4 py-2 text-sm rounded-lg",
  lg: "px-5 py-2.5 text-base rounded-lg",
  xl: "px-8 py-3.5 text-[15px] rounded-xl",
};

const BASE =
  "inline-flex items-center justify-center gap-2 font-medium transition-all duration-300 select-none" +
  " focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-offset-2 focus-visible:ring-offset-background" +
  " disabled:opacity-50 disabled:cursor-not-allowed";

// ---------------------------------------------------------------------------
// Polymorphic helper types (same pattern as Card)
// ---------------------------------------------------------------------------

type ButtonOwnProps<T extends ElementType> = {
  as?: T;
  variant?: ButtonVariant;
  size?: ButtonSize;
  /** Leading icon — replaced by Spinner when loading */
  icon?: ReactNode;
  /** Trailing icon */
  iconRight?: ReactNode;
  /** Adds a shadow/scale glow effect */
  glow?: boolean;
  /** Shows a Spinner, disables interaction */
  loading?: boolean;
  /** Stretch to fill the parent width */
  fullWidth?: boolean;
  className?: string;
  children?: ReactNode;
};

type ButtonProps<T extends ElementType = "button"> = ButtonOwnProps<T> &
  Omit<ComponentPropsWithoutRef<T>, keyof ButtonOwnProps<T>>;

// ---------------------------------------------------------------------------
// Component
// ---------------------------------------------------------------------------

export function Button<T extends ElementType = "button">({
  as,
  variant = "primary",
  size = "md",
  icon,
  iconRight,
  glow = false,
  loading = false,
  fullWidth = false,
  className,
  children,
  ...rest
}: ButtonProps<T>) {
  const Component = (as ?? "button") as ElementType;
  const isNativeButton = !as || as === "button";

  // Strip `disabled` and `type` from rest before computing interaction props so
  // they don't accidentally override loading-derived values when spread later.
  const {
    type: callerType,
    disabled: callerDisabled,
    ...restWithoutInteraction
  } = rest as Record<string, unknown> & { type?: unknown; disabled?: unknown };

  // When loading, always treat the button as disabled regardless of the
  // caller's `disabled` prop so the two can't fight each other.
  const isDisabled = loading || Boolean(callerDisabled);

  // For non-button elements (e.g. <a>) we use aria-disabled instead of disabled
  const interactionProps = isNativeButton
    ? {
        type: (callerType as string | undefined) ?? "button",
        disabled: isDisabled || undefined,
        "aria-busy": loading ? ("true" as const) : undefined,
      }
    : {
        "aria-disabled": isDisabled ? ("true" as const) : undefined,
      };

  // Remove `type` from rest for non-button elements so it isn't forwarded
  const { type: _type, ...restWithoutType } = restWithoutInteraction;
  const forwardedRest = isNativeButton
    ? restWithoutInteraction
    : restWithoutType;

  const glowClasses = glow
    ? "shadow-lg shadow-primary/30 scale-[1.02] hover:shadow-xl hover:shadow-primary/40"
    : undefined;

  const loadingClasses =
    !isNativeButton && loading ? "pointer-events-none" : undefined;

  const buttonClass = cn(
    BASE,
    VARIANT[variant],
    SIZE[size],
    fullWidth && "w-full",
    glowClasses,
    loadingClasses,
    className,
  );

  // Resolve leading icon: replaced by Spinner when loading.
  // The Spinner is decorative here — the button's own text and aria-busy
  // already convey the loading state to assistive technology.
  const leadingIcon = loading ? (
    <Spinner
      size="sm"
      tone={
        variant === "warning"
          ? "onBackground"
          : variant === "primary" ||
              variant === "destructive" ||
              variant === "success"
            ? "onPrimary"
            : "onSurface"
      }
    />
  ) : (
    icon
  );

  return (
    <Component className={buttonClass} {...interactionProps} {...forwardedRest}>
      {leadingIcon}
      {children}
      {iconRight}
    </Component>
  );
}
