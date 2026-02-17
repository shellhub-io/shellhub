import { useEffect, useState } from "react";
import { useParams, useNavigate, Link } from "react-router-dom";
import {
  ChevronRightIcon,
  ClockIcon,
  ShieldExclamationIcon,
  ShieldCheckIcon,
  XCircleIcon,
  CommandLineIcon,
  FolderOpenIcon,
  BoltIcon,
  ComputerDesktopIcon,
  VideoCameraIcon,
  ArrowRightCircleIcon,
  CheckCircleIcon,
  KeyIcon,
  ArrowsRightLeftIcon,
  SignalIcon,
} from "@heroicons/react/24/outline";
import { useSessionsStore } from "../stores/sessionsStore";
import CopyButton from "../components/common/CopyButton";
import DeviceChip from "../components/common/DeviceChip";
import DistroIcon from "../components/common/DistroIcon";
import { formatDateFull, formatRelative, formatDuration } from "../utils/date";
import type { Session } from "../types/session";

/* ── timeline builder ────────────────────────────── */

type EventStatus = "success" | "error" | "info" | "active" | "muted";

interface TLEvent {
  id: string;
  icon: React.ReactNode;
  title: string;
  detail?: string;
  status: EventStatus;
}

function buildTimeline(session: Session): TLEvent[] {
  const types = session.events?.types ?? [];
  const events: TLEvent[] = [];

  events.push({
    id: "connect",
    icon: <ArrowRightCircleIcon className="w-3.5 h-3.5" strokeWidth={2} />,
    title: "Connection established",
    detail: `from ${session.ip_address}`,
    status: "info",
  });

  events.push({
    id: "auth",
    icon: session.authenticated ? (
      <ShieldCheckIcon className="w-3.5 h-3.5" strokeWidth={2} />
    ) : (
      <ShieldExclamationIcon className="w-3.5 h-3.5" strokeWidth={2} />
    ),
    title: session.authenticated ? "Authenticated" : "Authentication failed",
    detail: `as ${session.username}`,
    status: session.authenticated ? "success" : "error",
  });

  if (types.includes("auth-agent-req")) {
    events.push({
      id: "agent",
      icon: <KeyIcon className="w-3.5 h-3.5" strokeWidth={2} />,
      title: "Agent forwarding enabled",
      status: "info",
    });
  }

  if (types.includes("pty-req")) {
    events.push({
      id: "pty",
      icon: <ComputerDesktopIcon className="w-3.5 h-3.5" strokeWidth={2} />,
      title: "Pseudo-terminal opened",
      detail:
        session.term && session.term !== "none" ? session.term : undefined,
      status: "info",
    });
  }

  if (types.includes("subsystem")) {
    events.push({
      id: "type",
      icon: <FolderOpenIcon className="w-3.5 h-3.5" strokeWidth={2} />,
      title: "SFTP subsystem started",
      status: "info",
    });
  } else if (types.includes("exec")) {
    events.push({
      id: "type",
      icon: <BoltIcon className="w-3.5 h-3.5" strokeWidth={2} />,
      title: "Command executed",
      status: "info",
    });
  } else if (types.includes("shell")) {
    events.push({
      id: "type",
      icon: <CommandLineIcon className="w-3.5 h-3.5" strokeWidth={2} />,
      title: "Shell session started",
      status: "info",
    });
  }

  if (types.includes("tcpip-forward")) {
    events.push({
      id: "forward",
      icon: <ArrowsRightLeftIcon className="w-3.5 h-3.5" strokeWidth={2} />,
      title: "Port forwarding enabled",
      status: "info",
    });
  }

  if (types.includes("signal")) {
    events.push({
      id: "signal",
      icon: <SignalIcon className="w-3.5 h-3.5" strokeWidth={2} />,
      title: "Signal sent",
      status: "info",
    });
  }

  if (session.recorded) {
    events.push({
      id: "recording",
      icon: <VideoCameraIcon className="w-3.5 h-3.5" strokeWidth={2} />,
      title: "Session recorded",
      status: "info",
    });
  }

  events.push({
    id: "end",
    icon: session.active ? null : (
      <CheckCircleIcon className="w-3.5 h-3.5" strokeWidth={2} />
    ),
    title: session.active ? "Session active" : "Session closed",
    detail: session.active
      ? `started ${formatRelative(session.started_at)}`
      : `closed ${formatRelative(session.last_seen)}`,
    status: session.active ? "active" : "muted",
  });

  return events;
}

