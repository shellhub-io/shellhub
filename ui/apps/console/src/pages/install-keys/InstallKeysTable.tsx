import { useNavigate } from "react-router-dom";
import {
  BoltIcon,
  ClockIcon,
  NoSymbolIcon,
  PauseCircleIcon,
} from "@heroicons/react/24/outline";
import { ExclamationCircleIcon as ExclamationCircleSolidIcon } from "@heroicons/react/24/solid";
import { type InstallKey } from "@/client";
import DataTable, { type Column } from "@/components/common/DataTable";
import InstallKeyActionsMenu from "./InstallKeyActionsMenu";
import InstallKeyEventsTable from "./InstallKeyEventsTable";
import KeyValueChip from "./KeyValueChip";
import StatusChip from "./StatusChip";
import UsageMeter, { usageLabel } from "./UsageMeter";
import { DeprecatedBadge, modeInfo } from "./constants";
import {
  getExpiryInfo,
  getKeyBlockers,
  getUsageInfo,
  isSystemKey,
} from "./helpers";

/**
 * The install key's value under its name: the masked secret hint with an inline reveal. This is the
 * artifact the feature is named for — the string you paste into an agent — so it sits in the identity
 * cell, not a separate column. The fingerprint (webhook/route identifier) lives in the reveal dialog.
 */
function KeyValue({
  installKey,
  onReveal,
}: {
  installKey: InstallKey;
  onReveal: (key: InstallKey) => void;
}) {
  const hasSecret = !!installKey.key_hint;

  return (
    <div className="mt-1.5">
      <KeyValueChip
        label="Install Key"
        labelTone="primary"
        value={hasSecret ? `${installKey.key_hint}••••••` : "••••••••••••••"}
        onClick={() => onReveal(installKey)}
        title="Reveal Install Key"
        ariaLabel="Reveal Install Key"
      />
    </div>
  );
}

/**
 * The Mode cell: the key's mode identity (icon tile + label). Its secondary line is the mode summary
 * for a live key, or — for the two key-level states that have no column of their own — "Revoked" or
 * "Disabled" with an icon and a contextual tone. Expired / limit-reached are chips in their own
 * columns instead. The tile dims when the key is inert.
 */
function EnrollmentCell({ installKey }: { installKey: InstallKey }) {
  const mode = modeInfo(installKey.mode);
  const Icon = mode.icon;
  const { revoked, disabled, inert } = getKeyBlockers(installKey);

  const state = revoked
    ? { label: "Revoked", glyph: NoSymbolIcon }
    : disabled
      ? { label: "Disabled", glyph: PauseCircleIcon }
      : null;

  return (
    <div className="flex items-center gap-2.5">
      <span
        className={`grid place-items-center w-7 h-7 rounded-lg shrink-0 ${
          inert
            ? "bg-text-muted/10 text-text-muted"
            : "bg-primary/10 text-primary"
        }`}
      >
        <Icon className="w-4 h-4" strokeWidth={1.8} />
      </span>
      <div className="min-w-0">
        <div
          className={`text-xs font-semibold whitespace-nowrap ${inert ? "text-text-muted" : "text-text-primary"}`}
        >
          {mode.label}
        </div>
        {state ? (
          <div className="flex items-center gap-1 text-2xs text-text-muted whitespace-nowrap">
            <state.glyph className="w-3 h-3 shrink-0" strokeWidth={2} />
            {state.label}
          </div>
        ) : (
          <div className="text-2xs text-text-muted whitespace-nowrap">
            {mode.summary}
          </div>
        )}
      </div>
    </div>
  );
}

/**
 * The install key list as a table with tall, content-rich rows: identity (name,
 * status, masked secret, tags) on the left, then the enrollment meter, expiry,
 * and the row actions. Inert keys (revoked/expired/overused) grey their meter so
 * a live key's colour is never confused with a dead one's.
 */
