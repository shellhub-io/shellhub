import { useState } from "react";
import { TicketIcon } from "@heroicons/react/24/outline";
import { Button, Spinner } from "@shellhub/design-system/primitives";
import { useInstallKeys } from "@/hooks/useInstallKeys";
import { usePaginatedListState } from "@/hooks/usePaginatedListState";
import { type InstallKey } from "@/client";
import PageHeader from "@/components/common/PageHeader";
import RestrictedAction from "@/components/common/RestrictedAction";
import InstallKeysTable from "./InstallKeysTable";
import CreateInstallKeyDrawer from "./CreateInstallKeyDrawer";
import EditInstallKeyDrawer from "./EditInstallKeyDrawer";
import RevokeInstallKeyDialog from "./RevokeInstallKeyDialog";
import { isSystemKey } from "./helpers";
import { useToggleInstallKey } from "./useToggleInstallKey";

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

  const totalPages = Math.ceil(totalCount / PER_PAGE);

  const [createOpen, setCreateOpen] = useState(false);
  const [editTarget, setEditTarget] = useState<InstallKey | null>(null);
  const [revokeTarget, setRevokeTarget] = useState<InstallKey | null>(null);
  const { toggle, error: toggleError } = useToggleInstallKey();

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
            onToggleDisabled={(k) => void toggle(k)}
            onRevoke={setRevokeTarget}
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
        onRevoked={() => setRevokeTarget(null)}
      />
    </div>
  );
}
