import type { ElementType, ComponentPropsWithoutRef, ReactNode } from "react";
import { cn } from "./cn";

type CardProps<T extends ElementType> = {
  as?: T;
  hover?: boolean;
  className?: string;
  children?: ReactNode;
} & Omit<ComponentPropsWithoutRef<T>, "as" | "className" | "children">;

/**
 * Bordered surface that groups related content. hover adds the border lift used when the whole
 * card is a link; leave it off for a static panel, where the movement reads as a false affordance.
 */
export function Card<T extends ElementType = "div">({
  as,
  hover,
  className,
  children,
  ...rest
}: CardProps<T>) {
  const Component: ElementType = as ?? "div";

  return (
    <Component
      className={cn(
        "bg-card border border-border rounded-xl",
        hover && "transition-all duration-300 hover:border-border-light",
        className,
      )}
      {...rest}
    >
      {children}
    </Component>
  );
}
