import { useState } from "react";
import { useNavigate } from "react-router-dom";
import {
  CommandLineIcon,
  ExclamationCircleIcon,
  ExclamationTriangleIcon,
  ShieldCheckIcon,
  ShieldExclamationIcon,
} from "@heroicons/react/24/outline";
import { useAdminSessionsList } from "@/hooks/useAdminSessionsList";
import type { Session } from "@/client";
import PageHeader from "@/components/common/PageHeader";
import DataTable, { type Column } from "@/components/common/DataTable";
import DeviceChip from "@/components/common/DeviceChip";
import { formatDateFull } from "@/utils/date";

const PER_PAGE = 10;

export default function AdminSessions() {
  const [page, setPage] = useState(1);
  const { sessions, totalCount, isLoading, error } = useAdminSessionsList(
    page,
    PER_PAGE,
  );
  const navigate = useNavigate();

  const totalPages = Math.ceil(totalCount / PER_PAGE);

  const columns: Column<Session>[] = [
    {
      key: "active",
      header: "Active",
      headerClassName: "w-14",
      render: (s) => (
        <span
          className={`w-2 h-2 rounded-full inline-block ${
            s.active
              ? "bg-accent-green shadow-[0_0_6px_rgba(130,165,104,0.4)]"
              : "bg-text-muted/40"
          }`}
        />
      ),
    },
    {
      key: "id",
      header: "ID",
      render: (s) => (
        <code
          className="text-xs font-mono text-text-muted bg-surface px-1.5 py-0.5 rounded"
          title={s.uid}
        >
          {s.uid.substring(0, 10)}
        </code>
      ),
    },
    {
      key: "device",
      header: "Device",
      render: (s) =>
        s.device ? (
          <DeviceChip
            disableLink
            name={s.device.name}
            online={s.device.online}
            osId={s.device.info?.id}
          />
        ) : (
          <span className="text-xs font-mono text-text-primary">
            {(s.device_uid ?? "").substring(0, 8)}
          </span>
        ),
    },
    {
      key: "username",
      header: "Username",
      render: (s) => {
        const suspicious = !s.authenticated;
        return (
          <div className="flex items-center gap-1.5">
            {suspicious && (
              <ExclamationTriangleIcon
                className="w-3.5 h-3.5 text-accent-red/70 shrink-0"
                strokeWidth={2}
                title="Not authenticated"
              />
            )}
            <code
              className={`text-xs font-mono ${
                suspicious ? "text-accent-red/60" : "text-text-secondary"
              }`}
            >
              {s.username}
            </code>
          </div>
        );
      },
    },
    {
      key: "auth",
      header: "Auth",
      headerClassName: "w-14",
      render: (s) =>
        s.authenticated ? (
          <ShieldCheckIcon
            className="w-4 h-4 text-accent-green"
            strokeWidth={2}
            title="Authenticated"
          />
        ) : (
          <ShieldExclamationIcon
            className="w-4 h-4 text-accent-red"
            strokeWidth={2}
            title="Not authenticated"
          />
        ),
    },
    {
      key: "ip",
      header: "IP Address",
      render: (s) => (
        <code className="text-xs font-mono text-text-muted bg-surface px-1.5 py-0.5 rounded">
          {s.ip_address}
        </code>
      ),
    },
    {
      key: "started",
      header: "Started",
      render: (s) => (
        <span className="text-xs text-text-secondary">
          {formatDateFull(s.started_at)}
        </span>
      ),
    },
    {
      key: "last_seen",
      header: "Last Seen",
      render: (s) => (
        <span className="text-xs text-text-secondary">
          {formatDateFull(s.last_seen)}
        </span>
      ),
    },
  ];

  return (
    <div>
      <PageHeader
        icon={<CommandLineIcon className="w-6 h-6" />}
        overline="Admin"
        title="Sessions"
        description="Track live and historical sessions happening across every namespace."
      />

      {error && (
        <div
          role="alert"
          className="flex items-center gap-2 bg-accent-red/8 border border-accent-red/20 text-accent-red px-3.5 py-2.5 rounded-md text-xs font-mono mb-4 animate-slide-down"
        >
          <ExclamationCircleIcon
            className="w-3.5 h-3.5 shrink-0"
            strokeWidth={2}
          />
          {error.message}
        </div>
      )}

      <DataTable
        columns={columns}
        data={sessions}
        rowKey={(s) => s.uid}
        isLoading={isLoading}
        loadingMessage="Loading sessions..."
        page={page}
        totalPages={totalPages}
        totalCount={totalCount}
        itemLabel="session"
        onPageChange={setPage}
        onRowClick={(s) => void navigate(`/admin/sessions/${s.uid}`)}
        // border-l-2 on every row (transparent by default) keeps the row
        // height stable when the red border appears on unauthenticated rows.
        rowClassName={(s) =>
          !s.authenticated
            ? "bg-accent-red/[0.03] hover:bg-accent-red/[0.06] border-l-2 border-l-accent-red/50"
            : "border-l-2 border-l-transparent"
        }
        emptyState={
          <div className="text-center">
            <CommandLineIcon
              className="w-10 h-10 text-text-muted/30 mx-auto mb-3"
              strokeWidth={1}
            />
            <p className="text-xs font-mono text-text-muted">
              No sessions found
            </p>
          </div>
        }
      />
    </div>
  );
}
