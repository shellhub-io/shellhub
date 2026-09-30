import { Fragment, useState, type ReactNode } from "react";
import { Link } from "react-router-dom";
import { ChevronRightIcon } from "@heroicons/react/24/outline";
import { format, isToday, isYesterday } from "date-fns";
import { Button, Callout } from "@shellhub/design-system/primitives";
import { cn } from "@shellhub/design-system/cn";
import { type ProvisioningKeyEvent } from "@/client";
import ActionDialog from "@/components/common/ActionDialog";
import DistroIcon from "@/components/common/DistroIcon";
import RestrictedAction from "@/components/common/RestrictedAction";
import { useActionDialog, type EntityBase } from "@/hooks/useActionDialog";
import { useDeviceActionRunner } from "@/hooks/useDeviceActionRunner";
import { useInvalidateByIds } from "@/hooks/useInvalidateQueries";
import { useLoadOnReach } from "@/hooks/useLoadOnReach";
import { useProvisioningKeyEvents } from "@/hooks/useProvisioningKeyEvents";
import { formatDateFull } from "@/utils/date";
import EventPublicKey from "./EventPublicKey";
import { formatCount } from "./helpers";

type Outcome = "waiting" | "accepted" | "rejected" | "settled";

type Decide = (entity: EntityBase, operation: "accept" | "reject") => void;

function outcomeOf(event: ProvisioningKeyEvent): Outcome {
  if (event.is_current && event.device_status === "pending") return "waiting";
  const decided = event.decided_status;
  if (decided === "accepted" || decided === "rejected") return decided;
  if (!event.is_current) return "settled";
  const status = event.device_status;
  if (status === "accepted" || status === "rejected") return status;
  return "settled";
}

function dayLabel(timestamp: string): string {
  const date = new Date(timestamp);
  if (isToday(date)) return "Today";
  if (isYesterday(date)) return "Yesterday";
  return format(date, "MMM d, yyyy");
}

const DOT: Record<Outcome, string> = {
  waiting: "bg-accent-yellow ring-4 ring-accent-yellow/20",
  accepted: "bg-accent-green",
  rejected: "bg-accent-red",
  settled: "bg-text-muted/40",
};

const VERDICT: Record<Outcome, { label: string; className: string }> = {
  waiting: { label: "", className: "" },
  accepted: { label: "accepted", className: "text-accent-green" },
  rejected: { label: "rejected", className: "text-accent-red" },
  settled: { label: "", className: "" },
};

function Dot({ outcome, top }: { outcome: Outcome; top: string }) {
  return (
    <span
      aria-hidden="true"
      className={cn(
        "absolute left-[5px] w-2.5 h-2.5 rounded-full",
        top,
        DOT[outcome],
      )}
    />
  );
}

function Headline({ event }: { event: ProvisioningKeyEvent }) {
  return (
    <>
      <time
        dateTime={event.timestamp}
        className="text-2xs font-mono text-text-muted w-10 shrink-0"
      >
        {format(new Date(event.timestamp), "HH:mm")}
      </time>
      <Link
        to={`/devices/${event.device_uid}`}
        className="text-sm font-medium text-text-primary hover:text-primary hover:underline truncate"
      >
        {event.hostname}
      </Link>
      {event.re_registration && (
        <span className="text-xs text-text-muted">· re-registered</span>
      )}
    </>
  );
}

function deviceOf(event: ProvisioningKeyEvent): EntityBase {
  return { uid: event.device_uid, name: event.hostname };
}

function useEventDetails(event: ProvisioningKeyEvent, toggleClassName: string) {
  const [open, setOpen] = useState(false);
  return {
    toggle: (
      <DetailsToggle
        open={open}
        onToggle={() => setOpen(!open)}
        label={event.info?.pretty_name ?? ""}
        hostname={event.hostname}
        className={toggleClassName}
      />
    ),
    details: open ? <Details event={event} /> : null,
  };
}

