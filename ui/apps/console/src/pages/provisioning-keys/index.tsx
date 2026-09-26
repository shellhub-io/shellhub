import { useState } from "react";
import { PlusIcon } from "@heroicons/react/24/outline";
import { Button, Spinner } from "@shellhub/design-system/primitives";
import { useProvisioningKeys } from "@/hooks/useProvisioningKeys";
import { usePaginatedListState } from "@/hooks/usePaginatedListState";
import { type ProvisioningKey } from "@/client";
import SettingsSection from "@/components/settings/SettingsSection";
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
  const { provisioningKeys, totalCount, isLoading } = useProvisioningKeys({
    page,
  });

  const totalPages = pageCount(totalCount);

  const [createOpen, setCreateOpen] = useState(false);
  const [editTarget, setEditTarget] = useState<ProvisioningKey | null>(null);
  const [revokeTarget, setRevokeTarget] = useState<ProvisioningKey | null>(
    null,
  );
  const { toggle, error: toggleError } = useToggleProvisioningKey();

  const noCustomKeys =
    !provisioningKeys.some((key) => !isSystemKey(key)) && totalPages <= 1;

  return (
    <div>
      <SettingsSection
        wide
        title="Provisioning keys"
        description="Keys devices register with. Each key's mode decides whether a device gets in."
        action={
          <RestrictedAction action="provisioningKey:create">
            <Button
              size="sm"
              onClick={() => setCreateOpen(true)}
              icon={<PlusIcon className="w-4 h-4" strokeWidth={2} />}
            >
              New key
            </Button>
          </RestrictedAction>
        }
      >
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
      </SettingsSection>
    </div>
  );
}
