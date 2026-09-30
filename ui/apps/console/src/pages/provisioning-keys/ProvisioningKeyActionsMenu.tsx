import { type ReactNode } from "react";
import {
  ClockIcon,
  EllipsisVerticalIcon,
  NoSymbolIcon,
  PauseIcon,
  PencilIcon,
  PlayIcon,
} from "@heroicons/react/24/outline";
import { Dropdown, IconButton } from "@shellhub/design-system/primitives";
import { type ProvisioningKey } from "@/client";
import RestrictedAction from "@/components/common/RestrictedAction";
import { type Action } from "@/utils/permission";
import { isPairingKey, isSystemKey } from "./helpers";

function MenuItem({
  action,
  icon,
  label,
  danger,
  onSelect,
}: {
  action: Action;
  icon: ReactNode;
  label: string;
  danger?: boolean;
  onSelect: () => void;
}) {
  return (
    <RestrictedAction action={action}>
      <Dropdown.Item
        label={label}
        variant={danger ? "danger" : "default"}
        onSelect={onSelect}
        className="gap-2.5 px-3 py-2"
      >
        <span className="shrink-0">{icon}</span>
        {label}
      </Dropdown.Item>
    </RestrictedAction>
  );
}

/**
 * The provisioning key overflow menu. Disabling is offered separately from deleting: a disabled key
 * can be re-enabled, and a deleted one cannot. Given onActivity, it also opens the key's
 * registration history, which every key has, the pairing code and revoked keys included.
 */
export default function ProvisioningKeyActionsMenu({
  provisioningKey,
  onEdit,
  onToggleDisabled,
  onRevoke,
  onActivity,
}: {
  provisioningKey: ProvisioningKey;
  onEdit: (key: ProvisioningKey) => void;
  onToggleDisabled: (key: ProvisioningKey) => void;
  onRevoke: (key: ProvisioningKey) => void;
  onActivity?: (key: ProvisioningKey) => void;
}) {
  const pairing = isPairingKey(provisioningKey);
  const editable = !pairing && !provisioningKey.revoked;
  if (!editable && !onActivity) return null;

  return (
    <Dropdown portal placement="bottom-end">
      <Dropdown.Trigger>
        <IconButton variant="ghost" aria-label="Provisioning key actions">
          <EllipsisVerticalIcon className="w-4 h-4" />
        </IconButton>
      </Dropdown.Trigger>

      <Dropdown.Panel className="w-44 py-1">
        {onActivity && (
          <Dropdown.Item
            label="Activity"
            onSelect={() => onActivity(provisioningKey)}
            className="gap-2.5 px-3 py-2"
          >
            <span className="shrink-0">
              <ClockIcon className="w-4 h-4" />
            </span>
            Activity
          </Dropdown.Item>
        )}
        {editable && (
          <>
            {onActivity && <Dropdown.Separator />}
            <MenuItem
              action="provisioningKey:edit"
              icon={<PencilIcon className="w-4 h-4" />}
              label="Edit"
              onSelect={() => onEdit(provisioningKey)}
            />
            <MenuItem
              action="provisioningKey:disable"
              icon={
                provisioningKey.disabled ? (
                  <PlayIcon className="w-4 h-4" />
                ) : (
                  <PauseIcon className="w-4 h-4" />
                )
              }
              label={provisioningKey.disabled ? "Enable" : "Disable"}
              onSelect={() => onToggleDisabled(provisioningKey)}
            />
            {!isSystemKey(provisioningKey) && (
              <>
                <Dropdown.Separator />
                <MenuItem
                  action="provisioningKey:revoke"
                  icon={<NoSymbolIcon className="w-4 h-4" />}
                  label="Revoke"
                  danger
                  onSelect={() => onRevoke(provisioningKey)}
                />
              </>
            )}
          </>
        )}
      </Dropdown.Panel>
    </Dropdown>
  );
}
