import { ConnectionGrid, GlowOrbs } from "@shellhub/design-system/components";

interface AmbientBackgroundProps {
  variant?: "default" | "error";
}

/**
 * The full-page background behind the unauthenticated screens. The error variant tints it red,
 * so a failure reads before the text is.
 */
export default function AmbientBackground({
  variant = "default",
}: AmbientBackgroundProps) {
  const isError = variant === "error";

  return (
    <div className="absolute inset-0 overflow-hidden pointer-events-none">
      <GlowOrbs preset="ambient" tone={isError ? "error" : "brand"} />

      <ConnectionGrid />

      {/* Grid */}
      <div className="absolute inset-0 grid-bg opacity-50" />
    </div>
  );
}
