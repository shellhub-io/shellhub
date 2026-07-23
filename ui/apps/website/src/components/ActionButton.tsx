import type { ReactNode } from "react";
import { Link } from "react-router-dom";
import { Button, type ButtonSize } from "@shellhub/design-system/primitives";
import { cn } from "@shellhub/design-system/cn";
import { ArrowRight } from "@/components";

export type CTAAction =
  | { label: string; to: string; href?: never; external?: never }
  | { label: string; href: string; to?: never; external?: boolean };

export interface ActionButtonProps {
  action: CTAAction;
  variant?: "primary" | "outline";
  size?: ButtonSize;
  glow?: boolean;
  icon?: ReactNode;
  iconRight?: ReactNode;
  fullWidth?: boolean;
}

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

export interface ActionButtonGroupProps {
  primaryAction: CTAAction;
  secondaryAction: CTAAction;
  size?: ButtonSize;
  className?: string;
}

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
