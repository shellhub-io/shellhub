import { ArrowPathIcon, PlusCircleIcon } from "@heroicons/react/24/outline";
import { format } from "date-fns";
import { type InstallKeyEvent } from "@/client";
import { type Column } from "@/components/common/DataTable";
import DistroIcon from "@/components/common/DistroIcon";
import EventPublicKey from "./EventPublicKey";
import KeyValueChip from "./KeyValueChip";
import InstallKeyEventReview from "./InstallKeyEventReview";
import StatusChip from "./StatusChip";

/**
 * Columns for an install key's registration-activity table: Device (the distro-logo-led identity, its
 * OS, and its facts as chips), Registration, and Review (the verdict plus the current device's live
 * accept/reject controls). The old System column is folded into Device.
 */
export const installKeyEventColumns: Column<InstallKeyEvent>[] = [
  {
    key: "device",
    header: "Device",
    render: (event) => (
      <div className="flex items-center gap-2.5">
        <DistroIcon
          id={event.info?.id ?? ""}
          className="text-[1.05rem] leading-none text-text-secondary shrink-0"
        />
        <div className="min-w-0">
          <div className="text-sm font-medium text-text-primary">
            {event.hostname}
          </div>
          {event.info?.pretty_name && (
            <div className="mt-0.5 text-2xs text-text-muted">
              {event.info.pretty_name}
            </div>
          )}
          <div className="mt-2 flex flex-wrap items-center gap-1.5">
            {event.mac && <KeyValueChip label="MAC" value={event.mac} />}
            {event.source_ip && (
              <KeyValueChip label="IP" value={event.source_ip} />
            )}
            <EventPublicKey event={event} />
            {event.info?.arch && (
              <KeyValueChip label="Arch" value={event.info.arch} />
            )}
            {event.info?.version && (
              <KeyValueChip label="Agent" value={event.info.version} />
            )}
          </div>
        </div>
      </div>
    ),
  },
  {
    key: "registration",
    header: "Registration",
    // The kind of registration and when it happened — two facets of the same event, merged into one
    // column so the near-identical "Registration"/"Registered" headers don't sit side by side.
    render: (event) => (
      <div className="space-y-1.5">
        {event.re_registration ? (
          <StatusChip icon={ArrowPathIcon} label="Re-registered" tone="muted" />
        ) : (
          // A credential never seen before (a re-key is a new key here, so it reads as New). The
          // registration kind is informational, not a status, so it stays muted like Re-registered —
          // only the icon distinguishes the two.
          <StatusChip icon={PlusCircleIcon} label="New" tone="muted" />
        )}
        <div className="text-2xs font-mono text-text-muted whitespace-nowrap">
          {format(new Date(event.timestamp), "MMM d, yyyy HH:mm")}
        </div>
      </div>
    ),
  },
  {
    key: "review",
    header: "Review",
    // The verdict and, for the current device, its live control right under it: the frozen accept/reject
    // decision with its timestamp, a Pending badge with Accept/Reject links, or a dash.
    render: (event) => <InstallKeyEventReview event={event} />,
  },
];
