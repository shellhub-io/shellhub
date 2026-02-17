import { useEffect, useState } from "react";
import { useApiKeysStore } from "../../stores/apiKeysStore";
import { type ApiKey } from "../../types/apiKey";
import ConfirmDialog from "../../components/common/ConfirmDialog";
import { RoleBadge } from "./constants";
import { isExpired } from "./helpers";
import { formatExpiry, formatDateShort } from "../../utils/date";
import { TH } from "../../utils/styles";
import GenerateKeyDrawer from "./GenerateKeyDrawer";
import EditKeyDrawer from "./EditKeyDrawer";
import {
  KeyIcon,
  ExclamationCircleIcon,
  PencilSquareIcon,
  TrashIcon,
} from "@heroicons/react/24/outline";
import Pagination from "../../components/common/Pagination";

/* ─── API Keys Tab ─── */

function ApiKeysTab() {
  const {
    apiKeys,
    totalCount,
    loading,
    page: currentPage,
    perPage,
    fetch: fetchApiKeys,
    setPage,
    remove: deleteApiKey,
  } = useApiKeysStore();
  const [generateOpen, setGenerateOpen] = useState(false);
  const [editTarget, setEditTarget] = useState<ApiKey | null>(null);
  const [deleteTarget, setDeleteTarget] = useState<ApiKey | null>(null);

  useEffect(() => {
    fetchApiKeys();
  }, [fetchApiKeys]);

  const totalPages = Math.ceil(totalCount / perPage);

  const handlePageChange = (page: number) => {
    setPage(page);
    fetchApiKeys(page);
  };

  return (
    <div className="animate-fade-in">
      <div className="flex items-center justify-between mb-5">
        <p className="text-sm text-text-muted">
          {totalCount} key{totalCount !== 1 ? "s" : ""}
        </p>
        <button
          onClick={() => setGenerateOpen(true)}
          className="flex items-center gap-2 px-4 py-2 bg-primary hover:bg-primary-600 text-white rounded-lg text-sm font-semibold transition-all"
        >
          <KeyIcon className="w-4 h-4" strokeWidth={2} />
          Generate Key
        </button>
      </div>

      {loading ? (
        <div className="flex items-center justify-center py-16">
          <span className="w-5 h-5 border-2 border-primary/30 border-t-primary rounded-full animate-spin" />
        </div>
      ) : apiKeys.length === 0 ? (
        <div className="text-center py-16">
          <KeyIcon className="w-10 h-10 text-text-muted/30 mx-auto mb-3" />
          <p className="text-sm text-text-muted">No API keys yet</p>
          <p className="text-2xs text-text-muted/60 mt-1">
            Generate a key to access the ShellHub API
          </p>
        </div>
      ) : (
        <div className="bg-card border border-border rounded-xl overflow-hidden">
          <table className="w-full">
            <thead>
              <tr className="border-b border-border">
                <th className={TH}>Name</th>
                <th className={TH}>Role</th>
                <th className={TH}>Created</th>
                <th className={TH}>Expires</th>
                <th className={`${TH} !text-right w-24`}>Actions</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-border">
              {apiKeys.map((key) => {
                const expired = isExpired(key.expires_in);
                return (
                  <tr
                    key={key.id}
                    className={`group transition-colors ${expired ? "bg-accent-red/[0.02]" : "hover:bg-hover-subtle"}`}
                  >
                    <td className="px-4 py-3.5">
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
                    </td>
                    <td className="px-4 py-3.5">
                      <RoleBadge role={key.role} />
                    </td>
                    <td className="px-4 py-3.5">
                      <span className="text-xs text-text-secondary">
                        {formatDateShort(key.created_at)}
                      </span>
                    </td>
                    <td className="px-4 py-3.5">
                      <span
                        className={`text-xs ${expired ? "text-accent-red" : "text-text-secondary"}`}
                      >
                        {formatExpiry(key.expires_in)}
                      </span>
                    </td>
                    <td className="px-4 py-3.5 text-right">
                      <div className="flex items-center justify-end gap-1 opacity-0 group-hover:opacity-100 transition-opacity">
                        <button
                          onClick={() => setEditTarget(key)}
                          className="p-1.5 rounded-md text-text-muted hover:text-text-primary hover:bg-hover-medium transition-colors"
                          title="Edit"
                        >
                          <PencilSquareIcon className="w-4 h-4" />
                        </button>
                        <button
                          onClick={() => setDeleteTarget(key)}
                          className="p-1.5 rounded-md text-text-muted hover:text-accent-red hover:bg-accent-red/5 transition-colors"
                          title="Delete"
                        >
                          <TrashIcon className="w-4 h-4" />
                        </button>
                      </div>
                    </td>
                  </tr>
                );
              })}
            </tbody>
          </table>

          <Pagination
            page={currentPage}
            totalPages={totalPages}
            onPageChange={handlePageChange}
          />
        </div>
      )}

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
          await deleteApiKey(deleteTarget!.name);
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
