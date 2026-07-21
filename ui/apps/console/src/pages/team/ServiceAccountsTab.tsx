import { useState } from "react";
import { CpuChipIcon, TrashIcon } from "@heroicons/react/24/outline";
import { Button, IconButton } from "@shellhub/design-system/primitives";
import { useServiceAccounts } from "@/hooks/useServiceAccounts";
import { useDeleteServiceAccount } from "@/hooks/useServiceAccountMutations";
import { type ServiceAccount } from "@/client";
import ConfirmDialog from "@/components/common/ConfirmDialog";
import DataTable, { type Column } from "@/components/common/DataTable";
import RestrictedAction from "@/components/common/RestrictedAction";
import { formatDateShort } from "@/utils/date";
import ServiceAccountDrawer from "./ServiceAccountDrawer";

function ServiceAccountsTab() {
  const { serviceAccounts, isLoading } = useServiceAccounts();
  const deleteAccount = useDeleteServiceAccount();

  const [createOpen, setCreateOpen] = useState(false);
  const [deleteTarget, setDeleteTarget] = useState<ServiceAccount | null>(null);
  const [deleteError, setDeleteError] = useState<string | null>(null);

  const closeDelete = () => {
    setDeleteError(null);
    setDeleteTarget(null);
  };

  const confirmDelete = async () => {
    if (!deleteTarget) return;
    setDeleteError(null);
    try {
      await deleteAccount.mutateAsync({ path: { id: deleteTarget.id } });
      closeDelete();
    } catch (err) {
      setDeleteError(
        err instanceof Error
          ? err.message
          : "Failed to delete service account.",
      );
    }
  };

  const columns: Column<ServiceAccount>[] = [
    {
      key: "name",
      header: "Name",
      render: (account) => (
        <span className="text-sm font-medium text-text-primary">
          {account.name}
        </span>
      ),
    },
    {
      key: "identities",
      header: "SSH key",
      render: (account) => {
        const [first, ...rest] = account.identities;
        if (!first)
          return <span className="text-xs text-text-muted">No key</span>;
        return (
          <div className="flex items-center gap-2">
            <span className="text-xs font-mono text-text-secondary truncate max-w-[22rem]">
              {first.fingerprint}
            </span>
            {rest.length > 0 && (
              <span className="text-2xs text-text-muted">+{rest.length}</span>
            )}
          </div>
        );
      },
    },
    {
      key: "created_at",
      header: "Created",
      render: (account) => (
        <span className="text-xs text-text-secondary">
          {formatDateShort(account.created_at)}
        </span>
      ),
    },
    {
      key: "actions",
      header: "Actions",
      headerClassName: "text-right",
      render: (account) => (
        <div className="flex items-center justify-end gap-1">
          <RestrictedAction action="serviceAccount:delete">
            <IconButton
              variant="danger"
              title="Delete"
              aria-label="Delete service account"
              onClick={() => setDeleteTarget(account)}
            >
              <TrashIcon className="w-4 h-4" />
            </IconButton>
          </RestrictedAction>
        </div>
      ),
    },
  ];

  return (
    <div className="animate-fade-in">
      <div className="flex items-center justify-between mb-5">
        <p className="text-sm text-text-muted">
          {serviceAccounts.length} service account
          {serviceAccounts.length !== 1 ? "s" : ""}
        </p>
        <RestrictedAction action="serviceAccount:create">
          <Button
            onClick={() => setCreateOpen(true)}
            icon={<CpuChipIcon className="w-4 h-4" strokeWidth={2} />}
          >
            Add Service Account
          </Button>
        </RestrictedAction>
      </div>

      <DataTable
        columns={columns}
        data={serviceAccounts}
        rowKey={(account) => account.id}
        isLoading={isLoading}
        loadingMessage="Loading service accounts..."
        emptyState={
          <div className="text-center">
            <CpuChipIcon className="w-10 h-10 text-text-muted/30 mx-auto mb-3" />
            <p className="text-sm text-text-muted">No service accounts yet</p>
            <p className="text-2xs text-text-muted/60 mt-1">
              Create one to give an automated system its own SSH identity
            </p>
          </div>
        }
      />

      <ServiceAccountDrawer
        open={createOpen}
        onClose={() => setCreateOpen(false)}
      />
      <ConfirmDialog
        open={!!deleteTarget}
        onClose={closeDelete}
        onConfirm={confirmDelete}
        title="Delete Service Account"
        description={
          <>
            Are you sure you want to delete{" "}
            <span className="font-medium text-text-primary">
              {deleteTarget?.name}
            </span>
            ? This revokes its SSH identities and anything using them will stop
            working.
          </>
        }
        confirmLabel="Delete"
      >
        {deleteError && (
          <p className="text-xs text-accent-red">{deleteError}</p>
        )}
      </ConfirmDialog>
    </div>
  );
}

export default ServiceAccountsTab;
