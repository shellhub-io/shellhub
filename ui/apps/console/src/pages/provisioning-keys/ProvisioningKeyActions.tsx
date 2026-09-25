import { useState } from "react";
import { type ProvisioningKey } from "@/client";
import ProvisioningKeyActionsMenu from "./ProvisioningKeyActionsMenu";
import EditProvisioningKeyModal from "./EditProvisioningKeyModal";
import RevokeProvisioningKeyDialog from "./RevokeProvisioningKeyDialog";
import { useToggleProvisioningKey } from "./useToggleProvisioningKey";

/**
 * The actions on a provisioning key row, and the dialogs behind them.
 */
export default function ProvisioningKeyActions({
  provisioningKey,
}: {
  provisioningKey: ProvisioningKey;
}) {
  const [editOpen, setEditOpen] = useState(false);
  const [revokeOpen, setRevokeOpen] = useState(false);
  const { toggle, error: toggleError } = useToggleProvisioningKey();

  return (
    <>
      <div className="flex flex-col items-end">
        <ProvisioningKeyActionsMenu
          provisioningKey={provisioningKey}
          onEdit={() => setEditOpen(true)}
          onToggleDisabled={() => void toggle(provisioningKey)}
          onRevoke={() => setRevokeOpen(true)}
        />
        {toggleError && (
          <p className="mt-1 text-xs text-accent-red">{toggleError}</p>
        )}
      </div>

      <EditProvisioningKeyModal
        provisioningKey={editOpen ? provisioningKey : null}
        onClose={() => setEditOpen(false)}
      />

      <RevokeProvisioningKeyDialog
        provisioningKey={revokeOpen ? provisioningKey : null}
        onRevoked={() => setRevokeOpen(false)}
      />
    </>
  );
}
