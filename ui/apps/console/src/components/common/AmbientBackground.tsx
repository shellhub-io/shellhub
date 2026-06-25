import { ConnectionGrid } from "@shellhub/design-system/components";

interface AmbientBackgroundProps {
  variant?: "default" | "error";
}

export default function AmbientBackground({
  variant = "default",
}: AmbientBackgroundProps) {
  const isError = variant === "error";

  return (
    <div className="absolute inset-0 overflow-hidden pointer-events-none">
      {/* Gradient blobs */}
      <div
        className={`absolute -top-32 -left-32 w-[500px] h-[500px] rounded-full blur-[120px] animate-pulse-subtle ${
          isError ? "bg-accent-red/[0.06]" : "bg-primary/10"
        }`}
      />
      <div
        className={`absolute -bottom-48 -right-32 w-[400px] h-[400px] rounded-full blur-[100px] animate-pulse-subtle ${
          isError ? "bg-primary/[0.04]" : "bg-accent-cyan/8"
        }`}
        style={{ animationDelay: "1s" }}
      />
      <div
        className={`absolute top-1/3 right-1/4 w-[300px] h-[300px] rounded-full blur-[80px] animate-pulse-subtle ${
          isError ? "bg-accent-red/[0.03]" : "bg-accent-blue/5"
        }`}
        style={{ animationDelay: "2s" }}
      />

      <ConnectionGrid />

      {/* Grid */}
      <div className="absolute inset-0 grid-bg opacity-50" />
    </div>
  );
}
