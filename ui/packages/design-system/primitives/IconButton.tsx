import type { ElementType, ComponentPropsWithoutRef, ReactNode } from "react";
import { Spinner } from "./Spinner";
import { cn } from "./cn";

/**
 * Intent an IconButton carries. Except for filled these rest at the muted text colour and only
 * take their accent on hover, so primary and danger here are quieter than the Button variants
 * of the same name.
 */
export type IconButtonVariant = "ghost" | "primary" | "danger" | "filled";
/**
 * Hit area around the icon. lg is a fixed square, so a row of them lines up whatever they hold.
 */
export type IconButtonSize = "sm" | "md" | "lg";

// Icon-button variants intentionally rest at text-text-muted and accent on hover,
// so `primary` and `danger` here differ from the always-filled Button variants of the same name.
const VARIANT: Record<IconButtonVariant, string> = {
  ghost: "hover:text-text-primary hover:bg-hover-subtle",
  primary: "hover:text-primary hover:bg-primary/10",
  danger:
    "hover:text-accent-red hover:bg-accent-red/10 focus-visible:ring-accent-red",
  filled: "bg-primary text-white hover:bg-primary/90",
};

const SIZE: Record<IconButtonSize, string> = {
  sm: "p-1 rounded",
  md: "p-1.5 rounded-md",
  lg: "w-8 h-8 rounded-lg",
};

const BASE =
  "inline-flex items-center justify-center shrink-0 transition-all duration-150 select-none text-text-muted bg-transparent" +
  " focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-offset-2 focus-visible:ring-offset-background focus-visible:ring-primary" +
  " disabled:opacity-50 disabled:cursor-not-allowed";

type IconButtonOwnProps<T extends ElementType> = {
  as?: T;
  variant?: IconButtonVariant;
  size?: IconButtonSize;
  loading?: boolean;
  className?: string;
  children?: ReactNode;
};

type IconButtonProps<T extends ElementType = "button"> = IconButtonOwnProps<T> &
  Omit<ComponentPropsWithoutRef<T>, keyof IconButtonOwnProps<T>>;

/**
 * Square icon-only button. It carries no text, so the caller owes it an accessible label.
 * As with Button, rendering it as something other than a button moves disabled and loading
 * onto aria-disabled, which the DOM does not enforce.
 */
export function IconButton<T extends ElementType = "button">({
  as,
  variant = "ghost",
  size = "md",
  loading = false,
  className,
  children,
  ...rest
}: IconButtonProps<T>) {
  const Component: ElementType = (as ?? "button");
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
          tone={variant === "filled" ? "onPrimary" : "onSurface"}
        />
      ) : (
        children
      )}
    </Component>
  );
}
