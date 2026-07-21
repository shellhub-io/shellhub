import { useState } from "react";
import { TicketIcon } from "@heroicons/react/24/outline";
import { Button, Spinner } from "@shellhub/design-system/primitives";
import { useInstallKeys } from "@/hooks/useInstallKeys";
import { useUpdateInstallKey } from "@/hooks/useInstallKeyMutations";
import { usePaginatedListState } from "@/hooks/usePaginatedListState";
import { type InstallKey } from "@/client";
import PageHeader from "@/components/common/PageHeader";
import RestrictedAction from "@/components/common/RestrictedAction";
import InstallKeysTable from "./InstallKeysTable";
import CreateInstallKeyDrawer from "./CreateInstallKeyDrawer";
import EditInstallKeyDrawer from "./EditInstallKeyDrawer";
import RevokeInstallKeyDialog from "./RevokeInstallKeyDialog";
import { isSystemKey } from "./helpers";

const PER_PAGE = 10;

type InstallKeyListParams = {
  page: number;
};

const INSTALL_KEY_LIST_DEFAULTS: InstallKeyListParams = { page: 1 };

export default function InstallKeys() {
  const { params, setPage } = usePaginatedListState<InstallKeyListParams>({
    prefix: "installKey",
    defaults: INSTALL_KEY_LIST_DEFAULTS,
  });
  const page = params.page;
  const { installKeys, totalCount, isLoading } = useInstallKeys({ page });

  // The store pins the namespace's auto-managed built-in keys (legacy, pairing) first and paginates
  // them together with the user's keys, so totalCount already counts them. Drive both the page count
  // and the item count off that one base, or the "N keys" label and the page controls disagree.
  const totalPages = Math.ceil(totalCount / PER_PAGE);

  const updateKey = useUpdateInstallKey();
  const [createOpen, setCreateOpen] = useState(false);
  const [editTarget, setEditTarget] = useState<InstallKey | null>(null);
  const [revokeTarget, setRevokeTarget] = useState<InstallKey | null>(null);
  const [revokeConfirmText, setRevokeConfirmText] = useState("");
  const [revokeError, setRevokeError] = useState<string | null>(null);
  const [toggleError, setToggleError] = useState<string | null>(null);

  // Pause/resume is reversible, so it fires the mutation directly (no confirm).
  const toggleDisabled = async (key: InstallKey) => {
    setToggleError(null);
    try {
      await updateKey.mutateAsync({
        path: { key: key.name },
        body: { disabled: !key.disabled },
      });
    } catch (err) {
      setToggleError(
        err instanceof Error
          ? err.message
          : `Failed to ${key.disabled ? "enable" : "disable"} Install Key.`,
      );
    }
  };

  // Revoke is irreversible (no un-revoke) and cuts off the key for new enrollments, so it's gated
  // behind typing the key's name — mirroring the vault-reset confirmation.
  const openRevoke = (key: InstallKey) => {
    setRevokeConfirmText("");
    setRevokeError(null);
    setRevokeTarget(key);
  };

  const closeRevoke = () => {
    setRevokeConfirmText("");
    setRevokeError(null);
    setRevokeTarget(null);
  };

  const confirmRevoke = async () => {
    if (!revokeTarget) return;
    setRevokeError(null);
    try {
      await updateKey.mutateAsync({
        path: { key: revokeTarget.name },
        body: { revoked: true },
      });
      closeRevoke();
    } catch (err) {
      setRevokeError(
        err instanceof Error ? err.message : "Failed to revoke Install Key.",
      );
    }
  };

  // The built-in keys (legacy + pairing) always fill the table, so a namespace with no user-created
  // keys is not an empty page — it's an empty "Custom keys" section. Detect that (any non-built-in key
  // on this page, or more than one page, means custom keys exist) so the table can show the onboarding
  // placeholder in that section instead. Replaces the old full-page hero, which never showed.
  const noCustomKeys =
    !installKeys.some((key) => !isSystemKey(key)) && totalPages <= 1;

  return (
    <div>
      <PageHeader
        icon={<TicketIcon className="w-6 h-6" />}
        overline="Provisioning"
        title="Install Keys"
        description="Install keys are reusable credentials that register devices with your namespace. Each key's mode decides how a device is admitted."
      >
        <RestrictedAction action="installKey:create">
          <Button
            onClick={() => setCreateOpen(true)}
            icon={<TicketIcon className="w-4 h-4" strokeWidth={2} />}
          >
            Create Install Key
          </Button>
        </RestrictedAction>
      </PageHeader>

      {isLoading ? (
        <div className="flex justify-center py-24">
          <Spinner />
        </div>
      ) : (
        <div className="animate-fade-in">
          {toggleError && (
            <p className="mb-3 text-xs text-accent-red">{toggleError}</p>
          )}

          <InstallKeysTable
            data={installKeys}
            page={page}
            totalPages={totalPages}
            totalCount={totalCount}
            noCustomKeys={noCustomKeys}
            onPageChange={setPage}
            onCreate={() => setCreateOpen(true)}
            onEdit={setEditTarget}
            onToggleDisabled={(k) => void toggleDisabled(k)}
            onRevoke={openRevoke}
          />
        </div>
      )}

      <CreateInstallKeyDrawer
        open={createOpen}
        onClose={() => setCreateOpen(false)}
      />
      <EditInstallKeyDrawer
        installKey={editTarget}
        onClose={() => setEditTarget(null)}
      />
      <RevokeInstallKeyDialog
        installKey={revokeTarget}
        open={!!revokeTarget}
        confirmText={revokeConfirmText}
        onConfirmTextChange={setRevokeConfirmText}
        onClose={closeRevoke}
        onConfirm={confirmRevoke}
        error={revokeError}
      />
    </div>
  );
}
