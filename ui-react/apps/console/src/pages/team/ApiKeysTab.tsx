import { useState } from "react";
import { useApiKeys } from "@/hooks/useApiKeys";
import { useDeleteApiKey } from "@/hooks/useApiKeyMutations";
import { type ApiKey } from "@/client";
import ConfirmDialog from "@/components/common/ConfirmDialog";
import DataTable, { type Column } from "@/components/common/DataTable";
import { RoleBadge } from "./constants";
import { isExpired } from "./helpers";
import { formatExpiry, formatDateShort } from "@/utils/date";
import GenerateKeyDrawer from "./GenerateKeyDrawer";
import EditKeyDrawer from "./EditKeyDrawer";
import {
  KeyIcon,
  ExclamationCircleIcon,
  PencilSquareIcon,
  TrashIcon,
} from "@heroicons/react/24/outline";
import RestrictedAction from "@/components/common/RestrictedAction";

const PER_PAGE = 10;

function ApiKeysTab() {
  const [page, setPage] = useState(1);
  const { apiKeys, totalCount, isLoading } = useApiKeys({ page });
  const deleteKey = useDeleteApiKey();
  const [generateOpen, setGenerateOpen] = useState(false);
  const [editTarget, setEditTarget] = useState<ApiKey | null>(null);
  const [deleteTarget, setDeleteTarget] = useState<ApiKey | null>(null);

  const totalPages = Math.ceil(totalCount / PER_PAGE);

  const columns: Column<ApiKey>[] = [
    {
      key: "name",
      header: "Name",
      render: (key) => {
        const expired = isExpired(key.expires_in);
        return (
          <div className="flex items-center gap-2">
            <span className="text-sm font-medium text-text-primary">
              {key.name}
            </span>
            {expired && (
              <span className="inline-flex items-center gap-1 px-1.5 py-0.5 text-2xs font-mono font-semibold text-accent-red bg-accent-red/10 border border-accent-red/20 rounded">
                <ExclamationCircleIcon
                  className="w-2.5 h-2.5"
                  strokeWidth={2}
                />
                Expired
              </span>
            )}
          </div>
        );
      },
    },
    {
      key: "role",
      header: "Role",
      render: (key) => <RoleBadge role={key.role} />,
    },
    {
      key: "created",
      header: "Created",
      render: (key) => (
        <span className="text-xs text-text-secondary">
          {formatDateShort(key.created_at)}
        </span>
      ),
    },
    {
      key: "expires",
      header: "Expires",
      render: (key) => {
        const expired = isExpired(key.expires_in);
        return (
          <span
            className={`text-xs ${expired ? "text-accent-red" : "text-text-secondary"}`}
          >
            {formatExpiry(key.expires_in)}
          </span>
        );
      },
    },
    {
      key: "actions",
      header: "Actions",
      headerClassName: "text-right",
      render: (key) => (
        <div className="flex items-center justify-end gap-1">
          <RestrictedAction action="apiKey:edit">
            <button
              onClick={() => setEditTarget(key)}
              className="p-1.5 rounded-md text-text-muted hover:text-text-primary hover:bg-hover-medium transition-colors"
              title="Edit"
            >
              <PencilSquareIcon className="w-4 h-4" />
            </button>
          </RestrictedAction>
          <RestrictedAction action="apiKey:delete">
            <button
              onClick={() => setDeleteTarget(key)}
              className="p-1.5 rounded-md text-text-muted hover:text-accent-red hover:bg-accent-red/5 transition-colors"
              title="Delete"
            >
              <TrashIcon className="w-4 h-4" />
            </button>
          </RestrictedAction>
        </div>
      ),
    },
  ];

  return (
    <div className="animate-fade-in">
      <div className="flex items-center justify-between mb-5">
        <p className="text-sm text-text-muted">
          {totalCount} key
          {totalCount !== 1 ? "s" : ""}
        </p>
        <RestrictedAction action="apiKey:create">
          <button
            onClick={() => setGenerateOpen(true)}
            className="flex items-center gap-2 px-4 py-2 bg-primary hover:bg-primary-600 text-white rounded-lg text-sm font-semibold transition-all"
          >
            <KeyIcon className="w-4 h-4" strokeWidth={2} />
            Generate Key
          </button>
        </RestrictedAction>
      </div>

      <DataTable
        columns={columns}
        data={apiKeys}
        rowKey={(key) => key.name}
        isLoading={isLoading}
        loadingMessage="Loading API keys..."
        page={page}
        totalPages={totalPages}
        totalCount={totalCount}
        itemLabel="key"
        onPageChange={setPage}
        // border-l-2 on every row (transparent by default) keeps the row
        // height stable when the red border appears on expired keys.
        // No hover darkening here — rows are not clickable.
        rowClassName={(key) =>
          isExpired(key.expires_in)
            ? "bg-accent-red/[0.03] border-l-2 border-l-accent-red/50"
            : "border-l-2 border-l-transparent"
        }
        emptyState={
          <div className="text-center">
            <KeyIcon className="w-10 h-10 text-text-muted/30 mx-auto mb-3" />
            <p className="text-sm text-text-muted">No API keys yet</p>
            <p className="text-2xs text-text-muted/60 mt-1">
              Generate a key to access the ShellHub API
            </p>
          </div>
        }
      />

      <GenerateKeyDrawer
        open={generateOpen}
        onClose={() => setGenerateOpen(false)}
      />
      <EditKeyDrawer
        open={!!editTarget}
        onClose={() => setEditTarget(null)}
        apiKey={editTarget}
      />
      <ConfirmDialog
        open={!!deleteTarget}
        onClose={() => setDeleteTarget(null)}
        onConfirm={async () => {
          await deleteKey.mutateAsync({ path: { key: deleteTarget!.name } });
          if (apiKeys.length === 1 && page > 1) setPage(page - 1);
          setDeleteTarget(null);
        }}
        title="Delete API Key"
        description={
          <>
            Are you sure you want to delete{" "}
            <span className="font-medium text-text-primary">
              {deleteTarget?.name}
            </span>
            ? Any integrations using this key will stop working.
          </>
        }
        confirmLabel="Delete"
      />
    </div>
  );
}

export default ApiKeysTab;
