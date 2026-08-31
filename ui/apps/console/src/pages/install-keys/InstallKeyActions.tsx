import { useState } from "react";
import { type InstallKey } from "@/client";
import InstallKeyActionsMenu from "./InstallKeyActionsMenu";
import EditInstallKeyDrawer from "./EditInstallKeyDrawer";
import RevokeInstallKeyDialog from "./RevokeInstallKeyDialog";
import { useToggleInstallKey } from "./useToggleInstallKey";

/**
 * The actions on an install key row, and the dialogs behind them.
 */
export default function InstallKeyActions({
  installKey,
}: {
  installKey: InstallKey;
}) {
  const [editOpen, setEditOpen] = useState(false);
  const [revokeOpen, setRevokeOpen] = useState(false);
  const { toggle, error: toggleError } = useToggleInstallKey();

  return (
    <>
      <div className="flex flex-col items-end">
        <InstallKeyActionsMenu
          installKey={installKey}
          onEdit={() => setEditOpen(true)}
          onToggleDisabled={() => void toggle(installKey)}
          onRevoke={() => setRevokeOpen(true)}
        />
        {toggleError && (
          <p className="mt-1 text-xs text-accent-red">{toggleError}</p>
        )}
      </div>

      <EditInstallKeyDrawer
        installKey={editOpen ? installKey : null}
        onClose={() => setEditOpen(false)}
      />

      <RevokeInstallKeyDialog
        installKey={revokeOpen ? installKey : null}
        onRevoked={() => setRevokeOpen(false)}
      />
    </>
  );
}
