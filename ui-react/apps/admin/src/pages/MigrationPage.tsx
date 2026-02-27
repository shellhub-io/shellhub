import { CircleStackIcon } from "@heroicons/react/24/outline";

export default function MigrationPage() {
  return (
    <div className="relative min-h-screen flex flex-col items-center justify-center bg-background overflow-hidden">
      {/* Background */}
      <div className="absolute inset-0 overflow-hidden pointer-events-none">
        <div className="absolute -top-32 -left-32 w-[500px] h-[500px] bg-primary/[0.06] rounded-full blur-[120px] animate-pulse-subtle" />
        <div className="absolute -bottom-48 -right-32 w-[400px] h-[400px] bg-primary/[0.04] rounded-full blur-[100px] animate-pulse-subtle delay-1000" />
        <div className="absolute top-1/3 right-1/4 w-[300px] h-[300px] bg-primary/[0.03] rounded-full blur-[80px] animate-pulse-subtle delay-[2s]" />

        <div className="connection-line top-[20%] left-0 w-[45%]" />
        <div className="connection-line top-[45%] left-[35%] w-[65%] delay-[1.5s]" />
        <div className="connection-line top-[70%] left-[10%] w-[40%] delay-[0.8s]" />
        <div className="connection-line-v left-[25%] top-0 h-[55%] delay-500" />
        <div className="connection-line-v left-[60%] top-[25%] h-[75%] delay-[1.8s]" />

        <div className="connection-dot top-[20%] left-[25%] delay-300" />
        <div className="connection-dot top-[45%] left-[60%] delay-[1.2s]" />

        <div className="absolute inset-0 grid-bg opacity-50" />
      </div>

      {/* Content */}
      <div className="relative z-10 flex flex-col items-center text-center px-6 animate-fade-in">
        <img
          src="/v2/ui/logo.svg"
          alt="ShellHub"
          className="h-8 mb-10 opacity-50"
        />

        <div className="animate-float mb-6">
          <div className="w-20 h-20 rounded-2xl bg-primary/10 border border-primary/20 flex items-center justify-center shadow-lg shadow-primary/5">
            <CircleStackIcon
              className="w-10 h-10 text-primary/60"
              strokeWidth={1.2}
            />
          </div>
        </div>

        <p className="text-2xs font-mono font-semibold uppercase tracking-wide text-primary/60 mb-2">
          Database Migration
        </p>
        <h1 className="text-2xl font-bold text-text-primary mb-3">
          Migration in progress
        </h1>
        <p className="text-sm text-text-muted max-w-md leading-relaxed mb-8">
          ShellHub is migrating its database to a new format. This may take a
          while depending on the amount of data. The application will be
          available once the migration is complete.
        </p>

        <div className="flex items-center gap-2.5 bg-card/80 border border-border rounded-lg px-4 py-2.5 backdrop-blur-sm">
          <span className="w-3 h-3 border-2 border-primary/30 border-t-primary rounded-full animate-spin" />
          <span className="text-xs font-mono text-text-secondary">
            Migrating data…
          </span>
        </div>
      </div>
    </div>
  );
}
