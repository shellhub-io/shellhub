import { useNavigate } from "react-router-dom";
import {
  NoSymbolIcon,
  PauseCircleIcon,
  PlusIcon,
  QrCodeIcon,
  TicketIcon,
} from "@heroicons/react/24/outline";
import { Button } from "@shellhub/design-system/primitives";
import { cn } from "@shellhub/design-system/cn";
import { type ProvisioningKey } from "@/client";
import DataTable, { type Column } from "@/components/common/DataTable";
import RestrictedAction from "@/components/common/RestrictedAction";
import ProvisioningKeyActionsMenu from "./ProvisioningKeyActionsMenu";
import StatusChip, { DeprecatedBadge } from "./StatusChip";
import UsageMeter from "./UsageMeter";
import { modeInfo } from "./constants";
import {
  getExpiryInfo,
  getKeyBlockers,
  provisioningKeyDisplayName,
  isPairingKey,
  isSystemKey,
} from "./helpers";

function KeyCell({
  provisioningKey: key,
}: {
  provisioningKey: ProvisioningKey;
}) {
  const { revoked, disabled, expired, inert, quiet } = getKeyBlockers(key);
  const system = isSystemKey(key);
  const mode = isPairingKey(key)
    ? { icon: QrCodeIcon, label: "Accepted by the code the agent prints" }
    : modeInfo(key.mode);
  const state = revoked
    ? { icon: NoSymbolIcon, label: "Revoked" }
    : disabled
      ? { icon: PauseCircleIcon, label: "Disabled" }
      : mode;
  const Icon = state.icon;
  const expiry = getExpiryInfo(key.expires_at);

  const facts = [
    state.label,
    key.expires_at &&
      (expired ? `expired ${expiry.label}` : `expires ${expiry.label}`),
    key.ephemeral && `removed after ${key.ephemeral_timeout ?? 10}m offline`,
  ].filter(Boolean);

  return (
    <div className="flex items-center gap-3 min-w-0">
      <span
        className={cn(
          "grid place-items-center w-8 h-8 rounded-lg shrink-0",
          inert
            ? "bg-text-muted/10 text-text-muted"
            : "bg-primary/10 text-primary",
        )}
        title={mode.label}
      >
        <Icon className="w-4 h-4" strokeWidth={1.8} />
      </span>
      <div className="min-w-0">
        <div className="flex items-center gap-2 min-w-0">
          <span
            className={cn(
              "text-sm font-medium truncate",
              inert ? "text-text-muted" : "text-text-primary",
            )}
          >
            {provisioningKeyDisplayName(key)}
          </span>
          {system && !isPairingKey(key) && <DeprecatedBadge />}
          {key.tags?.map((tag) => (
            <StatusChip key={tag} label={tag} tone="primary" mono />
          ))}
        </div>
        <p
          className={cn(
            "mt-0.5 text-2xs truncate",
            expired && !quiet ? "text-accent-red" : "text-text-muted",
          )}
        >
          {facts.join(" · ")}
        </p>
      </div>
    </div>
  );
}

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
        <RestrictedAction action="provisioningKey:create">
          <Button
            size="sm"
            onClick={onCreate}
            icon={<PlusIcon className="w-4 h-4" strokeWidth={2} />}
          >
            Create Provisioning Key
          </Button>
        </RestrictedAction>
      </div>
    </div>
  );
}

/**
 * The provisioning key list: each key on one line with its mode, expiry and tags folded under
 * its name, then its usage and the row actions. Inert keys (revoked/expired/overused) grey their
 * icon and meter so a live key's colour is never confused with a dead one's.
 */
export default function ProvisioningKeysTable({
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
  data: ProvisioningKey[];
  page: number;
  totalPages: number;
  totalCount: number;
  noCustomKeys: boolean;
  onPageChange: (page: number) => void;
  onCreate: () => void;
  onEdit: (key: ProvisioningKey) => void;
  onToggleDisabled: (key: ProvisioningKey) => void;
  onRevoke: (key: ProvisioningKey) => void;
}) {
  const navigate = useNavigate();

  const columns: Column<ProvisioningKey>[] = [
    {
      key: "name",
      header: "Key",
      render: (key) => <KeyCell provisioningKey={key} />,
    },
    {
      key: "usage",
      header: "Usage",
      headerClassName: "w-40",
      render: (key) => <UsageMeter provisioningKey={key} muted />,
    },
    {
      key: "actions",
      header: "",
      headerClassName: "w-12",
      render: (key) => (
        <div
          role="presentation"
          className="flex justify-end"
          onClick={(e) => e.stopPropagation()}
          onKeyDown={(e) => e.stopPropagation()}
        >
          <ProvisioningKeyActionsMenu
            provisioningKey={key}
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
      label="Provisioning keys"
      columns={columns}
      data={data}
      rowKey={(key) => key.name}
      sectionOf={(key) => (isSystemKey(key) ? "system" : "user")}
      sectionLabel={(section) =>
        section === "system" ? "Built-in" : "Custom keys"
      }
      trailingEmptyState={
        noCustomKeys ? <CustomKeysEmpty onCreate={onCreate} /> : undefined
      }
      rowClassName={(key) => {
        const base = "[&>td]:py-3.5";
        return key.revoked ? `${base} opacity-55` : base;
      }}
      onRowClick={(key) => {
        void navigate(
          `/settings/provisioning-keys/${encodeURIComponent(key.id)}/activity`,
          {
            state: {
              name: provisioningKeyDisplayName(key),
              key,
            },
          },
        );
      }}
      page={page}
      totalPages={totalPages}
      totalCount={totalCount}
      itemLabel="key"
      onPageChange={onPageChange}
    />
  );
}
