import {
  CheckCircleIcon,
  ClockIcon,
  XCircleIcon,
} from "@heroicons/react/24/outline";
import { format } from "date-fns";
import RestrictedAction from "@/components/common/RestrictedAction";
import { type InstallKeyEvent } from "@/client";
import type { RequestDeviceAction } from "./installKeyEventColumns";
import StatusChip from "./StatusChip";

type ReviewVerdict = "accepted" | "rejected";

function Verdict({
  status,
  at,
}: {
  status: ReviewVerdict;
  at?: string | null;
}) {
  const rejected = status === "rejected";

  return (
    <div className="space-y-1.5">
      <StatusChip
        icon={rejected ? XCircleIcon : CheckCircleIcon}
        label={rejected ? "Rejected" : "Accepted"}
        tone={rejected ? "red" : "green"}
      />
      {at && (
        <div className="text-2xs font-mono text-text-muted whitespace-nowrap">
          {format(new Date(at), "MMM d, yyyy HH:mm")}
        </div>
      )}
    </div>
  );
}

function ActionLink({
  children,
  color,
  onClick,
}: {
  children: string;
  color: "green" | "red";
  onClick: () => void;
}) {
  return (
    <button
      type="button"
      onClick={onClick}
      className={`font-medium hover:underline ${
        color === "green" ? "text-accent-green" : "text-accent-red"
      }`}
    >
      {children}
    </button>
  );
}

// decided_status is per-event (survives device removal); device_status is the live fallback for pre-stamping enrollments.
function resolveVerdict(event: InstallKeyEvent): ReviewVerdict | undefined {
  const decided = event.decided_status;
  if (decided === "accepted" || decided === "rejected") return decided;

  const status = event.device_status;
  if (event.is_current && (status === "accepted" || status === "rejected")) {
    return status;
  }
  return undefined;
}

export default function InstallKeyEventReview({
  event,
  requestAction,
}: {
  event: InstallKeyEvent;
  requestAction: RequestDeviceAction;
}) {
  const status = event.device_status;
  const actionable = event.is_current;
  const entity = { uid: event.device_uid, name: event.hostname };

  const acceptLink = (
    <RestrictedAction action="device:accept">
      <ActionLink color="green" onClick={() => requestAction(entity, "accept")}>
        Accept
      </ActionLink>
    </RestrictedAction>
  );

  if (actionable && status === "pending") {
    return (
      <div className="space-y-1.5">
        <StatusChip icon={ClockIcon} label="Pending" tone="yellow" />
        <div className="flex items-center gap-2 text-2xs">
          {acceptLink}
          <span className="text-text-muted">·</span>
          <RestrictedAction action="device:reject">
            <ActionLink
              color="red"
              onClick={() => requestAction(entity, "reject")}
            >
              Reject
            </ActionLink>
          </RestrictedAction>
        </div>
      </div>
    );
  }

  if (actionable && status === "rejected") {
    return (
      <div className="space-y-1">
        <Verdict status="rejected" at={event.decided_at} />
        <div className="flex items-center gap-2 text-2xs">{acceptLink}</div>
      </div>
    );
  }

  const verdict = resolveVerdict(event);
  if (verdict) {
    return <Verdict status={verdict} at={event.decided_at} />;
  }

  return <span className="text-2xs text-text-muted">—</span>;
}
