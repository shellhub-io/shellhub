import { ArrowPathIcon, PlusCircleIcon } from "@heroicons/react/24/outline";
import { type InstallKeyEvent } from "@/client";
import { formatDateFull } from "@/utils/date";
import { type Column } from "@/components/common/DataTable";
import DistroIcon from "@/components/common/DistroIcon";
import type { EntityBase, EntityOperation } from "@/hooks/useActionDialog";
import EventPublicKey from "./EventPublicKey";
import InstallKeyEventReview from "./InstallKeyEventReview";
import KeyValueChip from "@/components/common/KeyValueChip";
import StatusChip from "./StatusChip";

/**
 * Asks for a decision on an enrolment. Narrowed to accept and reject: an enrolment event cannot
 * be removed, only decided.
 */
export type RequestDeviceAction = (
  entity: EntityBase,
  operation: Extract<EntityOperation, "accept" | "reject">,
) => void;

/**
 * The columns of the enrolment history table. Built as a function rather than a constant because
 * the action column has to close over the caller's handler.
 */
export function getInstallKeyEventColumns(
  requestAction: RequestDeviceAction,
): Column<InstallKeyEvent>[] {
  return [
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
      render: (event) => (
        <div className="space-y-1.5">
          {event.re_registration ? (
            <StatusChip
              icon={ArrowPathIcon}
              label="Re-registered"
              tone="muted"
            />
          ) : (
            <StatusChip icon={PlusCircleIcon} label="New" tone="muted" />
          )}
          <div className="text-2xs font-mono text-text-muted whitespace-nowrap">
            {formatDateFull(event.timestamp)}
          </div>
        </div>
      ),
    },
    {
      key: "review",
      header: "Review",
      render: (event) => (
        <InstallKeyEventReview event={event} requestAction={requestAction} />
      ),
    },
  ];
}
