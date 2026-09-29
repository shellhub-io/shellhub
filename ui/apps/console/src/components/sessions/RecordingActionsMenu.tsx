import { useState, type ReactNode } from "react";
import {
  ArrowDownTrayIcon,
  EllipsisVerticalIcon,
  TrashIcon,
} from "@heroicons/react/24/outline";
import { Dropdown, IconButton } from "@shellhub/design-system/primitives";
import RestrictedAction from "@/components/common/RestrictedAction";
import { useRecordingPermissions } from "@/hooks/useRecordingPermissions";
import { type Action } from "@/utils/permission";
import DeleteRecordingDialog from "./DeleteRecordingDialog";

function MenuItem({
  action,
  icon,
  label,
  danger,
  onSelect,
}: {
  action?: Action;
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
 * The overflow menu for a session's recording: download it, or delete every copy after
 * confirming, each gated by useRecordingPermissions. loaded says the caller already holds the
 * recording, so downloading it reads nothing and needs no permission. portal is off inside an
 * element that goes fullscreen, where a panel portalled to the body would not show.
 */
export default function RecordingActionsMenu({
  sessionUid,
  recorded,
  onDownload,
  placement = "bottom-end",
  portal = true,
  loaded = false,
}: {
  sessionUid: string;
  recorded: boolean;
  onDownload: () => void;
  placement?: "bottom-end" | "top-end";
  portal?: boolean;
  loaded?: boolean;
}) {
  const permissions = useRecordingPermissions(sessionUid, recorded);
  const [confirming, setConfirming] = useState(false);

  return (
    <>
      <Dropdown portal={portal} placement={placement}>
        <Dropdown.Trigger>
          <IconButton variant="ghost" aria-label="Recording actions">
            <EllipsisVerticalIcon className="w-4 h-4" />
          </IconButton>
        </Dropdown.Trigger>

        <Dropdown.Panel className="w-48 py-1">
          <MenuItem
            action={loaded ? undefined : permissions.read}
            icon={<ArrowDownTrayIcon className="w-4 h-4" />}
            label="Download recording"
            onSelect={onDownload}
          />
          <MenuItem
            action={permissions.remove}
            icon={<TrashIcon className="w-4 h-4" />}
            label="Delete recording"
            danger
            onSelect={() => setConfirming(true)}
          />
        </Dropdown.Panel>
      </Dropdown>

      <DeleteRecordingDialog
        open={confirming}
        onClose={() => setConfirming(false)}
        sessionUid={sessionUid}
        recorded={recorded}
      />
    </>
  );
}
