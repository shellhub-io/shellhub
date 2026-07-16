import { useState } from "react";
import { useUpdateInstallKey } from "@/hooks/useInstallKeyMutations";
import { type InstallKey } from "@/client";
import InstallKeyActionsMenu from "./InstallKeyActionsMenu";
import EditInstallKeyDrawer from "./EditInstallKeyDrawer";
import RevokeInstallKeyDialog from "./RevokeInstallKeyDialog";

/**
 * The install key overflow menu wired to its own drawers/dialogs, for anywhere a single key is shown
 * on its own (the activity page header). Bundles Edit, Disable/Enable, and the type-to-confirm Revoke
 * so a caller only drops in <InstallKeyActions installKey={key} />. The list page keeps its own
 * shared wiring since it drives one set of dialogs for every row.
 */
export default function InstallKeyActions({
  installKey,
}: {
  installKey: InstallKey;
}) {
  const updateKey = useUpdateInstallKey();
  const [editOpen, setEditOpen] = useState(false);
  const [revokeOpen, setRevokeOpen] = useState(false);
  const [revokeConfirmText, setRevokeConfirmText] = useState("");
  const [revokeError, setRevokeError] = useState<string | null>(null);
  const [toggleError, setToggleError] = useState<string | null>(null);

  // Pause/resume is reversible, so it fires the mutation directly (no confirm).
  const toggleDisabled = async () => {
    setToggleError(null);
    try {
      await updateKey.mutateAsync({
        path: { key: installKey.name },
        body: { disabled: !installKey.disabled },
      });
    } catch (err) {
      setToggleError(
        err instanceof Error
          ? err.message
          : `Failed to ${installKey.disabled ? "enable" : "disable"} Install Key.`,
      );
    }
  };

  const openRevoke = () => {
    setRevokeConfirmText("");
    setRevokeError(null);
    setRevokeOpen(true);
  };

  const closeRevoke = () => {
    setRevokeConfirmText("");
    setRevokeError(null);
    setRevokeOpen(false);
  };

  const confirmRevoke = async () => {
    setRevokeError(null);
    try {
      await updateKey.mutateAsync({
        path: { key: installKey.name },
        body: { revoked: true },
      });
      closeRevoke();
    } catch (err) {
      setRevokeError(
        err instanceof Error ? err.message : "Failed to revoke Install Key.",
      );
    }
  };

  return (
    <>
      <div className="flex flex-col items-end">
        <InstallKeyActionsMenu
          installKey={installKey}
          onEdit={() => setEditOpen(true)}
          onToggleDisabled={() => void toggleDisabled()}
          onRevoke={openRevoke}
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
        installKey={installKey}
        open={revokeOpen}
        confirmText={revokeConfirmText}
        onConfirmTextChange={setRevokeConfirmText}
        onClose={closeRevoke}
        onConfirm={confirmRevoke}
        error={revokeError}
      />
    </>
  );
}