function WaitingEntry({
  event,
  decide,
}: {
  event: ProvisioningKeyEvent;
  decide: Decide;
}) {
  const { toggle, details } = useEventDetails(event, "text-text-muted");
  const entity = deviceOf(event);

  return (
    <li className="relative pl-8 py-2">
      <Dot outcome="waiting" top="top-[22px]" />
      <div className="rounded-xl border border-accent-yellow/25 bg-accent-yellow/[0.05] px-4 py-3">
        <div className="flex flex-wrap items-center gap-x-2 gap-y-2">
          <Headline event={event} />
          <span className="ml-auto flex items-center gap-1.5">
            <RestrictedAction action="device:accept">
              <Button
                variant="successSoft"
                size="sm"
                onClick={() => decide(entity, "accept")}
              >
                Accept
              </Button>
            </RestrictedAction>
            <RestrictedAction action="device:reject">
              <Button
                variant="ghost"
                size="sm"
                onClick={() => decide(entity, "reject")}
              >
                Reject
              </Button>
            </RestrictedAction>
          </span>
        </div>
        <div className="mt-1.5 ml-12 flex flex-wrap items-center gap-x-3 gap-y-1.5">
          <span className="text-2xs font-medium text-accent-yellow">
            Waiting for you
          </span>
          <DistroIcon
            id={event.info?.id ?? ""}
            className="text-sm leading-none text-text-secondary"
          />
          {toggle}
        </div>
        {details}
      </div>
    </li>
  );
}

function DetailsToggle({
  open,
  onToggle,
  label,
  hostname,
  className,
}: {
  open: boolean;
  onToggle: () => void;
  label: string;
  hostname: string;
  className: string;
}) {
  return (
    <button
      type="button"
      onClick={onToggle}
      aria-expanded={open}
      aria-label={`${open ? "Hide" : "Show"} details for ${hostname}`}
      className={cn(
        "inline-flex items-center gap-1 text-2xs hover:text-text-secondary min-w-0",
        className,
      )}
    >
      <span className="hidden md:inline truncate">{label}</span>
      <ChevronRightIcon
        aria-hidden="true"
        className={cn(
          "w-3 h-3 shrink-0 transition-transform",
          open && "rotate-90",
        )}
        strokeWidth={2}
      />
    </button>
  );
}

function Details({ event }: { event: ProvisioningKeyEvent }) {
  const rows: [string, ReactNode][] = [
    ["Registered", formatDateFull(event.timestamp)],
  ];
  if (event.decided_at)
    rows.push(["Decided", formatDateFull(event.decided_at)]);
  if (event.identity) rows.push(["Identity", event.identity]);
  if (event.source_ip) rows.push(["IP", event.source_ip]);
  rows.push([
    "System",
    [event.info?.pretty_name, event.info?.arch].filter(Boolean).join(" · ") ||
      "—",
  ]);
  if (event.info?.version) rows.push(["Agent", event.info.version]);

  return (
    <dl className="mt-2 ml-12 grid grid-cols-[auto_1fr] gap-x-6 gap-y-1.5 rounded-lg border border-border bg-card/40 px-4 py-3 text-2xs">
      {rows.map(([label, value]) => (
        <Fragment key={label}>
          <dt className="text-text-muted">{label}</dt>
          <dd className="font-mono text-text-secondary min-w-0 truncate">
            {value}
          </dd>
        </Fragment>
      ))}
      {event.fingerprint && (
        <>
          <dt className="text-text-muted">Device key</dt>
          <dd className="min-w-0">
            <EventPublicKey event={event} />
          </dd>
        </>
      )}
    </dl>
  );
}

function Entry({
  event,
  decide,
}: {
  event: ProvisioningKeyEvent;
  decide: Decide;
}) {
  const outcome = outcomeOf(event);
  return outcome === "waiting" ? (
    <WaitingEntry event={event} decide={decide} />
  ) : (
    <ResolvedEntry event={event} outcome={outcome} decide={decide} />
  );
}

