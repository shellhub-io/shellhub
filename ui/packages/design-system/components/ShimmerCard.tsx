import { cn } from "../primitives/cn";

/**
 * Lays a shimmer over its children on hover. The overlay is pointer-events-none, so it never
 * intercepts a click meant for the card beneath it.
 */
export function ShimmerCard({
  children,
  className = "",
}: {
  children: React.ReactNode;
  className?: string;
}) {
  return (
    <div className={cn("group relative", className)}>
      <div className="shimmer absolute inset-0 rounded-xl overflow-hidden opacity-0 group-hover:opacity-100 transition-opacity duration-500 pointer-events-none" />
      <div className="relative size-full">{children}</div>
    </div>
  );
}
