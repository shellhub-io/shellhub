import type { ReactNode } from "react";
import { Link } from "react-router-dom";
import { Button, type ButtonSize } from "@shellhub/design-system/primitives";
import { cn } from "@shellhub/design-system/cn";
import { ArrowRight } from "@/components";

/**
 * A call to action, discriminated on where it goes: to is an in-app route handed to the router,
 * href leaves the site. The never members are what stop a caller supplying both, which would
 * otherwise silently take the router path and break an external link.
 */
export type CTAAction =
  | { label: string; to: string; href?: never; external?: never }
  | { label: string; href: string; to?: never; external?: boolean };

/**
 * Props of ActionButton. Leaving glow and iconRight undefined is not the same as setting them
 * false: a primary action defaults both on, and passing the value explicitly is how a caller
 * turns them off.
 */
export interface ActionButtonProps {
  action: CTAAction;
  variant?: "primary" | "outline";
  size?: ButtonSize;
  glow?: boolean;
  icon?: ReactNode;
  iconRight?: ReactNode;
  fullWidth?: boolean;
}

/**
 * Renders a CTAAction as a button, wrapping it in a router Link or a plain anchor according to
 * which member the action carries. An external href is opened in a new tab with rel=noreferrer.
 */
export function ActionButton({
  action,
  variant = "primary",
  size = "xl",
  glow,
  icon,
  iconRight,
  fullWidth = false,
}: ActionButtonProps) {
  const isPrimary = variant === "primary";
  const resolvedGlow = glow ?? (isPrimary ? true : undefined);
  const resolvedIconRight =
    iconRight !== undefined ? iconRight : isPrimary ? <ArrowRight /> : undefined;

  const shared = {
    variant,
    size,
    glow: resolvedGlow,
    icon,
    iconRight: resolvedIconRight,
    fullWidth,
    children: action.label,
  };

  if (action.to) {
    return <Button as={Link} to={action.to} {...shared} />;
  }

  return (
    <Button
      as="a"
      href={action.href}
      {...shared}
      {...(action.external && {
        target: "_blank",
        rel: "noopener noreferrer",
      })}
    />
  );
}

/**
 * Props of ActionButtonGroup. Both actions are required: the pair is the point, and one on its
 * own should be an ActionButton.
 */
export interface ActionButtonGroupProps {
  primaryAction: CTAAction;
  secondaryAction: CTAAction;
  size?: ButtonSize;
  className?: string;
}

/**
 * The primary and secondary calls to action side by side, stacking on narrow viewports. The
 * secondary is forced to the outline variant, so the pair always reads as one choice and a
 * fallback rather than as two equal buttons.
 */
export function ActionButtonGroup({
  primaryAction,
  secondaryAction,
  size,
  className,
}: ActionButtonGroupProps) {
  return (
    <div
      className={cn("flex flex-col sm:flex-row items-center justify-center gap-3", className)}
    >
      <ActionButton action={primaryAction} size={size} />
      <ActionButton action={secondaryAction} variant="outline" size={size} />
    </div>
  );
}
