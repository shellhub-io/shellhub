import { Spinner } from "@shellhub/design-system/primitives";
import { cn } from "@shellhub/design-system/cn";

type Padding = "none" | "sm" | "md" | "lg" | "fill";

interface PageLoaderProps {
  label: string;
  showLabel?: boolean;
  padding?: Padding;
}

const PADDING: Record<Padding, string> = {
  none: "",
  sm: "py-12",
  md: "py-24",
  lg: "py-32",
  fill: "flex-1",
};

/**
 * The loading state for a whole page. The label is hidden by default: a spinner beside text
 * already saying "loading" is announced twice.
 */
export default function PageLoader({
  label,
  showLabel = false,
  padding = "md",
}: PageLoaderProps) {
  const wrapper = cn("flex h-full items-center justify-center", PADDING[padding]);

  if (showLabel) {
    return (
      <div role="status" aria-label={label} className={cn(wrapper, "gap-3")}>
        <Spinner size="md" />
        <span className="text-xs text-text-muted">{label}</span>
      </div>
    );
  }

  return (
    <div className={wrapper}>
      <Spinner size="lg" aria-label={label} />
    </div>
  );
}
