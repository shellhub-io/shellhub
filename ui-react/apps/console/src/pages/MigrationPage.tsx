import { useEffect, useState } from "react";
import {
  CircleStackIcon,
  CheckCircleIcon,
  ExclamationTriangleIcon,
} from "@heroicons/react/24/outline";
import AmbientBackground from "../components/common/AmbientBackground";

type MigrationStatus = "running" | "completed" | "failed" | "unknown";

export default function MigrationPage() {
  const [status, setStatus] = useState<MigrationStatus>("unknown");

  useEffect(() => {
    let timer: ReturnType<typeof setTimeout>;
    let active = true;

    const poll = () => {
      fetch("/api/migration/status")
        .then((res) => res.json())
        .then((data: { status: MigrationStatus }) => {
          if (active) {
            setStatus(data.status);
            if (data.status !== "completed" && data.status !== "failed") {
              timer = setTimeout(poll, 3000);
            }
          }
        })
        .catch(() => {
          if (active) {
            setStatus("unknown");
            timer = setTimeout(poll, 3000);
          }
        });
    };

    poll();

    return () => {
      active = false;
      clearTimeout(timer);
    };
  }, []);

  return (
    <div className="relative min-h-screen flex flex-col items-center justify-center bg-background overflow-hidden">
      <AmbientBackground variant={status === "failed" ? "error" : "default"} />

      <div className="flex flex-col items-center text-center px-6 animate-fade-in">
        <img src="/logo.svg" alt="ShellHub" className="h-8 mb-10 opacity-50" />

        <div className="animate-float mb-6">
          <div
            className={`w-20 h-20 rounded-2xl border flex items-center justify-center shadow-lg ${
              status === "completed"
                ? "bg-accent-green/10 border-accent-green/20 shadow-accent-green/5"
                : status === "failed"
                  ? "bg-accent-red/10 border-accent-red/20 shadow-accent-red/5"
                  : "bg-primary/10 border-primary/20 shadow-primary/5"
            }`}
          >
            {status === "completed" ? (
              <CheckCircleIcon
                className="w-10 h-10 text-accent-green/60"
                strokeWidth={1.2}
              />
            ) : status === "failed" ? (
              <ExclamationTriangleIcon
                className="w-10 h-10 text-accent-red/60"
                strokeWidth={1.2}
              />
            ) : (
              <CircleStackIcon
                className="w-10 h-10 text-primary/60"
                strokeWidth={1.2}
              />
            )}
          </div>
        </div>

        <p
          className={`text-2xs font-mono font-semibold uppercase tracking-wide mb-2 ${
            status === "completed"
              ? "text-accent-green/60"
              : status === "failed"
                ? "text-accent-red/60"
                : "text-primary/60"
          }`}
        >
          Database Migration
        </p>

        {status === "completed" ? (
          <>
            <h1 className="text-2xl font-bold text-text-primary mb-3">
              Migration completed
            </h1>
            <p className="text-sm text-text-muted max-w-md leading-relaxed mb-8">
              The database migration finished successfully. You can now update
              ShellHub to the next version to start using the new database.
            </p>
          </>
        ) : status === "failed" ? (
          <>
            <h1 className="text-2xl font-bold text-text-primary mb-3">
              Migration failed
            </h1>
            <p className="text-sm text-text-muted max-w-md leading-relaxed mb-8">
              Something went wrong during the database migration. Check the API
              logs for details.
            </p>
          </>
        ) : (
          <>
            <h1 className="text-2xl font-bold text-text-primary mb-3">
              Migration in progress
            </h1>
            <p className="text-sm text-text-muted max-w-md leading-relaxed mb-8">
              ShellHub is migrating its database to a new format. This may take
              a while depending on the amount of data.
            </p>
          </>
        )}

        {(status === "running" || status === "unknown") && (
          <div className="flex items-center gap-2.5 bg-card/80 border border-border rounded-lg px-4 py-2.5 backdrop-blur-sm">
            <span className="w-3 h-3 border-2 border-primary/30 border-t-primary rounded-full animate-spin" />
            <span className="text-xs font-mono text-text-secondary">
              Migrating data…
            </span>
          </div>
        )}
      </div>
    </div>
  );
}