/* ── node colors ─────────────────────────────────── */

const NODE_COLORS: Record<EventStatus, string> = {
  success: "text-accent-green border-accent-green/50 bg-accent-green/10",
  error: "text-accent-red border-accent-red/50 bg-accent-red/10",
  info: "text-primary border-primary/40 bg-primary/10",
  active: "text-accent-green border-accent-green/50 bg-accent-green/10",
  muted: "text-text-muted border-border bg-surface",
};

const LABEL =
  "text-2xs font-mono font-semibold uppercase tracking-compact text-text-muted";

/* ── sub-components ──────────────────────────────── */

function TimelineNode({ event, isLast }: { event: TLEvent; isLast: boolean }) {
  return (
    <div className="flex gap-3.5">
      {/* Spine */}
      <div className="flex flex-col items-center shrink-0">
        <div
          className={`w-6 h-6 rounded-full border flex items-center justify-center ${NODE_COLORS[event.status]}`}
        >
          {event.status === "active" && event.icon === null ? (
            <span className="relative flex w-2 h-2">
              <span className="animate-ping absolute inline-flex h-full w-full rounded-full bg-accent-green opacity-60" />
              <span className="relative inline-flex rounded-full w-2 h-2 bg-accent-green" />
            </span>
          ) : (
            event.icon
          )}
        </div>
        {!isLast && (
          <div className="w-px flex-1 bg-border/60 my-1 min-h-[20px]" />
        )}
      </div>

      {/* Content */}
      <div className={`${isLast ? "pb-0" : "pb-4"} min-w-0`}>
        <p
          className={`text-sm font-mono font-medium leading-6 ${
            event.status === "error"
              ? "text-accent-red"
              : event.status === "active"
                ? "text-accent-green"
                : event.status === "muted"
                  ? "text-text-muted"
                  : "text-text-primary"
          }`}
        >
          {event.title}
        </p>
        {event.detail && (
          <p className="text-xs font-mono text-text-muted mt-0.5">
            {event.detail}
          </p>
        )}
      </div>
    </div>
  );
}

function SessionTypeBadge({ types }: { types: string[] }) {
  if (types.includes("subsystem"))
    return (
      <span className="inline-flex items-center px-2 py-0.5 text-2xs font-mono font-semibold rounded-md border bg-accent-cyan/10 text-accent-cyan border-accent-cyan/20">
        sftp
      </span>
    );
  if (types.includes("exec"))
    return (
      <span className="inline-flex items-center px-2 py-0.5 text-2xs font-mono font-semibold rounded-md border bg-accent-yellow/10 text-accent-yellow border-accent-yellow/20">
        exec
      </span>
    );
  if (types.includes("shell") || types.includes("pty-req"))
    return (
      <span className="inline-flex items-center px-2 py-0.5 text-2xs font-mono font-semibold rounded-md border bg-primary/10 text-primary border-primary/20">
        shell
      </span>
    );
  return null;
}

function DurationStat({
  startedAt,
  lastSeen,
  active,
}: {
  startedAt: string;
  lastSeen: string;
  active: boolean;
}) {
  const duration = formatDuration(startedAt, lastSeen, active);
  return (
    <span className="inline-flex items-center gap-1 px-2 py-0.5 text-2xs font-mono font-semibold rounded-md border bg-surface text-text-secondary border-border tabular-nums">
      <ClockIcon className="w-3 h-3 text-text-muted" strokeWidth={2} />
      {duration}
    </span>
  );
}

/* ── page ────────────────────────────────────────── */

