import { Fragment, type ReactNode, useState } from "react";
import { useLocation, useParams } from "react-router-dom";
import {
  ChevronRightIcon,
  NoSymbolIcon,
  PauseCircleIcon,
  TicketIcon,
} from "@heroicons/react/24/outline";
import { IconBadge } from "@shellhub/design-system/primitives";
import { cn } from "@shellhub/design-system/cn";
import { type ProvisioningKey } from "@/client";
import Breadcrumb from "@/components/common/Breadcrumb";
import CopyButton from "@/components/common/CopyButton";
import PageLoader from "@/components/common/PageLoader";
import ResourceNotFound from "@/components/common/ResourceNotFound";
import RestrictedAction from "@/components/common/RestrictedAction";
import { useProvisioningKeys } from "@/hooks/useProvisioningKeys";
import { useRevealProvisioningKey } from "@/hooks/useRevealProvisioningKey";
import { formatDateShort } from "@/utils/date";
import { capitalize } from "@/utils/string";
import ProvisioningKeyActions from "./ProvisioningKeyActions";
import ProvisioningKeyTimeline from "./ProvisioningKeyTimeline";
import StatusChip, { DeprecatedBadge } from "./StatusChip";
import { UsageCount, UsageDetail } from "./UsageMeter";
import {
  ephemeralPhrase,
  expiryPhrase,
  formatCount,
  getUsageInfo,
  getWaitingInfo,
  isPairingKey,
  isSystemKey,
  keyModeInfo,
  lastUsedPhrase,
  provisioningKeyDisplayName,
} from "./helpers";

function elideUrl(raw: string): string {
  let url: URL;
  try {
    url = new URL(raw);
  } catch {
    return raw.length > 36 ? `${raw.slice(0, 20)}…${raw.slice(-12)}` : raw;
  }
  const segments = url.pathname.split("/").filter(Boolean);
  const path =
    segments.length > 2
      ? `/${segments[0]}/…/${segments[segments.length - 1]}`
      : url.pathname.replace(/\/$/, "");
  return `${url.host}${path}`;
}

function modeSummary(key: ProvisioningKey): ReactNode {
  const label = keyModeInfo(key).label;
  if (key.mode === "webhook" && key.webhook_url) {
    return (
      <>
        {label}{" "}
        <span className="font-mono text-text-muted" title={key.webhook_url}>
          {elideUrl(key.webhook_url)}
        </span>
      </>
    );
  }
  if (key.mode === "allowlist") {
    const count = key.allowed_identities?.length ?? 0;
    return `${label}, ${formatCount(count)} ${count === 1 ? "identity" : "identities"}`;
  }
  return label;
}

function SecretKey({
  provisioningKey: key,
}: {
  provisioningKey: ProvisioningKey;
}) {
  const [shown, setShown] = useState(false);
  const {
    key: secret,
    isLoading,
    error,
  } = useRevealProvisioningKey(key.name, shown);

  if (!shown) {
    return (
      <span className="inline-flex items-center gap-2">
        <span className="font-mono text-text-muted">
          {key.key_hint}••••••••••••••••
        </span>
        <RestrictedAction action="provisioningKey:reveal">
          <button
            type="button"
            onClick={() => setShown(true)}
            className="font-medium text-primary hover:underline"
          >
            Show
          </button>
        </RestrictedAction>
      </span>
    );
  }
  if (isLoading) return <span className="text-text-muted">Loading…</span>;
  if (error)
    return <span className="text-accent-red">Could not load the key.</span>;
  return (
    <span className="inline-flex flex-col gap-1">
      <span className="inline-flex items-center gap-1.5">
        <span className="font-mono break-all">{secret}</span>
        <CopyButton text={secret} />
      </span>
      <span className="text-2xs text-accent-yellow">
        Treat like a password. Anyone with it can register devices with your
        namespace.
      </span>
    </span>
  );
}

function KeyDetails({
  provisioningKey: key,
}: {
  provisioningKey: ProvisioningKey;
}) {
  const rows: [string, ReactNode][] = [];
  if (key.mode === "webhook" && key.webhook_url) {
    rows.push([
      "Webhook",
      <span className="font-mono break-all">{key.webhook_url}</span>,
    ]);
    if (key.webhook_timeout) rows.push(["Timeout", `${key.webhook_timeout}s`]);
    if (key.webhook_callback_ttl)
      rows.push(["Callback window", `${key.webhook_callback_ttl}s`]);
  }
  if (key.mode === "allowlist") {
    rows.push([
      "Allowed identities",
      <span className="flex flex-wrap gap-1">
        {(key.allowed_identities ?? []).map((identity) => (
          <StatusChip key={identity} label={identity} tone="muted" mono />
        ))}
      </span>,
    ]);
  }
  if (key.key_hint) rows.push(["Key", <SecretKey provisioningKey={key} />]);
  rows.push([
    "Fingerprint",
    <span className="inline-flex items-center gap-1.5 min-w-0">
      <span className="font-mono break-all">{key.id}</span>
      <CopyButton text={key.id} />
    </span>,
  ]);

  return (
    <dl className="mt-3 grid grid-cols-[max-content_1fr] gap-x-5 gap-y-1.5 rounded-lg border border-border bg-card px-4 py-3 text-xs">
      {rows.map(([label, value]) => (
        <Fragment key={label}>
          <dt className="text-text-muted">{label}</dt>
          <dd className="min-w-0 text-text-primary">{value}</dd>
        </Fragment>
      ))}
    </dl>
  );
}

