import { cn } from "@shellhub/design-system/cn";
import type { Column } from "@/components/common/DataTable";
import DeviceChip from "@/components/common/DeviceChip";
import EmptyCell from "@/components/common/EmptyCell";
import SessionLogin from "@/components/sessions/SessionLogin";
import SessionTypeBadge from "@/components/sessions/SessionTypeBadge";
import { formatRelative } from "@/utils/date";
import type { Session } from "@/client";

/**
 * The columns every session table shows, picked by name so a table can place its own columns
 * between them. The device column's chip opens the device and stops its click there, so a table
 * whose rows open the session does not open it too.
 */
export const sessionColumns = {
  active: {
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
  type: {
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
  device: {
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
  username: {
    key: "username",
    header: "Login",
    render: (s) => <SessionLogin session={s} />,
  },
  started: {
    key: "started",
    header: "Started",
    render: (s) => (
      <span className="text-xs text-text-secondary">
        {formatRelative(s.started_at)}
      </span>
    ),
  },
} satisfies Record<string, Column<Session>>;

/**
 * The row class every session table gives a session: a red edge and tint on one that failed to
 * authenticate, and a transparent edge on the rest so the rows stay aligned.
 */
export const sessionRowClassName = (s: Session) =>
  !s.authenticated
    ? "bg-accent-red/[0.03] hover:bg-accent-red/[0.06] border-l-2 border-l-accent-red/50"
    : "border-l-2 border-l-transparent";
