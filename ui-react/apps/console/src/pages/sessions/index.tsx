import { useState } from "react";
import { useNavigate } from "react-router-dom";
import {
  ExclamationCircleIcon,
  CommandLineIcon,
  ExclamationTriangleIcon,
  XCircleIcon,
} from "@heroicons/react/24/outline";
import { PlayIcon } from "@heroicons/react/24/solid";
import { useSessions } from "@/hooks/useSessions";
import { useCloseSession } from "@/hooks/useSessionMutations";
import { useSessionRecording } from "@/hooks/useSessionRecording";
import type { Session } from "@/client";
import PageHeader from "@/components/common/PageHeader";
import DeviceChip from "@/components/common/DeviceChip";
import DataTable, { type Column } from "@/components/common/DataTable";
import SessionPlayerDialog from "./SessionPlayerDialog";
import RestrictedAction from "@/components/common/RestrictedAction";
import { formatDate, formatDuration } from "@/utils/date";
import { sessionType } from "@/utils/session";

const PER_PAGE = 10;

function CloseButton({ onClose }: { onClose: () => Promise<unknown> }) {
  const [closing, setClosing] = useState(false);

  const handleClick = async (e: React.MouseEvent) => {
    e.stopPropagation();
    setClosing(true);
    try {
      await onClose();
    } finally {
      setClosing(false);
    }
  };

  return (
    <button
      onClick={(e) => void handleClick(e)}
      disabled={closing}
      title="Close session"
      className="p-1.5 rounded-md text-text-muted hover:text-accent-red hover:bg-accent-red/10 transition-colors disabled:opacity-dim"
    >
      <XCircleIcon className="w-4 h-4" strokeWidth={2} />
    </button>
  );
}

export default function Sessions() {
  const [page, setPage] = useState(1);
  const { sessions, totalCount, isLoading, error } = useSessions({
    page,
    perPage: PER_PAGE,
  });
  const closeSession = useCloseSession();
  const navigate = useNavigate();
  const [playTarget, setPlayTarget] = useState<string | null>(null);
  const {
    logs: sessionLogs,
    isLoading: logsLoading,
    error: logsError,
    fetchLogs,
    clearLogs,
  } = useSessionRecording();

  const totalPages = Math.ceil(totalCount / PER_PAGE);

  const handlePlayClick = async (e: React.MouseEvent, uid: string) => {
    e.stopPropagation();
    setPlayTarget(uid);
    await fetchLogs(uid);
  };

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
      key: "device",
      header: "Device",
      render: (s) =>
        s.device?.uid ? (
          <DeviceChip
            uid={s.device.uid}
            name={s.device.name ?? (s.device_uid ?? "").substring(0, 8)}
            online={s.device.online}
            osId={s.device.info?.id}
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
              className={`text-xs font-mono ${suspicious ? "text-accent-red/60" : "text-text-secondary"}`}
            >
              {s.username}
            </code>
          </div>
        );
      },
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
      key: "type",
      header: "Type",
      render: (s) => {
        const type = sessionType(s);
        return type ? (
          <span
            className={`inline-flex items-center px-2 py-0.5 text-2xs font-mono font-semibold rounded border ${type.color}`}
          >
            {type.label}
          </span>
        ) : (
          <span className="text-2xs text-text-muted">{"\u2014"}</span>
        );
      },
    },
    {
      key: "started",
      header: "Started",
      render: (s) => (
        <span className="text-xs text-text-secondary">
          {formatDate(s.started_at)}
        </span>
      ),
    },
    {
      key: "duration",
      header: "Duration",
      render: (s) => (
        <span className="text-xs font-mono text-text-secondary tabular-nums">
          {formatDuration(s.started_at, s.last_seen, s.active)}
        </span>
      ),
    },
    {
      key: "actions",
      header: "",
      headerClassName: "w-28",
      render: (s) => (
        <div className="flex items-center justify-end gap-1">
          {s.recorded && (
            <RestrictedAction action="session:play">
              <button
                onClick={(e) => void handlePlayClick(e, s.uid)}
                disabled={logsLoading && playTarget === s.uid}
                title="Play recording"
                className="inline-flex items-center gap-1 px-2.5 py-1 bg-primary/10 text-primary text-2xs font-semibold rounded-md hover:bg-primary/20 border border-primary/20 transition-all disabled:opacity-dim"
              >
                {logsLoading && playTarget === s.uid ? (
                  <span className="w-3 h-3 border border-primary/40 border-t-primary rounded-full animate-spin" />
                ) : (
                  <PlayIcon className="w-3 h-3" />
                )}
                Play
              </button>
            </RestrictedAction>
          )}
          {s.active && (
            <RestrictedAction action="session:close">
              <CloseButton
                onClose={() =>
                  closeSession.mutateAsync({
                    path: { uid: s.uid },
                    body: { device: s.device_uid ?? s.device?.uid ?? "" },
                  })
                }
              />
            </RestrictedAction>
          )}
        </div>
      ),
    },
  ];

  return (
    <div>
      <PageHeader
        icon={<CommandLineIcon className="w-6 h-6" />}
        overline="SSH Sessions"
        title="Sessions"
        description="View and monitor all SSH connections to your devices"
      />

      {error && (
        <div className="flex items-center gap-2 bg-accent-red/8 border border-accent-red/20 text-accent-red px-3.5 py-2.5 rounded-md text-xs font-mono mb-4 animate-slide-down">
          <ExclamationCircleIcon
            className="w-3.5 h-3.5 shrink-0"
            strokeWidth={2}
          />
          {error.message}
        </div>
      )}

      {logsError && (
        <div className="flex items-center gap-2 bg-accent-red/8 border border-accent-red/20 text-accent-red px-3.5 py-2.5 rounded-md text-xs font-mono mb-4 animate-slide-down">
          <ExclamationCircleIcon
            className="w-3.5 h-3.5 shrink-0"
            strokeWidth={2}
          />
          {logsError}
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
        onRowClick={(s) => void navigate(`/sessions/${s.uid}`)}
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

      {playTarget && !logsLoading && sessionLogs && (
        <SessionPlayerDialog
          open
          onClose={() => {
            setPlayTarget(null);
            clearLogs();
          }}
          logs={sessionLogs}
        />
      )}
    </div>
  );
}
