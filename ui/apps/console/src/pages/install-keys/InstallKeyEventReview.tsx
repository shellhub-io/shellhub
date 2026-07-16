import { useState } from "react";
import {
  CheckCircleIcon,
  ClockIcon,
  XCircleIcon,
} from "@heroicons/react/24/outline";
import { format } from "date-fns";
import { useAcceptDevice, useRejectDevice } from "@/hooks/useDeviceMutations";
import { useInvalidateByIds } from "@/hooks/useInvalidateQueries";
import RestrictedAction from "@/components/common/RestrictedAction";
import { type DeviceStatus, type InstallKeyEvent } from "@/client";
import StatusChip from "./StatusChip";

/**
 * The frozen verdict: a soft status chip (accepted green / rejected red) over when it was decided —
 * chip-over-timestamp, matching the Registration column's layout.
 */
function Verdict({ status, at }: { status: DeviceStatus; at?: string | null }) {
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

/** An inline text-link action (accept green / reject red), the review's control under its status. */
function ActionLink({
  children,
  color,
  onClick,
  disabled,
}: {
  children: string;
  color: "green" | "red";
  onClick: () => void;
  disabled: boolean;
}) {
  return (
    <button
      type="button"
      onClick={onClick}
      disabled={disabled}
      className={`font-medium hover:underline disabled:opacity-50 disabled:no-underline ${
        color === "green" ? "text-accent-green" : "text-accent-red"
      }`}
    >
      {children}
    </button>
  );
}

/**
 * Resolve an event's review verdict: the per-event stamped decision (survives the device being
 * removed, so an older re-registration keeps its own), falling back to the current device's live
 * status when the decision predates stamping so a decided device never shows a dash.
 */
function resolveVerdict(event: InstallKeyEvent): DeviceStatus | undefined {
  const status = event.device_status;

  return (
    event.decided_status ??
    (event.is_current && (status === "accepted" || status === "rejected")
      ? status
      : undefined)
  );
}

/**
 * The Review cell: the enrollment's verdict and, for the current device, its live control right under
 * it. A pending device shows a Pending badge with Accept / Reject links; a rejected current device
 * shows the Rejected verdict with a second-chance Accept link; a reviewed row shows its frozen verdict;
 * an older, never-decided row shows a dash. Controls only ever apply to the current device's newest
 * enrollment.
 */
export default function InstallKeyEventReview({
  event,
}: {
  event: InstallKeyEvent;
}) {
  const accept = useAcceptDevice();
  const reject = useRejectDevice();
  const refreshHistory = useInvalidateByIds("installKeyHistory");
  const [error, setError] = useState("");

  const busy = accept.isPending || reject.isPending;
  const status = event.device_status;
  const actionable = event.is_current;

  const run = async (fn: () => Promise<unknown>) => {
    setError("");
    try {
      await fn();
      await refreshHistory();
    } catch {
      setError("Action failed");
    }
  };

  const doAccept = () =>
    run(() => accept.mutateAsync({ path: { uid: event.device_uid } }));
  const doReject = () =>
    run(() =>
      reject.mutateAsync({ path: { uid: event.device_uid, status: "reject" } }),
    );

  const acceptLink = (
    <RestrictedAction action="device:accept">
      <ActionLink color="green" onClick={() => void doAccept()} disabled={busy}>
        Accept
      </ActionLink>
    </RestrictedAction>
  );

  // Current device still awaiting review: Pending badge with the live controls under it.
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
              onClick={() => void doReject()}
              disabled={busy}
            >
              Reject
            </ActionLink>
          </RestrictedAction>
          {error && <span className="text-accent-red">{error}</span>}
        </div>
      </div>
    );
  }

  // Current device that was rejected: the verdict with a second-chance Accept under it.
  if (actionable && status === "rejected") {
    return (
      <div className="space-y-1">
        <Verdict status="rejected" at={event.decided_at} />
        <div className="flex items-center gap-2 text-2xs">
          {acceptLink}
          {error && <span className="text-accent-red">{error}</span>}
        </div>
      </div>
    );
  }

  const verdict = resolveVerdict(event);
  if (verdict) {
    return <Verdict status={verdict} at={event.decided_at} />;
  }

  return <span className="text-2xs text-text-muted">—</span>;
}
