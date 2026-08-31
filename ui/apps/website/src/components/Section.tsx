import type { ComponentPropsWithoutRef, ElementType } from "react";
import { cn } from "@shellhub/design-system/cn";

type Padding = "lg" | "md" | "none";
type Background = "surface" | "none";

/**
 * The props Section itself understands, kept separate from the host element's so the two can be
 * subtracted in SectionProps.
 */
export type SectionOwnProps<E extends ElementType = "section"> = {
  as?: E;
  bordered?: boolean;
  padding?: Padding;
  background?: Background;
  centered?: boolean;
  container?: boolean;
  containerClassName?: string;
};

/**
 * Props of Section: its own, plus everything the element named by as accepts, minus the names
 * Section has already claimed.
 */
export type SectionProps<E extends ElementType = "section"> =
  SectionOwnProps<E> &
    Omit<ComponentPropsWithoutRef<E>, keyof SectionOwnProps<E>>;

const paddingClasses: Record<Padding, string> = {
  lg: "py-24",
  md: "py-12",
  none: "",
};

/**
 * A page section with the site's vertical rhythm, optional top border and centred container.
 * Set container false when the section needs to bleed to the viewport edge and manage its own
 * width — a nested container would otherwise pad it twice.
 */
export function Section<E extends ElementType = "section">({
  as,
  bordered = true,
  padding = "lg",
  background = "none",
  centered = false,
  container = true,
  className,
  containerClassName,
  children,
  ...rest
}: SectionProps<E>) {
  const Tag: ElementType = as ?? "section";

  const outerClass = cn(
    bordered && "border-t border-border",
    paddingClasses[padding],
    background === "surface" && "bg-surface",
    className,
  );

  return (
    <Tag className={outerClass} {...rest}>
      {container ? (
        <div
          className={cn(
            "max-w-7xl mx-auto px-8",
            centered && "text-center",
            containerClassName,
          )}
        >
          {children}
        </div>
      ) : (
        children
      )}
    </Tag>
  );
}
