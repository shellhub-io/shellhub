import { createContext, Fragment, use, useState, type ReactNode } from "react";
import { Link } from "react-router-dom";
import {
  ArrowsUpDownIcon,
  CheckIcon,
  ChevronDownIcon,
  ChevronRightIcon,
  MagnifyingGlassIcon,
} from "@heroicons/react/24/outline";
import { format, isToday, isYesterday } from "date-fns";
import { Button, Callout, Dropdown } from "@shellhub/design-system/primitives";
import { cn } from "@shellhub/design-system/cn";
import { type ProvisioningKeyEvent } from "@/client";
import ActionDialog from "@/components/common/ActionDialog";
import DistroIcon from "@/components/common/DistroIcon";
import SearchField from "@/components/common/fields/SearchField";
import RestrictedAction from "@/components/common/RestrictedAction";
import { useActionDialog, type EntityBase } from "@/hooks/useActionDialog";
import { useDeviceActionRunner } from "@/hooks/useDeviceActionRunner";
import { useInvalidateByIds } from "@/hooks/useInvalidateQueries";
import { useLoadOnReach } from "@/hooks/useLoadOnReach";
import {
  useProvisioningKeyEvents,
  useRegistrationView,
  REGISTRATION_SORTS,
  type RegistrationHistory,
  type RegistrationSort,
} from "@/hooks/useProvisioningKeyEvents";
import { TOOLBAR_CONTROL } from "@/pages/devices/toolbar";
import { formatCount } from "@/utils/count";
import { formatDateFull } from "@/utils/date";
import EventPublicKey from "./EventPublicKey";

type Outcome = "waiting" | "accepted" | "rejected" | "settled" | "refused";

type Decide = (entity: EntityBase, operation: "accept" | "reject") => void;

function outcomeOf(event: ProvisioningKeyEvent): Outcome {
  if (event.refusal) return "refused";
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
  refused: "bg-accent-red/60",
};

