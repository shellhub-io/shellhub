import { useState } from "react";
import { useNavigate } from "react-router-dom";
import { CommandLineIcon, XCircleIcon } from "@heroicons/react/24/outline";
import { PlayIcon } from "@heroicons/react/24/solid";
import {
  Callout,
  IconButton,
  Spinner,
} from "@shellhub/design-system/primitives";
import { cn } from "@shellhub/design-system/cn";
import { usePaginatedListState } from "@/hooks/usePaginatedListState";
import { useSessions } from "@/hooks/useSessions";
import { useCloseSession } from "@/hooks/useSessionMutations";
import { useSessionRecording } from "@/hooks/useSessionRecording";
import { useLocalRecordings } from "@/hooks/useLocalRecordings";
import { useRecordingPermissions } from "@/hooks/useRecordingPermissions";
import type { Session } from "@/client";
import PageHeader from "@/components/common/PageHeader";
import DeviceChip from "@/components/common/DeviceChip";
import DataTable, { type Column } from "@/components/common/DataTable";
import RecordingPaywallDialog from "@/components/sessions/RecordingPaywallDialog";
import RestrictedAction from "@/components/common/RestrictedAction";
import RecordingActionsMenu from "@/components/sessions/RecordingActionsMenu";
import { formatRelative, formatDuration } from "@/utils/date";
import { sessionHasTerminal } from "@/utils/session";
import { usePrincipalName } from "@/hooks/usePrincipalName";
import SessionPrincipal from "@/components/sessions/SessionPrincipal";
import SessionTypeBadge from "@/components/sessions/SessionTypeBadge";
import SessionLogin from "@/components/sessions/SessionLogin";
import EmptyCell from "@/components/common/EmptyCell";
import { isEnterpriseOrCloud } from "@/env";
import { apiErrorMessage } from "@/api/errors";
import { PER_PAGE, pageCount } from "@/utils/pagination";
import { useNavSectionTitle } from "@/components/layout/navSections";

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

function SessionActions({
  session,
  playing,
  premium,
  onPlay,
  onDownload,
  onClose,
}: {
  session: Session;
  playing: boolean;
  premium: boolean;
  onPlay: (canPlay: boolean) => void;
  onDownload: () => void;
  onClose: () => Promise<unknown>;
}) {
  const { available: canPlay, read } = useRecordingPermissions(
    session.uid,
    session.recorded,
  );
  const hasTerminal = sessionHasTerminal(session);

  return (
    <div className="flex items-center justify-end gap-1.5">
      {hasTerminal && (
        <RestrictedAction action={read}>
          <button
            type="button"
            className={PLAY_BTN}
            disabled={playing || (!canPlay && premium)}
            title={canPlay ? "Play recording" : "This session was not recorded"}
            aria-label="Play recording"
            onClick={(e) => {
              e.stopPropagation();
              onPlay(canPlay);
            }}
          >
            {playing ? (
              <Spinner size="xs" tone="onPrimary" />
            ) : (
              <PlayIcon className="w-3.5 h-3.5" />
            )}
            Play
          </button>
        </RestrictedAction>
      )}
      {session.active && (
        <RestrictedAction action="session:close">
          <CloseButton onClose={onClose} />
        </RestrictedAction>
      )}
      {hasTerminal && canPlay && (
        <div
          role="presentation"
          onClick={(e) => e.stopPropagation()}
          onKeyDown={(e) => e.stopPropagation()}
        >
          <RecordingActionsMenu
            sessionUid={session.uid}
            recorded={session.recorded}
            onDownload={onDownload}
          />
        </div>
      )}
    </div>
  );
}

/**
 * The sessions list: who connected to what, when, and for how long.
 */
export default function Sessions() {
  const sectionTitle = useNavSectionTitle("/sessions");
  const principalName = usePrincipalName();
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
  const [upsellOpen, setUpsellOpen] = useState(false);
  const {
    isLoading: logsLoading,
    error: logsError,
    play,
    download,
  } = useSessionRecording();

  useLocalRecordings();

  const totalPages = pageCount(totalCount);

  const handlePlay = async (s: Session, canPlay: boolean) => {
    if (!canPlay) {
      setUpsellOpen(true);
      return;
    }
    setPlayTarget(s.uid);
    await play(s);
    setPlayTarget(null);
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
            onClick={(e) => e.stopPropagation()}
          />
        ) : (
          <span className="text-xs font-mono text-text-primary">
            {s.device?.name ?? (s.device_uid ?? "").substring(0, 8)}
          </span>
        ),
    },
    {
      key: "principal",
      header: "Principal",
      render: (s) =>
        s.principal ? (
          <SessionPrincipal
            principal={s.principal}
            name={principalName(s.principal)}
          />
        ) : (
          <EmptyCell />
        ),
    },
    {
      key: "username",
      header: "Login",
      render: (s) => <SessionLogin session={s} />,
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
          {formatRelative(s.started_at)}
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
      render: (s) => (
        <SessionActions
          session={s}
          playing={logsLoading && playTarget === s.uid}
          premium={premium}
          onPlay={(canPlay) => void handlePlay(s, canPlay)}
          onDownload={() => void download(s)}
          onClose={() =>
            closeSession.mutateAsync({
              path: { uid: s.uid },
              body: { device: s.device_uid ?? s.device?.uid ?? "" },
            })
          }
        />
      ),
    },
  ];

  return (
    <div>
      <PageHeader
        icon={<CommandLineIcon className="w-6 h-6" />}
        overline={sectionTitle}
        title="Sessions"
        description="View and monitor all SSH connections to your devices"
      />

      {error && (
        <Callout variant="error" className="mb-4">
          {apiErrorMessage(error)}
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

      <RecordingPaywallDialog
        open={upsellOpen}
        onClose={() => setUpsellOpen(false)}
      />
    </div>
  );
}
