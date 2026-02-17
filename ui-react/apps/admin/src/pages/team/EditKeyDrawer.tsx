import { useState, useEffect } from "react";
import { useApiKeysStore } from "../../stores/apiKeysStore";
import { type ApiKey } from "../../types/apiKey";
import Drawer from "../../components/common/Drawer";
import { LABEL, INPUT } from "../../utils/styles";
import { RoleSelector } from "./constants";

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
  const update = useApiKeysStore((s) => s.update);
  const [name, setName] = useState(apiKey?.name ?? "");
  const [role, setRole] = useState(apiKey?.role ?? "administrator");
  const [submitting, setSubmitting] = useState(false);

  useEffect(() => {
    if (open && apiKey) {
      setName(apiKey.name);
      setRole(apiKey.role);
    }
  }, [open, apiKey]);

  const handleSubmit = async () => {
    if (!apiKey || !name.trim()) return;
    setSubmitting(true);
    try {
      await update(apiKey.name, name.trim(), role);
      onClose();
    } catch {
      /* */
    } finally {
      setSubmitting(false);
    }
  };

  return (
    <Drawer
      open={open}
      onClose={onClose}
      title="Edit API Key"
      footer={
        <>
          <button
            type="button"
            onClick={onClose}
            className="px-4 py-2.5 text-sm font-medium text-text-secondary hover:text-text-primary rounded-lg hover:bg-hover-subtle transition-colors"
          >
            Cancel
          </button>
          <button
            onClick={handleSubmit}
            disabled={!name.trim() || submitting}
            className="px-5 py-2.5 bg-primary hover:bg-primary-600 text-white rounded-lg text-sm font-semibold disabled:opacity-dim disabled:cursor-not-allowed transition-all"
          >
            Save Changes
          </button>
        </>
      }
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
      </div>
    </Drawer>
  );
}

export default EditKeyDrawer;
