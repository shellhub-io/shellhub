import { useState } from "react";
import { useNavigate } from "react-router-dom";
import {
  CommandLineIcon,
  ExclamationCircleIcon,
  ExclamationTriangleIcon,
  ShieldCheckIcon,
  ShieldExclamationIcon,
} from "@heroicons/react/24/outline";
import { useAdminSessionsList } from "../../hooks/useAdminSessionsList";
import PageHeader from "../../components/common/PageHeader";
import Pagination from "../../components/common/Pagination";
import DeviceChip from "../../components/common/DeviceChip";
import { formatDateFull } from "../../utils/date";
import { TH } from "../../utils/styles";

const PER_PAGE = 10;
const COL_SPAN = 8;

export default function AdminSessions() {
  const [page, setPage] = useState(1);
  const { sessions, totalCount, isLoading, error } = useAdminSessionsList(page, PER_PAGE);
  const navigate = useNavigate();

  const totalPages = Math.ceil(totalCount / PER_PAGE);

  return (
    <div>
      <PageHeader
        icon={<CommandLineIcon className="w-6 h-6" />}
        overline="Admin"
        title="Sessions"
        description="Track live and historical sessions happening across every namespace."
      />

      {error && (
        <div role="alert" className="flex items-center gap-2 bg-accent-red/8 border border-accent-red/20 text-accent-red px-3.5 py-2.5 rounded-md text-xs font-mono mb-4 animate-slide-down">
          <ExclamationCircleIcon className="w-3.5 h-3.5 shrink-0" strokeWidth={2} />
          {error.message}
        </div>
      )}

      <div className="bg-card border border-border rounded-lg overflow-hidden animate-fade-in">
        <table className="w-full">
          <thead>
            <tr className="border-b border-border bg-surface/50">
              <th className={`${TH} w-14`}>Active</th>
              <th className={TH}>ID</th>
              <th className={TH}>Device</th>
              <th className={TH}>Username</th>
              <th className={`${TH} w-14`}>Auth</th>
              <th className={TH}>IP Address</th>
              <th className={TH}>Started</th>
              <th className={TH}>Last Seen</th>
            </tr>
          </thead>
          <tbody className="divide-y divide-border/60">
            {isLoading
              ? (
                <tr>
                  <td colSpan={COL_SPAN} className="px-4 py-12 text-center">
                    <div className="flex items-center justify-center gap-3">
                      <span className="w-4 h-4 border-2 border-primary/30 border-t-primary rounded-full animate-spin" />
                      <span className="text-xs font-mono text-text-muted">Loading sessions…</span>
                    </div>
                  </td>
                </tr>
              )
              : sessions.length === 0
                ? (
                  <tr>
                    <td colSpan={COL_SPAN} className="px-4 py-12 text-center">
                      <p className="text-xs font-mono text-text-muted">No sessions found</p>
                    </td>
                  </tr>
                )
                : sessions.map((session) => {
                  const suspicious = !session.authenticated;
                  return (
                    <tr
                      key={session.uid}
                      onClick={() => void navigate(`/admin/sessions/${session.uid}`)}
                      className={`transition-colors cursor-pointer ${
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
                        <code
                          className="text-xs font-mono text-text-muted bg-surface px-1.5 py-0.5 rounded"
                          title={session.uid}
                        >
                          {session.uid.substring(0, 10)}
                        </code>
                      </td>
                      <td className="px-4 py-3.5">
                        {session.device
                          ? (
                            <DeviceChip
                              disableLink
                              name={session.device.name}
                              online={session.device.online}
                              osId={session.device.info?.id}
                            />
                          )
                          : (
                            <span className="text-xs font-mono text-text-primary">
                              {(session.device_uid ?? "").substring(0, 8)}
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
                              suspicious ? "text-accent-red/60" : "text-text-secondary"
                            }`}
                          >
                            {session.username}
                          </code>
                        </div>
                      </td>
                      <td className="px-4 py-3.5">
                        {session.authenticated
                          ? (
                            <ShieldCheckIcon
                              className="w-4 h-4 text-accent-green"
                              strokeWidth={2}
                              title="Authenticated"
                            />
                          )
                          : (
                            <ShieldExclamationIcon
                              className="w-4 h-4 text-accent-red"
                              strokeWidth={2}
                              title="Not authenticated"
                            />
                          )}
                      </td>
                      <td className="px-4 py-3.5">
                        <code className="text-xs font-mono text-text-muted bg-surface px-1.5 py-0.5 rounded">
                          {session.ip_address}
                        </code>
                      </td>
                      <td className="px-4 py-3.5">
                        <span className="text-xs text-text-secondary">
                          {formatDateFull(session.started_at)}
                        </span>
                      </td>
                      <td className="px-4 py-3.5">
                        <span className="text-xs text-text-secondary">
                          {formatDateFull(session.last_seen)}
                        </span>
                      </td>
                    </tr>
                  );
                })}
          </tbody>
        </table>
      </div>

      <Pagination
        page={page}
        totalPages={totalPages}
        totalCount={totalCount}
        itemLabel="session"
        onPageChange={setPage}
      />
    </div>
  );
}
