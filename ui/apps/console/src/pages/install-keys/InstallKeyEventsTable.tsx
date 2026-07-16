import { useState } from "react";
import { Link } from "react-router-dom";
import { useInstallKeyEvents } from "@/hooks/useInstallKeyEvents";
import DataTable from "@/components/common/DataTable";
import { installKeyEventColumns } from "./installKeyEventColumns";

const EMPTY_MESSAGE =
  "No registrations yet. Devices that register with this key will appear here.";

/**
 * An install key's registration-activity table, self-fetching by key id. Two shapes: the paginated
 * full view (the activity page) or a compact view — the first `perPage` rows, no card, with a "View
 * all" link when there are more. The compact form is embedded inline under the tenant-only row on the
 * list, so it drops the wrapper and pagination to sit flush inside its host.
 */
export default function InstallKeyEventsTable({
  id,
  perPage = 15,
  compact = false,
  viewAll,
}: {
  id: string;
  perPage?: number;
  compact?: boolean;
  /** Compact only: a link to the full activity page, shown when more rows exist than are displayed. */
  viewAll?: { to: string; state?: unknown };
}) {
  const [page, setPage] = useState(1);
  const { events, totalCount, isLoading, error } = useInstallKeyEvents({
    id,
    page: compact ? 1 : page,
    perPage,
  });

  if (error) {
    return (
      <div
        role="alert"
        className="text-xs text-accent-red bg-accent-red/[0.06] border border-accent-red/20 rounded-lg px-3 py-2.5"
      >
        Could not load registration activity. Check your connection and try
        again.
      </div>
    );
  }

  return (
    <>
      <DataTable
        label="Registration Activity"
        columns={installKeyEventColumns}
        data={events}
        rowKey={(event) => event.id}
        rowClassName={() => "[&>td]:py-4"}
        isLoading={isLoading}
        loadingMessage="Loading activity..."
        emptyMessage={EMPTY_MESSAGE}
        noWrapper={compact}
        headerTopBorder={compact}
        {...(compact
          ? {}
          : {
              page,
              totalPages: Math.ceil(totalCount / perPage),
              totalCount,
              itemLabel: "registration",
              onPageChange: setPage,
            })}
      />
      {compact && viewAll && totalCount > events.length && (
        <div className="border-t border-border/60 px-4 py-2.5 text-center">
          <Link
            to={viewAll.to}
            state={viewAll.state}
            className="text-2xs font-medium text-primary hover:underline"
          >
            View all {totalCount} registrations →
          </Link>
        </div>
      )}
    </>
  );
}
