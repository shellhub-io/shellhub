import { useState } from "react";
import { TicketIcon } from "@heroicons/react/24/outline";
import { Button, Spinner } from "@shellhub/design-system/primitives";
import { useProvisioningKeys } from "@/hooks/useProvisioningKeys";
import { usePaginatedListState } from "@/hooks/usePaginatedListState";
import { type ProvisioningKey } from "@/client";
import PageHeader from "@/components/common/PageHeader";
import RestrictedAction from "@/components/common/RestrictedAction";
import ProvisioningKeysTable from "./ProvisioningKeysTable";
import CreateProvisioningKeyModal from "./CreateProvisioningKeyModal";
import EditProvisioningKeyModal from "./EditProvisioningKeyModal";
import RevokeProvisioningKeyDialog from "./RevokeProvisioningKeyDialog";
import { isSystemKey } from "./helpers";
import { useToggleProvisioningKey } from "./useToggleProvisioningKey";
import { pageCount } from "@/utils/pagination";

type ProvisioningKeyListParams = {
  page: number;
};

const PROVISIONING_KEY_LIST_DEFAULTS: ProvisioningKeyListParams = { page: 1 };

/**
 * The provisioning keys list. Its URL state is prefixed, because the page carries a second paginated
 * list — the key's enrolment history — and the two must not share query parameters.
 */
export default function ProvisioningKeys() {
  const { params, setPage } = usePaginatedListState<ProvisioningKeyListParams>({
    prefix: "provisioningKey",
    defaults: PROVISIONING_KEY_LIST_DEFAULTS,
  });
  const page = params.page;
  const { provisioningKeys, totalCount, isLoading } = useProvisioningKeys({ page });

  const totalPages = pageCount(totalCount);

  const [createOpen, setCreateOpen] = useState(false);
  const [editTarget, setEditTarget] = useState<ProvisioningKey | null>(null);
  const [revokeTarget, setRevokeTarget] = useState<ProvisioningKey | null>(null);
  const { toggle, error: toggleError } = useToggleProvisioningKey();

  const noCustomKeys =
    !provisioningKeys.some((key) => !isSystemKey(key)) && totalPages <= 1;

  return (
    <div>
      <PageHeader
        icon={<TicketIcon className="w-6 h-6" />}
        overline="Settings"
        title="Provisioning Keys"
        description="Provisioning keys are reusable credentials that register devices with your namespace. Each key's mode decides how a device is admitted."
      >
        <RestrictedAction action="provisioningKey:create">
          <Button
            onClick={() => setCreateOpen(true)}
            icon={<TicketIcon className="w-4 h-4" strokeWidth={2} />}
          >
            Create Provisioning Key
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

          <ProvisioningKeysTable
            data={provisioningKeys}
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

      <CreateProvisioningKeyModal
        open={createOpen}
        onClose={() => setCreateOpen(false)}
      />
      <EditProvisioningKeyModal
        provisioningKey={editTarget}
        onClose={() => setEditTarget(null)}
      />
      <RevokeProvisioningKeyDialog
        provisioningKey={revokeTarget}
        onRevoked={() => setRevokeTarget(null)}
      />
    </div>
  );
}
