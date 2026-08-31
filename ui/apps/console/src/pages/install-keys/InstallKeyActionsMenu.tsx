import { type ReactNode } from "react";
import {
  EllipsisVerticalIcon,
  NoSymbolIcon,
  PauseIcon,
  PencilIcon,
  PlayIcon,
} from "@heroicons/react/24/outline";
import { Dropdown, IconButton } from "@shellhub/design-system/primitives";
import { type InstallKey } from "@/client";
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
 * The install key overflow menu. Disabling is offered separately from deleting: a disabled key
 * can be re-enabled, and a deleted one cannot.
 */
export default function InstallKeyActionsMenu({
  installKey,
  onEdit,
  onToggleDisabled,
  onRevoke,
}: {
  installKey: InstallKey;
  onEdit: (key: InstallKey) => void;
  onToggleDisabled: (key: InstallKey) => void;
  onRevoke: (key: InstallKey) => void;
}) {
  if (isPairingKey(installKey)) return null;
  if (installKey.revoked) return null;

  return (
    <Dropdown portal placement="bottom-end">
      <Dropdown.Trigger>
        <IconButton
          variant="ghost"
          aria-label="Install Key actions"
        >
          <EllipsisVerticalIcon className="w-4 h-4" />
        </IconButton>
      </Dropdown.Trigger>

      <Dropdown.Panel className="w-40 py-1">
        <MenuItem
          action="installKey:edit"
          icon={<PencilIcon className="w-4 h-4" />}
          label="Edit"
          onSelect={() => onEdit(installKey)}
        />
        <MenuItem
          action="installKey:disable"
          icon={
            installKey.disabled ? (
              <PlayIcon className="w-4 h-4" />
            ) : (
              <PauseIcon className="w-4 h-4" />
            )
          }
          label={installKey.disabled ? "Enable" : "Disable"}
          onSelect={() => onToggleDisabled(installKey)}
        />
        {!isSystemKey(installKey) && (
          <MenuItem
            action="installKey:revoke"
            icon={<NoSymbolIcon className="w-4 h-4" />}
            label="Revoke"
            danger
            onSelect={() => onRevoke(installKey)}
          />
        )}
      </Dropdown.Panel>
    </Dropdown>
  );
}
