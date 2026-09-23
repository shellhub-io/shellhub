import { Link, useNavigate } from "react-router-dom";
import { CommandLineIcon } from "@heroicons/react/24/outline";
import { cn } from "@shellhub/design-system/cn";
import { Callout, Card } from "@shellhub/design-system/primitives";
import DataTable, { type Column } from "@/components/common/DataTable";
import DeviceChip from "@/components/common/DeviceChip";
import { useSessions } from "@/hooks/useSessions";
import { useAdminSessions } from "@/hooks/useAdminSessions";
import { formatRelative } from "@/utils/date";
import SessionTypeBadge from "@/components/sessions/SessionTypeBadge";
import SessionLogin from "@/components/sessions/SessionLogin";
import EmptyCell from "@/components/common/EmptyCell";
import type { Session } from "@/client";
import { apiErrorMessage } from "@/api/errors";

/**
 * The five most recent sessions, for a dashboard. isAdmin switches it to the instance-wide list,
 * which is the only difference between the two dashboards' tables.
 */
export default function RecentSessionsTable({ isAdmin = false }) {
  const sessionsHook = isAdmin ? useAdminSessions : useSessions;
  const { sessions, isLoading, error } = sessionsHook({ page: 1, perPage: 5 });
  const navigate = useNavigate();
  const prefix = isAdmin ? "/admin" : "";

  const columns: Column<Session>[] = [
    {
      key: "active",
      header: "Active",
      headerClassName: "w-14",
      render: (s) => (
        <span
          className={cn(
            "w-2 h-2 rounded-full inline-block",
            s.active
              ? "bg-accent-green shadow-[0_0_6px_rgba(130,165,104,0.4)]"
              : "bg-text-muted/40",
          )}
        />
      ),
    },
    {
      key: "type",
      header: "Type",
      render: (s) => (
        <SessionTypeBadge
          session={s}
          shape="rounded"
          fallback={<EmptyCell />}
        />
      ),
    },
    {
      key: "device",
      header: "Device",
      render: (s) =>
        s.device?.uid ? (
          <DeviceChip
            uid={s.device.uid}
            name={s.device.name ?? (s.device_uid ?? "").substring(0, 8)}
            online={s.device.online}
            osId={s.device.info?.id}
            isAdmin={isAdmin}
            onClick={(e) => e.stopPropagation()}
          />
        ) : (
          <span className="text-xs font-mono text-text-primary">
            {s.device?.name ?? (s.device_uid ?? "").substring(0, 8)}
          </span>
        ),
    },
    {
      key: "username",
      header: "Login",
      render: (s) => <SessionLogin session={s} />,
    },
    {
      key: "started",
      header: "Started",
      render: (s) => (
        <span className="text-xs text-text-secondary">
          {formatRelative(s.started_at)}
        </span>
      ),
    },
  ];

  return (
    <>
      {error && (
        <Callout variant="error" className="mb-4">
          {apiErrorMessage(error)}
        </Callout>
      )}

      <div className="mb-4 flex items-center justify-between">
        <p className="text-2xs font-mono font-semibold uppercase tracking-label text-text-muted">
          Recent Sessions
        </p>
        <Link
          to={`${prefix}/sessions`}
          className="text-xs font-medium text-primary hover:text-primary-400 transition-colors"
        >
          View all &rarr;
        </Link>
      </div>

      <Card
        className="overflow-hidden animate-slide-up"
        style={{ animationDelay: "300ms" }}
      >
        <DataTable
          columns={columns}
          data={sessions}
          rowKey={(s) => s.uid}
          noWrapper
          isLoading={isLoading}
          loadingMessage="Loading sessions..."
          onRowClick={(s) => void navigate(`${prefix}/sessions/${s.uid}`)}
          rowClassName={(s) =>
            !s.authenticated
              ? "bg-accent-red/[0.03] hover:bg-accent-red/[0.06] border-l-2 border-l-accent-red/50"
              : "border-l-2 border-l-transparent"
          }
          emptyState={
            <div className="flex flex-col items-center justify-center">
              <CommandLineIcon className="w-8 h-8 mb-3 opacity-40 text-text-muted" />
              <p className="text-sm text-text-muted">No recent sessions</p>
            </div>
          }
        />
      </Card>
    </>
  );
}
