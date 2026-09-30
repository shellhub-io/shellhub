import { useEffect, useState } from "react";
import { Callout, Spinner } from "@shellhub/design-system/primitives";
import FramedShell from "@/components/layout/FramedShell";
import ScreenIntro from "@/components/layout/ScreenIntro";

type MigrationStatus = "running" | "completed" | "failed" | "unknown";

const COPY: Record<
  Exclude<MigrationStatus, "unknown">,
  { title: string; lead: string }
> = {
  completed: {
    title: "Migration completed",
    lead: "The database migration finished successfully. You can now update ShellHub to the next version to start using the new database.",
  },
  failed: {
    title: "Migration failed",
    lead: "Something went wrong during the database migration. Check the API logs for details.",
  },
  running: {
    title: "Migration in progress",
    lead: "ShellHub is migrating its database to a new format. This may take a while depending on the amount of data.",
  },
};

/**
 * The screen shown while the server migrates its database, served from its own entry point
 * rather than the SPA. It polls the migration status until it completes or fails.
 */
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

  const { title, lead } = COPY[status === "unknown" ? "running" : status];

  return (
    <FramedShell>
      <ScreenIntro eyebrow="Database migration" title={title} lead={lead} />
      {status === "completed" && (
        <Callout variant="success">The database is ready.</Callout>
      )}
      {status === "failed" && (
        <Callout variant="error">The migration did not complete.</Callout>
      )}
      {(status === "running" || status === "unknown") && (
        <div className="flex items-center gap-2.5 text-xs font-mono text-text-secondary">
          <Spinner size="xs" />
          Migrating data…
        </div>
      )}
    </FramedShell>
  );
}
