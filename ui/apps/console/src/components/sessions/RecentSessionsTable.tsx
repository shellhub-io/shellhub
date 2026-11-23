import { Link, useNavigate } from "react-router-dom";
import { useQuery } from "@tanstack/react-query";
import { CommandLineIcon } from "@heroicons/react/24/outline";
import { Callout, Card } from "@shellhub/design-system/primitives";
import DataTable, { type Column } from "@/components/common/DataTable";
import {
  sessionColumns,
  sessionRowClassName,
} from "@/components/sessions/sessionColumns";
import { getSessions, getSessionsQueryKey, type Session } from "@/client";
import { paginatedQueryOptions } from "@/api/pagination";
import { apiErrorMessage } from "@/api/errors";

/**
 * The five most recent sessions, for the dashboard.
 */
export default function RecentSessionsTable() {
  const { data, isLoading, error } = useQuery(
    paginatedQueryOptions(getSessions, getSessionsQueryKey, {
      query: { page: 1, per_page: 5 },
    }),
  );
  const sessions = data?.data ?? [];
  const navigate = useNavigate();

  const columns: Column<Session>[] = [
    sessionColumns.active,
    sessionColumns.type,
    sessionColumns.device,
    sessionColumns.username,
    sessionColumns.started,
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
          to="/sessions"
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
          onRowClick={(s) => void navigate(`/sessions/${s.uid}`)}
          rowClassName={sessionRowClassName}
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
