import { useNavigate } from "react-router-dom";
import {
  BoltIcon,
  ClockIcon,
  NoSymbolIcon,
  PauseCircleIcon,
  PlusIcon,
  QrCodeIcon,
  TicketIcon,
} from "@heroicons/react/24/outline";
import { ExclamationCircleIcon as ExclamationCircleSolidIcon } from "@heroicons/react/24/solid";
import { Button } from "@shellhub/design-system/primitives";
import { type InstallKey } from "@/client";
import DataTable, { type Column } from "@/components/common/DataTable";
import RestrictedAction from "@/components/common/RestrictedAction";
import InstallKeyActionsMenu from "./InstallKeyActionsMenu";
import StatusChip from "./StatusChip";
import UsageMeter, { usageLabel } from "./UsageMeter";
import { DeprecatedBadge, modeInfo } from "./constants";
import {
  getExpiryInfo,
  getKeyBlockers,
  getUsageInfo,
  installKeyDisplayName,
  isPairingKey,
  isSystemKey,
} from "./helpers";

/**
 * The Mode cell: the key's mode identity (icon tile + label). Its secondary line is the mode summary
 * for a live key, or — for the two key-level states that have no column of their own — "Revoked" or
 * "Disabled" with an icon and a contextual tone. Expired / limit-reached are chips in their own
 * columns instead. The tile dims when the key is inert.
 */
function EnrollmentCell({ installKey }: { installKey: InstallKey }) {
  const { revoked, disabled, inert } = getKeyBlockers(installKey);

  // The pairing key has no configurable admission mode — it always accepts via the printed code — so
  // it reads as a pairing source, not a mode. Every other key shows its actual mode.
  const info = isPairingKey(installKey)
    ? { icon: QrCodeIcon, label: "Code pairing", summary: "Accepted by code" }
    : modeInfo(installKey.mode);
  const Icon = info.icon;

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
          {info.label}
        </div>
        {state ? (
          <div className="flex items-center gap-1 text-2xs text-text-muted whitespace-nowrap">
            <state.glyph className="w-3 h-3 shrink-0" strokeWidth={2} />
            {state.label}
          </div>
        ) : (
          <div className="text-2xs text-text-muted whitespace-nowrap">
            {info.summary}
          </div>
        )}
      </div>
    </div>
  );
}

/**
 * Onboarding placeholder for the "Custom keys" section when the namespace has none yet. The built-in
 * keys always fill the table, so this replaces the old full-page hero: it lives inside the section it
 * belongs to, keeping the built-in rows (and their queues) reachable above it.
 */
function CustomKeysEmpty({ onCreate }: { onCreate: () => void }) {
  return (
    <div className="px-4 py-6">
      <div className="flex flex-col items-center gap-3 rounded-xl border border-dashed border-border-light px-5 py-9 text-center">
        <TicketIcon className="w-8 h-8 text-text-muted" strokeWidth={1.5} />
        <h3 className="text-sm font-semibold text-text-primary">
          No custom keys yet
        </h3>
        <p className="max-w-xl text-xs text-text-muted">
          Create a key with its own secret and mode to register devices with
          your namespace.
        </p>
        <RestrictedAction action="installKey:create">
          <Button
            size="sm"
            onClick={onCreate}
            icon={<PlusIcon className="w-4 h-4" strokeWidth={2} />}
          >
            Create Install Key
          </Button>
        </RestrictedAction>
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
  noCustomKeys,
  onPageChange,
  onCreate,
  onEdit,
  onToggleDisabled,
  onRevoke,
}: {
  data: InstallKey[];
  page: number;
  totalPages: number;
  totalCount: number;
  /** True when the namespace has only built-in keys (no user-created ones), so the Custom keys
   * section shows its onboarding placeholder instead of rows. Computed by the caller off the true
   * count — page data alone can't distinguish "none" from "on a later page". */
  noCustomKeys: boolean;
  onPageChange: (page: number) => void;
  onCreate: () => void;
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
                {installKeyDisplayName(key)}
              </span>
              {!isPairingKey(key) && <DeprecatedBadge />}
            </div>
            <p className="mt-0.5 text-2xs text-text-muted">
              {isPairingKey(key)
                ? "Devices that pair by the code the agent prints during install — or by opening the link printed with it. Accepted on the code itself."
                : "The legacy path for devices that register without an install key. Kept for compatibility and will be removed in a future release. Its mode controls how those devices are handled."}
            </p>
          </div>
        ) : (
          <div>
            <span className="text-sm font-semibold text-text-primary">
              {key.name}
            </span>
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

  return (
    <DataTable
      label="Install keys"
      columns={columns}
      data={data}
      rowKey={(key) => key.name}
      // The auto-managed keys (legacy, pairing) come with the namespace ("Built-in"); the keys an admin
      // creates are grouped under "Custom keys". `data` is already ordered system-first, so the
      // sections stay contiguous.
      sectionOf={(key) => (isSystemKey(key) ? "system" : "user")}
      sectionLabel={(section) =>
        section === "system" ? "Built-in" : "Custom keys"
      }
      // With only built-in keys, the "Custom keys" section has no rows — render its onboarding
      // placeholder under the built-in rows instead of the old full-page hero (which never showed,
      // since the built-in keys keep the table non-empty). The placeholder names the section itself,
      // so no "Custom keys" header precedes it.
      trailingEmptyState={
        noCustomKeys ? <CustomKeysEmpty onCreate={onCreate} /> : undefined
      }
      // Tall, roomy rows; clicking a row opens the key's full registration history.
      rowClassName={(key) => {
        const base = "[&>td]:py-5";
        // A revoked key is terminal and archival: fade the whole row so it recedes, while a disabled
        // key (reversible pause) stays at full presence.
        return key.revoked ? `${base} opacity-55` : base;
      }}
      onRowClick={(key) => {
        void navigate(`/install-keys/${encodeURIComponent(key.id)}/activity`, {
          state: {
            name: installKeyDisplayName(key),
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
