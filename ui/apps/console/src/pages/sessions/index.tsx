import { useEffect, useMemo, useState } from "react";
import { useNavigate } from "react-router-dom";
import { usePaginatedListState } from "@/hooks/usePaginatedListState";
import {
  CommandLineIcon,
  ExclamationTriangleIcon,
  GlobeAltIcon,
  XCircleIcon,
} from "@heroicons/react/24/outline";
import { PlayIcon } from "@heroicons/react/24/solid";
import { useSessions } from "@/hooks/useSessions";
import { useCloseSession } from "@/hooks/useSessionMutations";
import { useSessionRecording } from "@/hooks/useSessionRecording";
import { useRecordingsStore } from "@/stores/recordingsStore";
import { isRecordingSupported, readRecording } from "@/utils/recordings";
import type { Session } from "@/client";
import PageHeader from "@/components/common/PageHeader";
import DeviceChip from "@/components/common/DeviceChip";
import DataTable, { type Column } from "@/components/common/DataTable";
import SessionPlayerDialog from "./SessionPlayerDialog";
import RecordingPaywallDialog from "@/components/sessions/RecordingPaywallDialog";
import RestrictedAction from "@/components/common/RestrictedAction";
import { formatDate, formatDuration } from "@/utils/date";
import { sessionType } from "@/utils/session";
import { isEnterpriseOrCloud } from "@/env";
import {
  Callout,
  IconButton,
  Spinner,
} from "@shellhub/design-system/primitives";
import { cn } from "@shellhub/design-system/cn";

const PER_PAGE = 10;

const PLAY_BTN =
  "inline-flex items-center gap-1.5 px-2.5 py-1.5 text-2xs font-semibold text-white bg-primary rounded-md hover:bg-primary-600 transition-colors disabled:opacity-dim disabled:cursor-not-allowed disabled:hover:bg-primary";

type SessionsParams = {
  page: number;
};

const DEFAULTS: SessionsParams = { page: 1 };

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
    <IconButton
      variant="danger"
      title="Close session"
      aria-label="Close session"
      disabled={closing}
      onClick={(e) => void handleClick(e)}
    >
      <XCircleIcon className="w-4 h-4" strokeWidth={2} />
    </IconButton>
  );
}

export default function Sessions() {
  const { params, setPage } = usePaginatedListState<SessionsParams>({
    defaults: DEFAULTS,
  });
  const { sessions, totalCount, isLoading, error } = useSessions({
    page: params.page,
    perPage: PER_PAGE,
  });
  const closeSession = useCloseSession();
  const navigate = useNavigate();
  const premium = isEnterpriseOrCloud();
  const [playTarget, setPlayTarget] = useState<string | null>(null);
  const [localLogs, setLocalLogs] = useState<string | null>(null);
  const [upsellOpen, setUpsellOpen] = useState(false);
  const {
    logs: sessionLogs,
    isLoading: logsLoading,
    error: logsError,
    fetchLogs,
    clearLogs,
  } = useSessionRecording();

  // Local (OPFS) recordings live only in the browser that made them, keyed by
  // session uid. Surfacing their playback here makes a session the single home
  // for both server and local recordings.
  const recordings = useRecordingsStore((s) => s.recordings);
  const refreshRecordings = useRecordingsStore((s) => s.refresh);

  useEffect(() => {
    if (isRecordingSupported()) void refreshRecordings();
  }, [refreshRecordings]);

  const localBySessionUid = useMemo(
    () =>
      new Map(
        recordings
          .filter((r) => r.sessionUid)
          .map((r) => [r.sessionUid, r] as const),
      ),
    [recordings],
  );

  const totalPages = Math.ceil(totalCount / PER_PAGE);

  // Play routing: prefer a local recording (inline read), fall back to the
  // server recording, and on Community pitch the paid feature when a shell
  // session has no recording at all.
  const handlePlayClick = async (e: React.MouseEvent, s: Session) => {
    e.stopPropagation();
    const local = localBySessionUid.get(s.uid);
    if (local) {
      setPlayTarget(s.uid);
      try {
        setLocalLogs(await readRecording(local));
      } catch {
        setPlayTarget(null);
      }
      return;
    }
    if (s.recorded) {
      setPlayTarget(s.uid);
      await fetchLogs(s.uid);
      return;
    }
    setUpsellOpen(true);
  };

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
      key: "origin",
      header: "Origin",
      render: (s) =>
        s.web ? (
          <span className="inline-flex items-center gap-1.5 text-xs">
            <GlobeAltIcon
              className="w-3.5 h-3.5 text-text-muted"
              strokeWidth={2}
            />
            <span className="text-text-secondary">Web</span>
          </span>
        ) : (
          <span className="inline-flex items-center gap-1.5 text-xs">
            <CommandLineIcon
              className="w-3.5 h-3.5 text-text-muted"
              strokeWidth={2}
            />
            <span className="text-text-secondary">SSH</span>
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
              className={cn(
                "text-xs font-mono",
                suspicious ? "text-accent-red/60" : "text-text-secondary",
              )}
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
            className={cn(
              "inline-flex items-center px-2 py-0.5 text-2xs font-mono font-semibold rounded border",
              type.color,
            )}
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
      render: (s) => {
        const local = localBySessionUid.get(s.uid);
        const isShell = sessionType(s)?.label === "shell";
        const canPlay = Boolean(local) || s.recorded;
        const playing = logsLoading && playTarget === s.uid;
        // A local recording is the user's own browser data, so it plays without
        // the server-side session:play permission; only server playback needs it.
        const needsPermission = !local && s.recorded;
        const playButton = (
          <button
            type="button"
            className={PLAY_BTN}
            disabled={playing || (!canPlay && premium)}
            title={canPlay ? "Play recording" : "This session was not recorded"}
            aria-label="Play recording"
            onClick={(e) => void handlePlayClick(e, s)}
          >
            {playing ? (
              <Spinner size="xs" tone="onPrimary" />
            ) : (
              <PlayIcon className="w-3.5 h-3.5" />
            )}
            Play
          </button>
        );
        return (
          <div className="flex items-center justify-end gap-1.5">
            {isShell &&
              (needsPermission ? (
                <RestrictedAction action="session:play">
                  {playButton}
                </RestrictedAction>
              ) : (
                playButton
              ))}
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
        );
      },
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
        <Callout variant="error" className="mb-4">
          {error.message}
        </Callout>
      )}

      {logsError && (
        <Callout variant="error" className="mb-4">
          {logsError}
        </Callout>
      )}

      <DataTable
        columns={columns}
        data={sessions}
        rowKey={(s) => s.uid}
        isLoading={isLoading}
        loadingMessage="Loading sessions..."
        page={params.page}
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

      {playTarget && (localLogs || (!logsLoading && sessionLogs)) && (
        <SessionPlayerDialog
          open
          onClose={() => {
            setPlayTarget(null);
            setLocalLogs(null);
            clearLogs();
          }}
          logs={localLogs ?? sessionLogs ?? ""}
        />
      )}

      <RecordingPaywallDialog
        open={upsellOpen}
        onClose={() => setUpsellOpen(false)}
      />
    </div>
  );
}
