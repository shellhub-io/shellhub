import type { ElementType, ComponentPropsWithoutRef, ReactNode } from "react";
import { cn } from "./cn";

type CardProps<T extends ElementType> = {
  as?: T;
  hover?: boolean;
  className?: string;
  children?: ReactNode;
} & Omit<ComponentPropsWithoutRef<T>, "as" | "className" | "children">;

export function Card<T extends ElementType = "div">({
  as,
  hover,
  className,
  children,
  ...rest
}: CardProps<T>) {
  const Component = (as ?? "div") as ElementType;

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
