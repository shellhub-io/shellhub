import { type ReactNode } from "react";
import { Link, useNavigate } from "react-router-dom";
import {
  ChevronRightIcon,
  PlusIcon,
  TicketIcon,
} from "@heroicons/react/24/outline";
import { Button } from "@shellhub/design-system/primitives";
import { cn } from "@shellhub/design-system/cn";
import { type ProvisioningKey } from "@/client";
import DataTable, { type Column } from "@/components/common/DataTable";
import { capitalize } from "@/utils/string";
import RestrictedAction from "@/components/common/RestrictedAction";
import ProvisioningKeyActionsMenu from "./ProvisioningKeyActionsMenu";
import StatusChip, { DeprecatedBadge } from "./StatusChip";
import UsageMeter from "./UsageMeter";
import {
  ephemeralPhrase,
  expiryPhrase,
  getKeyBlockers,
  isInstallable,
  isPairingKey,
  isSystemKey,
  keyModeInfo,
  provisioningKeyDisplayName,
  provisioningKeyLink,
} from "./helpers";

function KeyCell({
  provisioningKey: key,
  selected,
}: {
  provisioningKey: ProvisioningKey;
  selected: boolean;
}) {
  const { revoked, disabled, expired, inert, quiet } = getKeyBlockers(key);
  const system = isSystemKey(key);
  const mode = keyModeInfo(key);
  const state = revoked
    ? "Revoked"
    : disabled
      ? "Disabled"
      : isPairingKey(key)
        ? capitalize(mode.outcome)
        : mode.label;

  const facts = [
    state,
    key.expires_at && expiryPhrase(key),
    ephemeralPhrase(key),
  ].filter(Boolean);

  return (
    <div className="flex items-center gap-3 min-w-0">
      <span
        title={inert ? "Not accepting devices" : "Accepting devices"}
        className={cn(
          "w-2 h-2 rounded-full shrink-0",
          inert ? "bg-text-muted/40" : "bg-accent-green",
        )}
      />
      <div className="min-w-0">
        <div className="flex items-center gap-2 min-w-0">
          <Link
            {...provisioningKeyLink(key)}
            onClick={(e) => e.stopPropagation()}
            className={cn(
              "text-sm font-medium truncate hover:text-primary hover:underline",
              inert
                ? "text-text-muted"
                : selected
                  ? "text-primary"
                  : "text-text-primary",
            )}
          >
            {provisioningKeyDisplayName(key)}
          </Link>
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
    <div>
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
            Create provisioning key
          </Button>
        </RestrictedAction>
      </div>
    </div>
  );
}

/**
 * The provisioning key list: each key on one line with its mode, expiry and tags folded under
 * its name, then its usage and the row actions. A dot before the name says whether the key still
 * lets devices in: green while it does, grey once it is revoked, disabled, expired or used up.
 * Custom keys come first and the built-in ones after them. Clicking an installable key's row
 * calls onSelect, and the key named by selectedName has renderInstall open beneath it; the name
 * opens the key's page.
 */
export default function ProvisioningKeysTable({
  data,
  selectedName,
  onSelect,
  renderInstall,
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
  selectedName: string | undefined;
  onSelect: (key: ProvisioningKey) => void;
  renderInstall: (key: ProvisioningKey) => ReactNode;
  noCustomKeys: boolean;
  onPageChange: (page: number) => void;
  onCreate: () => void;
  onEdit: (key: ProvisioningKey) => void;
  onToggleDisabled: (key: ProvisioningKey) => void;
  onRevoke: (key: ProvisioningKey) => void;
}) {
  const navigate = useNavigate();
  const openActivity = (key: ProvisioningKey) => {
    const { to, state } = provisioningKeyLink(key);
    void navigate(to, { state });
  };

  const columns: Column<ProvisioningKey>[] = [
    {
      key: "name",
      header: "Key",
      render: (key) => (
        <KeyCell provisioningKey={key} selected={key.name === selectedName} />
      ),
    },
    {
      key: "usage",
      header: "Usage",
      headerClassName: "w-40",
      render: (key) => <UsageMeter provisioningKey={key} muted linkWaiting />,
    },
    {
      key: "actions",
      header: "",
      headerClassName: "w-20",
      render: (key) => (
        <div
          role="presentation"
          className="flex items-center justify-end gap-1"
          onClick={(e) => e.stopPropagation()}
          onKeyDown={(e) => e.stopPropagation()}
        >
          <ProvisioningKeyActionsMenu
            provisioningKey={key}
            onEdit={onEdit}
            onToggleDisabled={onToggleDisabled}
            onRevoke={onRevoke}
            onActivity={openActivity}
          />
          <ChevronRightIcon
            aria-hidden="true"
            className={cn(
              "w-4 h-4 shrink-0 text-text-muted transition-transform",
              !isInstallable(key) && "invisible",
              key.name === selectedName && "rotate-90 text-primary",
            )}
            strokeWidth={2}
          />
        </div>
      ),
    },
  ];

  const ordered = [
    ...data.filter((key) => !isSystemKey(key)),
    ...data.filter(isSystemKey),
  ];
  const paged = totalPages > 1;

  return (
    <div className="space-y-4">
      {noCustomKeys && <CustomKeysEmpty onCreate={onCreate} />}
      <DataTable
        label="Provisioning keys"
        columns={columns}
        data={ordered}
        rowKey={(key) => key.name}
        sectionOf={(key) => (isSystemKey(key) ? "Built-in" : undefined)}
        expandedRowKey={selectedName ?? null}
        renderExpandedRow={renderInstall}
        rowClassName={(key) =>
          cn(
            "[&>td]:py-3.5",
            key.revoked && "opacity-55",
            !isInstallable(key) && "cursor-default",
            key.name === selectedName && "bg-primary/[0.06]",
          )
        }
        onRowClick={(key) => {
          if (isInstallable(key)) onSelect(key);
        }}
        page={paged ? page : undefined}
        totalPages={paged ? totalPages : undefined}
        totalCount={paged ? totalCount : undefined}
        itemLabel="key"
        onPageChange={paged ? onPageChange : undefined}
      />
    </div>
  );
}