function ResolvedEntry({
  event,
  outcome,
  decide,
}: {
  event: ProvisioningKeyEvent;
  outcome: Outcome;
  decide: Decide;
}) {
  const { toggle, details } = useEventDetails(event, "text-text-muted/80");
  const verdict = VERDICT[outcome];
  const reconsider = event.is_current && outcome === "rejected";

  return (
    <li className="relative pl-8 py-2">
      <Dot outcome={outcome} top="top-[15px]" />
      <div className="flex flex-wrap items-baseline gap-x-2 gap-y-0.5">
        <Headline event={event} />
        {verdict.label && (
          <span className={cn("text-xs", verdict.className)}>
            · {verdict.label}
          </span>
        )}
        {reconsider && (
          <span className="text-2xs font-medium">
            <RestrictedAction action="device:accept">
              <button
                type="button"
                onClick={() => decide(deviceOf(event), "accept")}
                className="text-primary hover:underline"
              >
                Accept instead
              </button>
            </RestrictedAction>
          </span>
        )}
        {toggle}
      </div>
      {details}
    </li>
  );
}

/**
 * What a provisioning key registered, newest first and grouped by day, one line per registration
 * with how it ended. A device still waiting for a decision stands out as a card in the line, with
 * what it reported and its accept and reject; any other line expands to the same details. Older
 * registrations load a hundred at a time as the reader nears the end; a page that fails keeps the
 * ones already shown and offers a retry.
 */
export default function ProvisioningKeyTimeline({ id }: { id: string }) {
  const refreshHistory = useInvalidateByIds("provisioningKeyHistory");
  const deviceActions = useActionDialog({
    onSuccess: () => void refreshHistory(),
  });
  const runDeviceAction = useDeviceActionRunner();
  const {
    events,
    totalCount,
    hasMore,
    loadMore,
    isLoadingMore,
    loadMoreFailed,
    isLoading,
    error,
  } = useProvisioningKeyEvents({ id });
  const sentinel = useLoadOnReach(
    loadMore,
    hasMore && !isLoadingMore && !loadMoreFailed,
  );

  if (error && events.length === 0) {
    return (
      <Callout variant="error">
        Could not load registration activity. Check your connection and try
        again.
      </Callout>
    );
  }

  if (isLoading) {
    return (
      <p className="py-12 text-center text-xs font-mono text-text-muted">
        Loading activity…
      </p>
    );
  }

  if (events.length === 0) {
    return (
      <p className="py-12 text-center text-xs text-text-muted">
        No registrations yet. Devices that register with this key will appear
        here.
      </p>
    );
  }

  const days: { label: string; events: ProvisioningKeyEvent[] }[] = [];
  for (const event of events) {
    const label = dayLabel(event.timestamp);
    const last = days[days.length - 1];
    if (last?.label === label) last.events.push(event);
    else days.push({ label, events: [event] });
  }

  return (
    <section aria-label="Registration activity" className="space-y-8">
      {days.map((day) => (
        <div key={day.label}>
          <h3 className="mb-2 text-2xs font-mono font-semibold uppercase tracking-label text-text-muted">
            {day.label}
          </h3>
          <ol className="relative before:absolute before:left-[9px] before:top-4 before:bottom-4 before:w-0.5 before:rounded-full before:bg-text-muted/25">
            {day.events.map((event) => (
              <Entry
                key={event.id}
                event={event}
                decide={deviceActions.requestAction}
              />
            ))}
          </ol>
        </div>
      ))}

      <div className="flex items-center justify-between gap-4 pl-8">
        <span className="text-2xs font-mono text-text-muted">
          {formatCount(events.length)} of {formatCount(totalCount)}{" "}
          registrations
        </span>
        {isLoadingMore && (
          <span className="text-2xs font-mono text-text-muted">
            Loading more…
          </span>
        )}
        {loadMoreFailed && !isLoadingMore && (
          <span role="alert" className="flex items-center gap-2 text-2xs">
            <span className="text-accent-red">Could not load more.</span>
            <button
              type="button"
              onClick={loadMore}
              className="font-medium text-primary hover:underline"
            >
              Try again
            </button>
          </span>
        )}
      </div>
      <div ref={sentinel} aria-hidden="true" />

      {deviceActions.action && (
        <ActionDialog
          key={deviceActions.actionKey}
          action={deviceActions.action}
          onClose={deviceActions.close}
          onSuccess={deviceActions.handleSuccess}
          entityType="device"
          runAction={runDeviceAction}
        />
      )}
    </section>
  );
}
