import { useState } from "react";
import { KeyIcon, TrashIcon } from "@heroicons/react/24/outline";
import { Button, IconButton } from "@shellhub/design-system/primitives";
import { cn } from "@shellhub/design-system/cn";
import { useInstanceApiKeys } from "@/hooks/useInstanceApiKeys";
import { useDeleteInstanceApiKey } from "@/hooks/useInstanceApiKeyMutations";
import { usePaginatedListState } from "@/hooks/usePaginatedListState";
import { type InstanceApiKey } from "@/client";
import PageHeader from "@/components/common/PageHeader";
import ConfirmDialog from "@/components/common/ConfirmDialog";
import DataTable, { type Column } from "@/components/common/DataTable";
import { formatDateShort } from "@/utils/date";
import { pageCount } from "@/utils/pagination";
import GenerateInstanceKeyDrawer from "./GenerateInstanceKeyDrawer";

type InstanceApiKeyListParams = {
  page: number;
};

const INSTANCE_API_KEY_LIST_DEFAULTS: InstanceApiKeyListParams = { page: 1 };

function hasExpired(expiresAt: string) {
  return new Date(expiresAt).getTime() <= Date.now();
}

/**
 * The instance API keys page of the admin panel. These keys authenticate as an instance
 * administrator rather than as a member of a namespace, so they are managed here rather than
 * alongside a namespace's own keys.
 */
function InstanceApiKeys() {
  const { params, setPage } = usePaginatedListState<InstanceApiKeyListParams>({
    defaults: INSTANCE_API_KEY_LIST_DEFAULTS,
  });
  const page = params.page;
  const { apiKeys, totalCount, isLoading } = useInstanceApiKeys({ page });

  const deleteKey = useDeleteInstanceApiKey();
  const [generateOpen, setGenerateOpen] = useState(false);
  const [deleteTarget, setDeleteTarget] = useState<InstanceApiKey | null>(null);
  const [deleteError, setDeleteError] = useState<string | null>(null);

  const closeDelete = () => {
    setDeleteError(null);
    setDeleteTarget(null);
  };

  const confirmDelete = async () => {
    if (!deleteTarget) return;
    setDeleteError(null);
    try {
      await deleteKey.mutateAsync({ path: { name: deleteTarget.name } });
      if (apiKeys.length === 1 && page > 1) setPage(page - 1);
      closeDelete();
    } catch (err) {
      setDeleteError(
        err instanceof Error
          ? err.message
          : "Failed to revoke the instance API key.",
      );
    }
  };

  const columns: Column<InstanceApiKey>[] = [
    {
      key: "name",
      header: "Name",
      render: (key) => (
        <span className="text-sm font-medium text-text-primary">
          {key.name}
        </span>
      ),
    },
    {
      key: "created_at",
      header: "Created",
      render: (key) => (
        <span className="text-xs text-text-secondary">
          {formatDateShort(key.created_at)}
        </span>
      ),
    },
    {
      key: "expires_at",
      header: "Expires",
      render: (key) => (
        <span
          className={cn(
            "text-xs",
            hasExpired(key.expires_at)
              ? "text-accent-red"
              : "text-text-secondary",
          )}
        >
          {formatDateShort(key.expires_at)}
        </span>
      ),
    },
    {
      key: "actions",
      header: "Actions",
      headerClassName: "text-right",
      render: (key) => (
        <div className="flex items-center justify-end gap-1">
          <IconButton
            variant="danger"
            title="Revoke"
            aria-label="Revoke instance API key"
            onClick={() => setDeleteTarget(key)}
          >
            <TrashIcon className="w-4 h-4" />
          </IconButton>
        </div>
      ),
    },
  ];

  return (
    <div className="animate-fade-in">
      <PageHeader
        icon={<KeyIcon className="w-6 h-6" />}
        overline="Settings"
        title="Instance API Keys"
        description="Credentials that authenticate automation as an instance administrator"
      >
        <Button
          onClick={() => setGenerateOpen(true)}
          icon={<KeyIcon className="w-4 h-4" strokeWidth={2} />}
        >
          Generate Key
        </Button>
      </PageHeader>

      <div className="flex items-center justify-between mb-5">
        <p className="text-sm text-text-muted">
          {totalCount} key{totalCount !== 1 ? "s" : ""}
        </p>
      </div>

      <DataTable
        columns={columns}
        data={apiKeys}
        rowKey={(key) => key.name}
        isLoading={isLoading}
        loadingMessage="Loading instance API keys..."
        page={page}
        totalPages={pageCount(totalCount)}
        onPageChange={setPage}
        rowClassName={(key) =>
          hasExpired(key.expires_at)
            ? "bg-accent-red/[0.03] border-l-2 border-l-accent-red/50"
            : "border-l-2 border-l-transparent"
        }
        emptyState={
          <div className="text-center">
            <KeyIcon className="w-10 h-10 text-text-muted/30 mx-auto mb-3" />
            <p className="text-sm text-text-muted">No instance API keys yet</p>
            <p className="text-2xs text-text-muted/60 mt-1">
              Generate a key to automate the admin API
            </p>
          </div>
        }
      />

      <GenerateInstanceKeyDrawer
        open={generateOpen}
        onClose={() => setGenerateOpen(false)}
      />
      <ConfirmDialog
        open={!!deleteTarget}
        onClose={closeDelete}
        onConfirm={confirmDelete}
        title="Revoke Instance API Key"
        description={
          <>
            Are you sure you want to revoke{" "}
            <span className="font-medium text-text-primary">
              {deleteTarget?.name}
            </span>
            ? Any automation using this key will stop working immediately.
          </>
        }
        confirmLabel="Revoke"
      >
        {deleteError && (
          <p className="text-xs text-accent-red">{deleteError}</p>
        )}
      </ConfirmDialog>
    </div>
  );
}

export default InstanceApiKeys;
