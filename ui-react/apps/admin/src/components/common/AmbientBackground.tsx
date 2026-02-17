export default function AmbientBackground() {
  return (
    <div className="absolute inset-0 overflow-hidden pointer-events-none">
      {/* Gradient blobs */}
      <div className="absolute -top-32 -left-32 w-[500px] h-[500px] bg-primary/10 rounded-full blur-[120px] animate-pulse-subtle" />
      <div
        className="absolute -bottom-48 -right-32 w-[400px] h-[400px] bg-accent-cyan/8 rounded-full blur-[100px] animate-pulse-subtle"
        style={{ animationDelay: "1s" }}
      />
      <div
        className="absolute top-1/3 right-1/4 w-[300px] h-[300px] bg-accent-blue/5 rounded-full blur-[80px] animate-pulse-subtle"
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
