import type { ElementType, ComponentPropsWithoutRef, ReactNode } from "react";
import { Spinner } from "./Spinner";
import { cn } from "./cn";

/**
 * Intent a Button carries. The soft variants are tinted rather than filled, for a destructive
 * or cautionary action that is not the primary one on the screen.
 */
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
/**
 * Padding, text size and corner radius together; md is the size a form or a toolbar uses.
 */
export type ButtonSize = "sm" | "md" | "lg" | "xl";

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

/**
 * Button, or whatever element as names — a router Link, an anchor. Rendered as anything but a
 * button it cannot be disabled by the DOM, so loading and a caller's disabled are carried by
 * aria-disabled and pointer-events instead, and the click has to be guarded by the caller.
 * type defaults to button so a button inside a form does not submit it by accident.
 */
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
  const Component: ElementType = as ?? "button";
  const isNativeButton = !as || as === "button";

  const {
    type: callerType,
    disabled: callerDisabled,
    ...restWithoutInteraction
  } = rest as Record<string, unknown> & { type?: unknown; disabled?: unknown };

  const isDisabled = loading || Boolean(callerDisabled);

  const interactionProps = isNativeButton
    ? {
        type: (callerType as string | undefined) ?? "button",
        disabled: isDisabled || undefined,
        "aria-busy": loading ? ("true" as const) : undefined,
      }
    : {
        "aria-disabled": isDisabled ? ("true" as const) : undefined,
      };

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