function UsageSummary({
  provisioningKey: key,
  detailsOpen,
  onToggleDetails,
}: {
  provisioningKey: ProvisioningKey;
  detailsOpen: boolean;
  onToggleDetails: () => void;
}) {
  const usage = getUsageInfo(key);
  const { waiting } = getWaitingInfo(key);
  const usedPct = usage.kind === "unlimited" ? 0 : usage.ratio * 100;
  const waitingPct =
    usage.kind === "unlimited"
      ? 0
      : Math.min(100 - usedPct, (waiting / usage.limit) * 100);

  return (
    <div>
      <div className="flex flex-wrap items-baseline gap-x-3 gap-y-1">
        <UsageCount provisioningKey={key} large />
        <UsageDetail provisioningKey={key} className="text-sm" />
        <span className="ml-auto text-xs text-text-muted">
          {lastUsedPhrase(key)}
        </span>
      </div>
      <div
        className={cn(
          "mt-2.5 mb-3 flex h-1.5 overflow-hidden rounded-full",
          usage.kind === "unlimited"
            ? "bg-[repeating-linear-gradient(90deg,rgb(var(--c-border))_0_6px,transparent_6px_10px)]"
            : "bg-border",
        )}
      >
        {usage.kind !== "unlimited" && (
          <>
            <span
              className="h-full bg-primary"
              style={{ width: `${usedPct}%` }}
            />
            <span
              className="h-full bg-accent-yellow/60"
              style={{ width: `${waitingPct}%` }}
            />
          </>
        )}
      </div>
      <div className="flex items-start gap-4 text-xs text-text-secondary">
        <div className="flex flex-1 min-w-0 flex-wrap items-center gap-x-4 gap-y-1.5">
          <span>{modeSummary(key)}</span>
          <span>{capitalize(expiryPhrase(key))}</span>
          {key.ephemeral && <span>Ephemeral, {ephemeralPhrase(key)}</span>}
          <span>Created {formatDateShort(key.created_at)}</span>
          {key.tags.length > 0 && (
            <span className="flex flex-wrap items-center gap-1">
              {key.tags.map((tag) => (
                <StatusChip key={tag} label={tag} tone="primary" mono />
              ))}
            </span>
          )}
        </div>
        {!isPairingKey(key) && (
          <button
            type="button"
            onClick={onToggleDetails}
            aria-expanded={detailsOpen}
            className="shrink-0 inline-flex items-center gap-1 font-medium text-primary hover:text-primary/80"
          >
            <ChevronRightIcon
              className={cn(
                "w-3 h-3 transition-transform",
                detailsOpen && "rotate-90",
              )}
              strokeWidth={2.5}
            />
            Details
          </button>
        )}
      </div>
      {detailsOpen && <KeyDetails provisioningKey={key} />}
    </div>
  );
}

/**
 * A single provisioning key's page: its name and what happens to each device, its usage against
 * the allowance with the configuration beneath, then its registration activity. Keyed by the
 * key's id (digest). Details open on their own for a webhook key, whose endpoint is what a
 * visitor usually comes to check.
 */
export default function ProvisioningKeyHistoryPage() {
  const { id = "" } = useParams();
  const location = useLocation();
  const state = location.state as {
    name?: string;
    key?: ProvisioningKey;
  } | null;
  const [detailsChoice, setDetailsChoice] = useState<boolean | null>(null);

  const { provisioningKeys, isLoading } = useProvisioningKeys({ perPage: 100 });
  const key = provisioningKeys.find((k) => k.id === id) ?? state?.key ?? null;
  const name = key ? provisioningKeyDisplayName(key) : (state?.name ?? "");

  const detailsOpen = detailsChoice ?? key?.mode === "webhook";

  if (isLoading && !key) {
    return <PageLoader label="Loading provisioning key" />;
  }

  if (!key) {
    return (
      <ResourceNotFound
        icon={TicketIcon}
        resource="Provisioning key"
        backTo="/devices/add/fleet"
      />
    );
  }

  return (
    <div className="max-w-3xl">
      <Breadcrumb
        items={[
          { label: "Devices", to: "/devices" },
          { label: "Add Device", to: "/devices/add" },
          { label: "Fleet", to: "/devices/add/fleet" },
          { label: name || "Provisioning key" },
        ]}
        className="mb-4"
      />

      <div className="mb-6">
        <div className="flex items-start justify-between gap-4">
          <div className="flex items-start gap-4 min-w-0">
            <IconBadge size="lg" color="primary">
              <TicketIcon className="w-6 h-6" />
            </IconBadge>
            <div className="min-w-0">
              <div className="flex flex-wrap items-center gap-2">
                <h1 className="text-xl font-semibold text-text-primary leading-tight truncate">
                  {name || "Provisioning key"}
                </h1>
                {isSystemKey(key) && !isPairingKey(key) && <DeprecatedBadge />}
                {key.revoked && (
                  <StatusChip icon={NoSymbolIcon} label="Revoked" tone="red" />
                )}
                {key.disabled && !key.revoked && (
                  <StatusChip
                    icon={PauseCircleIcon}
                    label="Disabled"
                    tone="muted"
                  />
                )}
              </div>
              <p className="mt-1 text-sm text-text-muted">
                Each device is {keyModeInfo(key).outcome}.
              </p>
            </div>
          </div>
          <ProvisioningKeyActions provisioningKey={key} />
        </div>
      </div>

      <div className="mb-10 sm:ml-16 pb-8 border-b border-border">
        <UsageSummary
          provisioningKey={key}
          detailsOpen={detailsOpen}
          onToggleDetails={() => setDetailsChoice(!detailsOpen)}
        />
      </div>

      <div className="sm:ml-16">
        <ProvisioningKeyTimeline id={id} />
      </div>
    </div>
  );
}
