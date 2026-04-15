import { useState } from "react";
import { ExclamationCircleIcon } from "@heroicons/react/24/outline";
import { useResetOnOpen } from "@/hooks/useResetOnOpen";
import { useUpdateApiKey } from "@/hooks/useApiKeyMutations";
import { type ApiKey } from "@/client";
import Drawer from "@/components/common/Drawer";
import { LABEL, INPUT } from "@/utils/styles";
import { RoleSelector } from "./constants";
import { isAssignableRole, type AssignableRole } from "./helpers";

/* ─── Edit API Key Drawer ─── */

function EditKeyDrawer({
  open,
  onClose,
  apiKey,
}: {
  open: boolean;
  onClose: () => void;
  apiKey: ApiKey | null;
}) {
  const updateKey = useUpdateApiKey();
  const [name, setName] = useState("");
  const [role, setRole] = useState<AssignableRole>("administrator");
  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState<string | null>(null);

  useResetOnOpen(open, () => {
    setName(apiKey?.name ?? "");
    setRole(isAssignableRole(apiKey?.role) ? apiKey.role : "administrator");
    setSubmitting(false);
    setError(null);
  });

  const handleSubmit = async () => {
    if (!apiKey || !name.trim()) return;
    setSubmitting(true);
    setError(null);
    try {
      await updateKey.mutateAsync({ path: { key: apiKey.name }, body: { name: name.trim(), role } });
      onClose();
    } catch (err: unknown) {
      setError(err instanceof Error ? err.message : "Failed to update API key.");
    } finally {
      setSubmitting(false);
    }
  };

  return (
    <Drawer
      open={open}
      onClose={onClose}
      title="Edit API Key"
      footer={(
        <>
          <button
            type="button"
            onClick={onClose}
            className="px-4 py-2.5 text-sm font-medium text-text-secondary hover:text-text-primary rounded-lg hover:bg-hover-subtle transition-colors"
          >
            Cancel
          </button>
          <button
            onClick={() => void handleSubmit()}
            disabled={!name.trim() || submitting}
            className="px-5 py-2.5 bg-primary hover:bg-primary-600 text-white rounded-lg text-sm font-semibold disabled:opacity-dim disabled:cursor-not-allowed transition-all"
          >
            Save Changes
          </button>
        </>
      )}
    >
      <div className="space-y-5">
        <div>
          <label className={LABEL}>Name</label>
          <input
            type="text"
            value={name}
            onChange={(e) => setName(e.target.value)}
            autoFocus={open}
            className={INPUT}
          />
        </div>
        <div>
          <label className={LABEL}>Role</label>
          <RoleSelector value={role} onChange={setRole} />
        </div>

        {error && (
          <p className="text-xs font-mono text-accent-red flex items-center gap-1.5">
            <ExclamationCircleIcon className="w-3.5 h-3.5 shrink-0" strokeWidth={2} />
            {error}
          </p>
        )}
      </div>
    </Drawer>
  );
}

export default EditKeyDrawer;
