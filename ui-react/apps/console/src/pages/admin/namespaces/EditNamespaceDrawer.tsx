import { useState, type FormEvent } from "react";
import { useResetOnOpen } from "../../../hooks/useResetOnOpen";
import { useAdminEditNamespace } from "../../../hooks/useAdminNamespaceMutations";
import { isSdkError } from "../../../api/errors";
import Drawer from "../../../components/common/Drawer";
import { LABEL, INPUT } from "../../../utils/styles";
import type { Namespace } from "../../../client";

interface EditNamespaceDrawerProps {
  open: boolean;
  onClose: () => void;
  namespace: Namespace | null;
}

export default function EditNamespaceDrawer({
  open,
  onClose,
  namespace,
}: EditNamespaceDrawerProps) {
  const editNamespace = useAdminEditNamespace();

  const [name, setName] = useState("");
  const [maxDevices, setMaxDevices] = useState(-1);
  const [sessionRecord, setSessionRecord] = useState(false);
  const [error, setError] = useState("");

  useResetOnOpen(open, () => {
    setName(namespace?.name ?? "");
    setMaxDevices(namespace?.max_devices ?? -1);
    setSessionRecord(namespace?.settings?.session_record ?? false);
    setError("");
  });

  const canSubmit = name.trim().length > 0;

  const handleSubmit = async (e?: FormEvent) => {
    e?.preventDefault();
    if (!canSubmit || !namespace) return;
    setError("");
    try {
      await editNamespace.mutateAsync({
        path: { tenantID: namespace.tenant_id },
        // The SDK types body as full Namespace; we spread the original
        // to satisfy the type while only changing the editable fields.
        body: {
          ...namespace,
          name: name.trim(),
          max_devices: maxDevices,
          settings: {
            connection_announcement:
              namespace.settings?.connection_announcement ?? "",
            session_record: sessionRecord,
            disable_password: namespace.settings?.disable_password ?? false,
            disable_public_key: namespace.settings?.disable_public_key ?? false,
          },
        },
      });
      onClose();
    } catch (err) {
      if (isSdkError(err) && err.status === 409) {
        setError("A namespace with this name already exists.");
      } else {
        setError("Failed to update namespace. Please try again.");
      }
    }
  };

  return (
    <Drawer
      open={open}
      onClose={onClose}
      title="Edit Namespace"
      subtitle={
        namespace ? (
          <span className="font-mono">{namespace.name}</span>
        ) : undefined
      }
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
            disabled={!canSubmit || editNamespace.isPending}
            className="px-5 py-2.5 bg-primary hover:bg-primary-600 text-white rounded-lg text-sm font-semibold disabled:opacity-dim disabled:cursor-not-allowed transition-all flex items-center gap-2"
          >
            {editNamespace.isPending && (
              <span
                aria-hidden="true"
                className="w-4 h-4 border-2 border-white/30 border-t-white rounded-full animate-spin"
              />
            )}
            Save Changes
          </button>
        </>
      )}
    >
      <form onSubmit={(e) => void handleSubmit(e)} className="space-y-5">
        <div>
          <label className={LABEL} htmlFor="edit-ns-name">
            Name
          </label>
          <input
            id="edit-ns-name"
            type="text"
            value={name}
            onChange={(e) => setName(e.target.value)}
            autoFocus={open}
            className={INPUT}
          />
        </div>

        <div>
          <label className={LABEL} htmlFor="edit-ns-max-devices">
            Max Devices
          </label>
          <input
            id="edit-ns-max-devices"
            type="number"
            value={maxDevices}
            onChange={(e) => {
              const parsed = parseInt(e.target.value, 10);
              setMaxDevices(Number.isNaN(parsed) ? -1 : parsed);
            }}
            min={-1}
            className={INPUT}
          />
          <p className="text-2xs text-text-muted mt-1.5">
            Use -1 for unlimited devices
          </p>
        </div>

        <label className="flex items-center gap-2 cursor-pointer">
          <input
            type="checkbox"
            checked={sessionRecord}
            onChange={(e) => setSessionRecord(e.target.checked)}
            className="w-4 h-4 rounded border-border bg-card text-primary focus:ring-primary/20"
          />
          <span className="text-sm text-text-primary">Session Recording</span>
        </label>

        {error && (
          <p role="alert" className="text-2xs text-accent-red">
            {error}
          </p>
        )}
      </form>
    </Drawer>
  );
}
