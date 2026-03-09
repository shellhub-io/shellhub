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

      {/* Connection lines */}
      <div
        className="connection-line"
        style={{ top: "18%", left: 0, width: "45%", animationDelay: "0s" }}
      />
      <div
        className="connection-line"
        style={{
          top: "42%",
          left: "35%",
          width: "65%",
          animationDelay: "1.5s",
        }}
      />
      <div
        className="connection-line"
        style={{
          top: "68%",
          left: "10%",
          width: "40%",
          animationDelay: "0.8s",
        }}
      />
      <div
        className="connection-line"
        style={{
          top: "88%",
          left: "50%",
          width: "50%",
          animationDelay: "2.5s",
        }}
      />
      <div
        className="connection-line-v"
        style={{ left: "22%", top: 0, height: "55%", animationDelay: "0.5s" }}
      />
      <div
        className="connection-line-v"
        style={{
          left: "58%",
          top: "25%",
          height: "75%",
          animationDelay: "1.8s",
        }}
      />
      <div
        className="connection-line-v"
        style={{ left: "82%", top: "10%", height: "45%", animationDelay: "3s" }}
      />

      {/* Dots */}
      <div
        className="connection-dot"
        style={{ top: "18%", left: "22%", animationDelay: "0.3s" }}
      />
      <div
        className="connection-dot"
        style={{ top: "42%", left: "58%", animationDelay: "1.2s" }}
      />
      <div
        className="connection-dot"
        style={{ top: "68%", left: "22%", animationDelay: "2s" }}
      />
      <div
        className="connection-dot"
        style={{ top: "42%", left: "82%", animationDelay: "0.8s" }}
      />

      {/* Grid */}
      <div className="absolute inset-0 grid-bg opacity-50" />
    </div>
  );
}