export default function InstallKeysTable({
  data,
  page,
  totalPages,
  totalCount,
  onPageChange,
  onReveal,
  onEdit,
  onToggleDisabled,
  onRevoke,
}: {
  data: InstallKey[];
  page: number;
  totalPages: number;
  totalCount: number;
  onPageChange: (page: number) => void;
  onReveal: (key: InstallKey) => void;
  onEdit: (key: InstallKey) => void;
  onToggleDisabled: (key: InstallKey) => void;
  onRevoke: (key: InstallKey) => void;
}) {
  const navigate = useNavigate();

  const columns: Column<InstallKey>[] = [
    {
      key: "enrollment",
      header: "Mode",
      render: (key) => <EnrollmentCell installKey={key} />,
    },
    {
      key: "name",
      header: "Key",
      render: (key) =>
        isSystemKey(key) ? (
          <div>
            <div className="flex items-center gap-2">
              <span className="text-sm font-semibold text-text-primary">
                Tenant-only registration
              </span>
              <DeprecatedBadge />
            </div>
            <p className="mt-0.5 text-2xs text-text-muted">
              The legacy path for devices that register without an install key.
              Kept for compatibility and will be removed in a future release.
              Its mode controls how those devices are handled.
            </p>
            <KeyValue installKey={key} onReveal={onReveal} />
          </div>
        ) : (
          <div>
            <span className="text-sm font-semibold text-text-primary">
              {key.name}
            </span>
            <KeyValue installKey={key} onReveal={onReveal} />
            {key.tags && key.tags.length > 0 && (
              <div className="mt-1.5 flex items-center gap-1.5 flex-wrap">
                {key.tags.map((tag) => (
                  <StatusChip key={tag} label={tag} tone="primary" mono />
                ))}
              </div>
            )}
          </div>
        ),
    },
    {
      key: "usage",
      header: "Usage limit",
      render: (key) => {
        const usage = getUsageInfo(key);
        const { inert, overused, revoked, disabled } = getKeyBlockers(key);
        const reached = overused && !revoked && !disabled;
        return (
          <div title={reached ? "Limit reached" : undefined}>
            <div className="mb-1.5 text-2xs font-mono text-text-secondary">
              {usageLabel(usage)}
            </div>
            <UsageMeter usage={usage} dimmed={inert} reached={reached} />
          </div>
        );
      },
    },
    {
      key: "expiry",
      header: "Expires",
      render: (key) => {
        const expiry = getExpiryInfo(key.expires_at);
        const { expired, revoked, disabled } = getKeyBlockers(key);
        // The date stays muted (never an alarm-red string). An elapsed expiry swaps the clock for a
        // solid exclamation glyph — shown even on a revoked/disabled key, so the "expired" fact never
        // hides behind another state. Only the red pip is conditional: a revoked/disabled key mutes
        // it (that expiry no longer matters), a live one keeps it as the at-a-glance signal.
        const quiet = revoked || disabled;
        const dateTitle = expired ? "Expired" : undefined;
        return (
          <div className="flex flex-col gap-1">
            <span
              title={dateTitle}
              className="flex items-center gap-1 text-2xs font-mono text-text-muted whitespace-nowrap"
            >
              {expired ? (
                <ExclamationCircleSolidIcon
                  className={`w-3.5 h-3.5 shrink-0 ${quiet ? "" : "text-accent-red"}`}
                />
              ) : (
                <ClockIcon className="w-3.5 h-3.5 shrink-0" />
              )}
              {expiry.label}
            </span>
            {key.ephemeral && (
              <span
                className="flex items-center gap-1 text-2xs text-text-muted whitespace-nowrap"
                title="Registered devices are removed after staying offline past the timeout"
              >
                <BoltIcon className="w-3.5 h-3.5 shrink-0" />
                {key.ephemeral_timeout ?? 10}m
              </span>
            )}
          </div>
        );
      },
    },
    {
      key: "actions",
      header: "",
      headerClassName: "w-12",
      render: (key) => (
        <div className="flex justify-end" onClick={(e) => e.stopPropagation()}>
          <InstallKeyActionsMenu
            installKey={key}
            onEdit={onEdit}
            onToggleDisabled={onToggleDisabled}
            onRevoke={onRevoke}
          />
        </div>
      ),
    },
  ];

  // The tenant-only (system) row stays permanently expanded, previewing its recent registrations
  // inline so the legacy path's activity is visible without leaving the list. When it's disabled the
  // path is off, so the preview collapses.
  const systemKey = data.find(isSystemKey);
  const expandSystemKey = systemKey && !systemKey.disabled;

  return (
    <DataTable
      label="Install keys"
      columns={columns}
      data={data}
      rowKey={(key) => key.name}
      expandedRowKey={expandSystemKey ? systemKey.name : null}
      renderExpandedRow={(key) => (
        // The key's registration preview, tied to the row above by a shared left rail and a recessed
        // background so it reads as this row's nested detail.
        <div className="-ml-px border-l-2 border-l-primary/50 bg-surface/60">
          <InstallKeyEventsTable
            id={key.id}
            perPage={10}
            compact
            viewAll={{
              to: `/install-keys/${encodeURIComponent(key.id)}/activity`,
              state: { name: "Tenant-only registration", key },
            }}
          />
        </div>
      )}
      // Tall, roomy rows; clicking a row opens the key's full enrollment history. The pinned legacy
      // row gets a faint tint so it reads as the default/system entry.
      rowClassName={(key) => {
        const system = isSystemKey(key);
        // An expanded tenant-only row shares its preview's recessed background and left rail so the two
        // read as one block. `[&+tr]:!border-t-transparent` drops the table's divider between the row
        // and the accordion, so the rail runs unbroken (no grey notch at the seam). Collapsed, the row
        // keeps only the faint primary tint that marks the system row.
        const systemClass = system
          ? expandSystemKey
            ? " [&>td:first-child]:border-l-2 [&>td:first-child]:border-l-primary/50 [&+tr]:!border-t-transparent"
            : " bg-primary/[0.03]"
          : "";
        const base = `[&>td]:py-5${systemClass}`;
        // A revoked key is terminal and archival: fade the whole row so it recedes, while a disabled
        // key (reversible pause) stays at full presence.
        return key.revoked ? `${base} opacity-55` : base;
      }}
      onRowClick={(key) => {
        void navigate(`/install-keys/${encodeURIComponent(key.id)}/activity`, {
          state: {
            name: isSystemKey(key) ? "Tenant-only registration" : key.name,
            key,
          },
        });
      }}
      page={page}
      totalPages={totalPages}
      totalCount={totalCount}
      itemLabel="key"
      onPageChange={onPageChange}
    />
  );
}
