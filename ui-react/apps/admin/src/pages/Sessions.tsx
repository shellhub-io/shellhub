import { useEffect, useState } from "react";
import { useNavigate } from "react-router-dom";
import {
  ExclamationCircleIcon,
  CommandLineIcon,
  ExclamationTriangleIcon,
  XCircleIcon,
} from "@heroicons/react/24/outline";
import { useSessionsStore } from "../stores/sessionsStore";
import PageHeader from "../components/common/PageHeader";
import DeviceChip from "../components/common/DeviceChip";
import { formatDate, formatDuration } from "../utils/date";
import { sessionType } from "../utils/session";
import { TH } from "../utils/styles";
import Pagination from "../components/common/Pagination";

function CloseButton({ onClose }: { onClose: () => Promise<void> }) {
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
      onClick={handleClick}
      disabled={closing}
      title="Close session"
      className="p-1.5 rounded-md text-text-muted hover:text-accent-red hover:bg-accent-red/10 transition-colors disabled:opacity-dim"
    >
      <XCircleIcon className="w-4 h-4" strokeWidth={2} />
    </button>
  );
}

const COL_SPAN = 8;

export default function Sessions() {
  const {
    sessions,
    totalCount,
    loading,
    error,
    page,
    perPage,
    fetch,
    close,
    setPage,
  } = useSessionsStore();
  const navigate = useNavigate();

  useEffect(() => {
    fetch();
  }, [fetch]);

  const totalPages = Math.ceil(totalCount / perPage);

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
          {error}
        </div>
      )}

      <div className="bg-card border border-border rounded-lg overflow-hidden animate-fade-in">
        <table className="w-full">
          <thead>
            <tr className="border-b border-border bg-surface/50">
              <th className={`${TH} w-14`}>Active</th>
              <th className={TH}>Device</th>
              <th className={TH}>Username</th>
              <th className={TH}>IP Address</th>
              <th className={TH}>Type</th>
              <th className={TH}>Started</th>
              <th className={TH}>Duration</th>
              <th className="px-4 py-3 w-16" />
            </tr>
          </thead>
          <tbody className="divide-y divide-border/60">
            {loading ? (
              <tr>
                <td colSpan={COL_SPAN} className="px-4 py-12 text-center">
                  <div className="flex items-center justify-center gap-3">
                    <span className="w-4 h-4 border-2 border-primary/30 border-t-primary rounded-full animate-spin" />
                    <span className="text-xs font-mono text-text-muted">
                      Loading sessions…
                    </span>
                  </div>
                </td>
              </tr>
            ) : sessions.length === 0 ? (
              <tr>
                <td colSpan={COL_SPAN} className="px-4 py-12 text-center">
                  <p className="text-xs font-mono text-text-muted">
                    No sessions found
                  </p>
                </td>
              </tr>
            ) : (
              sessions.map((session) => {
                const type = sessionType(session);
                const suspicious = !session.authenticated;
                return (
                  <tr
                    key={session.uid}
                    onClick={() => navigate(`/sessions/${session.uid}`)}
                    className={`transition-colors group cursor-pointer relative ${
                      suspicious
                        ? "bg-accent-red/[0.03] hover:bg-accent-red/[0.06] border-l-2 border-l-accent-red/50"
                        : "hover:bg-hover-subtle border-l-2 border-l-transparent"
                    }`}
                  >
                    <td className="px-4 py-3.5">
                      <span
                        className={`w-2 h-2 rounded-full inline-block ${
                          session.active
                            ? "bg-accent-green shadow-[0_0_6px_rgba(130,165,104,0.4)]"
                            : "bg-text-muted/40"
                        }`}
                      />
                    </td>
                    <td className="px-4 py-3.5">
                      {session.device?.uid ? (
                        <DeviceChip
                          uid={session.device.uid}
                          name={
                            session.device.name ??
                            session.device_uid.substring(0, 8)
                          }
                          online={session.device.online}
                          osId={session.device.info?.id}
                          onClick={(e) => e.stopPropagation()}
                        />
                      ) : (
                        <span className="text-xs font-mono text-text-primary">
                          {session.device?.name ??
                            session.device_uid.substring(0, 8)}
                        </span>
                      )}
                    </td>
                    <td className="px-4 py-3.5">
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
                            suspicious
                              ? "text-accent-red/60"
                              : "text-text-secondary"
                          }`}
                        >
                          {session.username}
                        </code>
                      </div>
                    </td>
                    <td className="px-4 py-3.5">
                      <code className="text-xs font-mono text-text-muted bg-surface px-1.5 py-0.5 rounded">
                        {session.ip_address}
                      </code>
                    </td>
                    <td className="px-4 py-3.5">
                      {type ? (
                        <span
                          className={`inline-flex items-center px-2 py-0.5 text-2xs font-mono font-semibold rounded border ${type.color}`}
                        >
                          {type.label}
                        </span>
                      ) : (
                        <span className="text-2xs text-text-muted">—</span>
                      )}
                    </td>
                    <td className="px-4 py-3.5">
                      <span className="text-xs text-text-secondary">
                        {formatDate(session.started_at)}
                      </span>
                    </td>
                    <td className="px-4 py-3.5">
                      <span className="text-xs font-mono text-text-secondary tabular-nums">
                        {formatDuration(
                          session.started_at,
                          session.last_seen,
                          session.active,
                        )}
                      </span>
                    </td>
                    <td className="px-4 py-3.5 text-right">
                      {session.active && (
                        <CloseButton onClose={() => close(session.uid)} />
                      )}
                    </td>
                  </tr>
                );
              })
            )}
          </tbody>
        </table>
      </div>

      <Pagination
        page={page}
        totalPages={totalPages}
        totalCount={totalCount}
        itemLabel="session"
        onPageChange={(p) => {
          setPage(p);
          fetch(p);
        }}
      />
    </div>
  );
}
