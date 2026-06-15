import type { ElementType, ComponentPropsWithoutRef, ReactNode } from "react";
import { Spinner } from "./Spinner";
import { cn } from "./cn";

// ---------------------------------------------------------------------------
// Public types
// ---------------------------------------------------------------------------

export type IconButtonVariant = "ghost" | "primary" | "danger";
export type IconButtonSize = "sm" | "md" | "lg";

// ---------------------------------------------------------------------------
// Static maps
// ---------------------------------------------------------------------------

const VARIANT: Record<IconButtonVariant, string> = {
  ghost:
    "bg-transparent text-text-primary hover:bg-hover-subtle focus-visible:ring-primary",
  primary:
    "bg-primary text-white hover:bg-primary/90 focus-visible:ring-primary",
  danger:
    "bg-transparent text-accent-red hover:bg-accent-red/10 focus-visible:ring-accent-red",
};

const SIZE: Record<IconButtonSize, string> = {
  sm: "p-1 rounded",
  md: "p-1.5 rounded-md",
  lg: "w-8 h-8 rounded-lg",
};

const BASE =
  "inline-flex items-center justify-center shrink-0 transition-all duration-150 select-none" +
  " focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-offset-2 focus-visible:ring-offset-background" +
  " disabled:opacity-50 disabled:cursor-not-allowed";

// ---------------------------------------------------------------------------
// Polymorphic helper types (same pattern as Button)
// ---------------------------------------------------------------------------

type IconButtonOwnProps<T extends ElementType> = {
  as?: T;
  variant?: IconButtonVariant;
  size?: IconButtonSize;
  /** Shows a Spinner, disables interaction */
  loading?: boolean;
  className?: string;
  children?: ReactNode;
};

type IconButtonProps<T extends ElementType = "button"> = IconButtonOwnProps<T> &
  Omit<ComponentPropsWithoutRef<T>, keyof IconButtonOwnProps<T>>;

// ---------------------------------------------------------------------------
// Component
// ---------------------------------------------------------------------------

export function IconButton<T extends ElementType = "button">({
  as,
  variant = "ghost",
  size = "md",
  loading = false,
  className,
  children,
  ...rest
}: IconButtonProps<T>) {
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

  const loadingClasses =
    !isNativeButton && loading ? "pointer-events-none" : undefined;

  const buttonClass = cn(
    BASE,
    VARIANT[variant],
    SIZE[size],
    loadingClasses,
    className,
  );

  return (
    <Component className={buttonClass} {...interactionProps} {...forwardedRest}>
      {loading ? (
        <Spinner
          size="sm"
          tone={variant === "primary" ? "onPrimary" : "onSurface"}
        />
      ) : (
        children
      )}
    </Component>
  );
}