const VERDICT: Record<Outcome, { label: string; className: string }> = {
  waiting: { label: "", className: "" },
  accepted: { label: "accepted", className: "text-accent-green" },
  rejected: { label: "rejected", className: "text-accent-red" },
  settled: { label: "", className: "" },
  refused: { label: "refused", className: "text-accent-red" },
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

const SearchedHostname = createContext("");

function MatchedHostname({ hostname }: { hostname: string }) {
  const searched = use(SearchedHostname);
  const at = searched
    ? hostname.toLowerCase().indexOf(searched.toLowerCase())
    : -1;

  if (at < 0) return hostname;

  const end = at + searched.length;

  return (
    <>
      {hostname.slice(0, at)}
      <mark className="rounded-sm bg-primary/15 px-px text-primary">
        {hostname.slice(at, end)}
      </mark>
      {hostname.slice(end)}
    </>
  );
}

function Headline({
  event,
  linked = true,
}: {
  event: ProvisioningKeyEvent;
  linked?: boolean;
}) {
  return (
    <>
      <time
        dateTime={event.timestamp}
        className="text-2xs font-mono text-text-muted w-10 shrink-0"
      >
        {format(new Date(event.timestamp), "HH:mm")}
      </time>
      {!linked ? (
        <span className="text-sm font-medium text-text-primary truncate">
          <MatchedHostname hostname={event.hostname} />
        </span>
      ) : (
        <Link
          to={`/devices/${event.device_uid}`}
          className="text-sm font-medium text-text-primary hover:text-primary hover:underline truncate"
        >
          <MatchedHostname hostname={event.hostname} />
        </Link>
      )}
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
    [
      event.refusal ? "Last attempt" : "Registered",
      formatDateFull(event.timestamp),
    ],
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

function refusalReason(event: ProvisioningKeyEvent): string {
  const agent = event.info?.version
    ? `agent ${event.info.version}`
    : "the agent";
  const times =
    event.attempts === 1 ? "once" : `${formatCount(event.attempts)} times`;
  return `Refused ${times}: ${agent} can't sign its connections.`;
}

function EntryRow({
  event,
  outcome,
  linked = true,
  action,
  note,
}: {
  event: ProvisioningKeyEvent;
  outcome: Outcome;
  linked?: boolean;
  action?: ReactNode;
  note?: ReactNode;
}) {
  const { toggle, details } = useEventDetails(event, "text-text-muted/80");
  const verdict = VERDICT[outcome];

  return (
    <li className="relative pl-8 py-2">
      <Dot outcome={outcome} top="top-[15px]" />
      <div className="flex flex-wrap items-baseline gap-x-2 gap-y-0.5">
        <Headline event={event} linked={linked} />
        {verdict.label && (
          <span className={cn("text-xs", verdict.className)}>
            · {verdict.label}
          </span>
        )}
        {action}
        {toggle}
      </div>
      {note}
      {details}
    </li>
  );
}

function RefusedEntry({ event }: { event: ProvisioningKeyEvent }) {
  return (
    <EntryRow
      event={event}
      outcome="refused"
      linked={false}
      note={
        <p className="mt-0.5 ml-12 text-2xs text-text-muted">
          <span>{refusalReason(event)}</span> Update the agent, or turn off{" "}
          <Link to="/settings/ssh" className="text-primary hover:underline">
            Device signature enforcement
          </Link>
          .
        </p>
      }
    />
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
  if (outcome === "refused") return <RefusedEntry event={event} />;
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
  const reconsider = event.is_current && outcome === "rejected";

  return (
    <EntryRow
      event={event}
      outcome={outcome}
      action={
        reconsider && (
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
        )
      }
    />
  );
}

function RegistrationDays({
  events,
  decide,
}: {
  events: ProvisioningKeyEvent[];
  decide: Decide;
}) {
  const days: { label: string; events: ProvisioningKeyEvent[] }[] = [];
  for (const event of events) {
    const label = dayLabel(event.timestamp);
    const last = days[days.length - 1];
    if (last?.label === label) last.events.push(event);
    else days.push({ label, events: [event] });
  }

  return days.map((day) => (
    <div key={day.label}>
      <h3 className="mb-2 text-2xs font-mono font-semibold uppercase tracking-label text-text-muted">
        {day.label}
      </h3>
      <ol className="relative before:absolute before:left-[9px] before:top-4 before:bottom-4 before:w-0.5 before:rounded-full before:bg-text-muted/25">
        {day.events.map((event) => (
          <Entry key={event.id} event={event} decide={decide} />
        ))}
      </ol>
    </div>
  ));
}

function RegistrationActivity({
  history,
  empty,
}: {
  history: RegistrationHistory;
  empty: ReactNode;
}) {
  const refreshHistory = useInvalidateByIds(
    "provisioningKeyHistory",
    "listTenantRegistrations",
  );
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
    isRefreshing,
    error,
  } = history;
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

  if (isLoading || (isRefreshing && events.length === 0)) {
    return (
      <p className="py-12 text-center text-xs font-mono text-text-muted">
        Loading activity…
      </p>
    );
  }

  if (events.length === 0) return empty;

  return (
    <section
      aria-label="Registration activity"
      aria-busy={isRefreshing}
      className={cn(
        "space-y-8 transition-opacity",
        isRefreshing && "opacity-50",
      )}
    >
      <RegistrationDays events={events} decide={deviceActions.requestAction} />

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

function NoMatch({
  hostname,
  onClear,
}: {
  hostname: string;
  onClear: () => void;
}) {
  return (
    <div className="flex flex-col items-center gap-3 rounded-lg border border-dashed border-border px-6 py-12 text-center">
      <MagnifyingGlassIcon
        aria-hidden="true"
        className="h-6 w-6 text-text-muted"
        strokeWidth={1.5}
      />
      <div className="space-y-1">
        <p className="text-sm font-medium text-text-primary">
          No registrations matching &ldquo;{hostname}&rdquo;
        </p>
        <p className="text-xs text-text-muted">
          The search looks for the text anywhere in the hostname a device
          registered with.
        </p>
      </div>
      <Button variant="secondary" size="sm" onClick={onClear}>
        Clear search
      </Button>
    </div>
  );
}

function SortMenu({
  sort,
  onChange,
}: {
  sort: RegistrationSort;
  onChange: (next: RegistrationSort) => void;
}) {
  const [open, setOpen] = useState(false);
  const sorts = Object.entries(REGISTRATION_SORTS) as [
    RegistrationSort,
    (typeof REGISTRATION_SORTS)[RegistrationSort],
  ][];

  return (
    <Dropdown placement="bottom-end" portal open={open} onOpenChange={setOpen}>
      <Dropdown.Trigger>
        <button
          type="button"
          aria-label={`Sort registrations: ${REGISTRATION_SORTS[sort].label}`}
          className={TOOLBAR_CONTROL}
        >
          <ArrowsUpDownIcon className="w-3.5 h-3.5" strokeWidth={2} />
          {REGISTRATION_SORTS[sort].label}
          <ChevronDownIcon
            className={cn("w-3 h-3 transition-transform", open && "rotate-180")}
            strokeWidth={2}
          />
        </button>
      </Dropdown.Trigger>
      <Dropdown.Panel aria-label="Sort registrations" className="w-[200px]">
        {sorts.map(([key, { label }]) => (
          <Dropdown.Item
            key={key}
            label={label}
            onSelect={() => onChange(key)}
            className={cn(
              "justify-between",
              key === sort && "text-primary font-medium",
            )}
          >
            {label}
            {key === sort && (
              <CheckIcon className="w-3.5 h-3.5" strokeWidth={2} />
            )}
          </Dropdown.Item>
        ))}
      </Dropdown.Panel>
    </Dropdown>
  );
}

function SearchSummary({
  history,
  hostname,
  onClear,
}: {
  history: RegistrationHistory;
  hostname: string;
  onClear: () => void;
}) {
  if (
    !hostname ||
    history.isLoading ||
    history.isRefreshing ||
    history.events.length === 0
  ) {
    return null;
  }

  return (
    <p aria-live="polite" className="text-xs text-text-secondary">
      {formatCount(history.totalCount)}{" "}
      {history.totalCount === 1 ? "registration" : "registrations"} with &ldquo;
      {hostname}&rdquo; in the hostname &middot;{" "}
      <button
        type="button"
        onClick={onClear}
        className="font-medium text-primary hover:underline"
      >
        Clear
      </button>
    </p>
  );
}

/**
 * A registration history under a search by hostname and an order, one line per registration with
 * how it ended and the searched text marked in each hostname, newest or oldest first and grouped
 * by day. A device still waiting for a decision stands out as a card in the line, with what it reported and
 * its accept and reject; any other line expands to the same details. A refused agent is a line with
 * the reason and a link to the setting that refused it, and nothing to accept. Older registrations
 * load as the reader nears the end; a page that fails keeps the ones already shown and offers a
 * retry. history must be the one built from view. emptyMessage
 * is what an empty history says when nothing is searched; a search that matches nothing says so
 * and offers to clear it.
 */
export function RegistrationTimeline({
  history,
  view,
  emptyMessage,
}: {
  history: RegistrationHistory;
  view: ReturnType<typeof useRegistrationView>;
  emptyMessage: string;
}) {
  const clear = () => view.setSearch("");

  return (
    <SearchedHostname value={view.hostname}>
      <div className="space-y-6">
        <div className="space-y-3">
          <div className="flex items-center gap-2">
            <SearchField
              full
              toolbar
              className="flex-1"
              value={view.search}
              onChange={view.setSearch}
              placeholder="Search by hostname..."
              aria-label="Search registrations by hostname"
            />
            <SortMenu sort={view.sort} onChange={view.setSort} />
          </div>
          <SearchSummary
            history={history}
            hostname={view.hostname}
            onClear={clear}
          />
        </div>
        <RegistrationActivity
          history={history}
          empty={
            view.hostname ? (
              <NoMatch hostname={view.hostname} onClear={clear} />
            ) : (
              <p className="py-12 text-center text-xs text-text-muted">
                {emptyMessage}
              </p>
            )
          }
        />
      </div>
    </SearchedHostname>
  );
}

/**
 * What a provisioning key registered, as a RegistrationTimeline.
 */
export default function ProvisioningKeyTimeline({ id }: { id: string }) {
  const view = useRegistrationView();

  return (
    <RegistrationTimeline
      history={useProvisioningKeyEvents({ id, view })}
      view={view}
      emptyMessage="No registrations yet. Devices that register with this key will appear here."
    />
  );
}