export default function SessionDetails() {
  const { uid } = useParams<{ uid: string }>();
  const navigate = useNavigate();
  const { session, loading, error, fetchOne, close } = useSessionsStore();
  const [showClose, setShowClose] = useState(false);
  const [closing, setClosing] = useState(false);
  const [closeError, setCloseError] = useState<string | null>(null);

  useEffect(() => {
    if (uid) fetchOne(uid);
  }, [uid, fetchOne]);

  const handleClose = async () => {
    if (!uid) return;
    setClosing(true);
    setCloseError(null);
    try {
      await close(uid);
      setShowClose(false);
    } catch {
      setCloseError("Failed to close session. Check your permissions.");
    } finally {
      setClosing(false);
    }
  };

  if (loading || !session) {
    return (
      <div className="flex items-center justify-center py-24">
        <span className="w-5 h-5 border-2 border-primary/30 border-t-primary rounded-full animate-spin" />
      </div>
    );
  }

  if (error) {
    return (
      <div className="flex flex-col items-center justify-center py-24 gap-3">
        <XCircleIcon className="w-8 h-8 text-accent-red/60" />
        <p className="text-sm font-mono text-text-muted">{error}</p>
        <button
          onClick={() => navigate("/sessions")}
          className="text-xs font-mono text-primary hover:underline"
        >
          ← Back to sessions
        </button>
      </div>
    );
  }

  const timeline = buildTimeline(session);

  return (
    <div className="animate-fade-in">
      {/* Breadcrumb */}
      <div className="flex items-center gap-1.5 mb-6">
        <Link
          to="/sessions"
          className="text-2xs font-mono text-text-muted hover:text-primary transition-colors"
        >
          Sessions
        </Link>
        <ChevronRightIcon
          className="w-3 h-3 text-text-muted/40"
          strokeWidth={2}
        />
        <span className="text-2xs font-mono text-text-secondary truncate max-w-[28ch]">
          {session.uid}
        </span>
      </div>

      {/* Body */}
      <div className="grid grid-cols-1 lg:grid-cols-3 gap-5">
        {/* Timeline */}
        <div className="lg:col-span-2 bg-card border border-border rounded-xl p-5">
          <div className="flex items-center justify-between mb-5">
            <h3 className={LABEL}>Session flow</h3>
            <div className="flex items-center gap-2">
              {!session.authenticated && (
                <span className="inline-flex items-center gap-1 px-2 py-0.5 text-2xs font-mono font-semibold rounded-md bg-accent-red/10 text-accent-red border border-accent-red/20">
                  <ShieldExclamationIcon className="w-3 h-3" strokeWidth={2} />
                  not authenticated
                </span>
              )}
              <SessionTypeBadge types={session.events?.types ?? []} />
              <span
                className={`inline-flex items-center gap-1 px-2 py-0.5 text-2xs font-mono font-semibold rounded-md border ${
                  session.active
                    ? "bg-accent-green/10 text-accent-green border-accent-green/20"
                    : "bg-surface text-text-muted border-border"
                }`}
              >
                <span
                  className={`w-1.5 h-1.5 rounded-full ${session.active ? "bg-accent-green" : "bg-text-muted/50"}`}
                />
                {session.active ? "active" : "closed"}
              </span>
              <DurationStat
                startedAt={session.started_at}
                lastSeen={session.last_seen}
                active={session.active}
              />
              {session.active && (
                <button
                  onClick={() => setShowClose(true)}
                  className="flex items-center gap-1.5 px-2.5 py-1 border border-accent-red/30 text-accent-red hover:bg-accent-red/10 rounded-md text-2xs font-mono font-medium transition-all"
                >
                  <XCircleIcon className="w-3 h-3" strokeWidth={2} />
                  close
                </button>
              )}
            </div>
          </div>
          {timeline.map((event, i) => (
            <TimelineNode
              key={event.id}
              event={event}
              isLast={i === timeline.length - 1}
            />
          ))}
        </div>

        {/* Right column */}
        <div className="space-y-5">
          {/* Session meta */}
          <div className="bg-card border border-border rounded-xl p-5 space-y-3.5">
            <h3 className={`${LABEL} mb-4`}>Details</h3>
            <div>
              <dt className={LABEL}>Session UID</dt>
              <dd className="flex items-center gap-1 mt-1">
                <code className="text-2xs font-mono text-text-muted truncate">
                  {session.uid}
                </code>
                <CopyButton text={session.uid} />
              </dd>
            </div>
            <div>
              <dt className={LABEL}>From</dt>
              <dd className="flex items-center gap-1 mt-1">
                <code className="text-xs font-mono text-text-secondary">
                  {session.ip_address}
                </code>
                <CopyButton text={session.ip_address} />
              </dd>
            </div>
            <div>
              <dt className={LABEL}>Started</dt>
              <dd className="text-xs font-mono text-text-secondary mt-1">
                {formatDateFull(session.started_at)}
              </dd>
            </div>
            {!session.active && (
              <div>
                <dt className={LABEL}>Ended</dt>
                <dd className="flex items-center gap-1.5 mt-1 flex-wrap">
                  <span className="text-xs font-mono text-text-secondary">
                    {formatRelative(session.last_seen)}
                  </span>
                  <span className="text-2xs text-text-muted">
                    {formatDateFull(session.last_seen)}
                  </span>
                </dd>
              </div>
            )}
          </div>

          {/* Device */}
          {session.device?.uid && (
            <div className="bg-card border border-border rounded-xl p-5">
              <div className="flex items-center justify-between mb-4">
                <h3 className={LABEL}>Device</h3>
                <DeviceChip
                  uid={session.device.uid}
                  name={
                    session.device.name ?? session.device_uid.substring(0, 8)
                  }
                  online={session.device.online}
                  osId={session.device.info?.id}
                />
              </div>
              {session.device.info && (
                <dl className="space-y-3">
                  <div>
                    <dt className={LABEL}>OS</dt>
                    <dd className="flex items-center gap-1.5 mt-1">
                      <DistroIcon
                        id={session.device.info.id}
                        className="text-[0.85rem] leading-none text-text-muted shrink-0"
                      />
                      <span className="text-xs text-text-primary">
                        {session.device.info.pretty_name}
                      </span>
                    </dd>
                  </div>
                  <div>
                    <dt className={LABEL}>Architecture</dt>
                    <dd className="mt-1">
                      <code className="text-xs font-mono text-text-secondary">
                        {session.device.info.arch}
                      </code>
                    </dd>
                  </div>
                  <div>
                    <dt className={LABEL}>Agent Version</dt>
                    <dd className="mt-1">
                      <code className="text-xs font-mono text-text-secondary">
                        {session.device.info.version}
                      </code>
                    </dd>
                  </div>
                </dl>
              )}
            </div>
          )}
        </div>
      </div>

      {/* Close Session Dialog */}
      {showClose && (
        <div className="fixed inset-0 z-[70] flex items-center justify-center">
          <div
            className="absolute inset-0 bg-black/60 backdrop-blur-sm"
            onClick={() => setShowClose(false)}
          />
          <div className="relative bg-surface border border-border rounded-2xl w-full max-w-sm mx-4 p-6 shadow-2xl animate-slide-up">
            <div className="flex items-center gap-3 mb-4">
              <div className="w-10 h-10 rounded-full bg-accent-red/10 border border-accent-red/20 flex items-center justify-center shrink-0">
                <XCircleIcon
                  className="w-5 h-5 text-accent-red"
                  strokeWidth={2}
                />
              </div>
              <div>
                <h2 className="text-base font-semibold text-text-primary">
                  Close Session
                </h2>
                <p className="text-xs text-text-muted mt-0.5">
                  This will terminate the SSH connection
                </p>
              </div>
            </div>
            <p className="text-sm text-text-muted mb-6">
              Are you sure you want to close the session for{" "}
              <span className="font-medium text-text-primary">
                {session.username}
              </span>{" "}
              on{" "}
              <span className="font-medium text-text-primary">
                {session.device?.name ?? session.device_uid.substring(0, 8)}
              </span>
              ?
            </p>
            {closeError && (
              <p className="text-xs text-accent-red mb-4 font-mono">
                {closeError}
              </p>
            )}
            <div className="flex justify-end gap-2">
              <button
                onClick={() => setShowClose(false)}
                className="px-4 py-2.5 text-sm font-medium text-text-secondary hover:text-text-primary rounded-lg hover:bg-hover-subtle transition-colors"
              >
                Cancel
              </button>
              <button
                onClick={handleClose}
                disabled={closing}
                className="px-5 py-2.5 bg-accent-red/90 hover:bg-accent-red text-white rounded-lg text-sm font-semibold disabled:opacity-dim transition-all"
              >
                {closing ? "Closing…" : "Close Session"}
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
